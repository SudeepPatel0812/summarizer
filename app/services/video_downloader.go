package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"summarizer/app/utils"
	"summarizer/app/validators"

	"github.com/kkdai/youtube/v2"
)

// VideoDownloader downloads a YouTube video. It prefers combined (audio+video) formats.
// If the video only has separate audio/video (DASH), it downloads both and uses ffmpeg to mux.
func VideoDownloader(url string) utils.Response {
	// Validate URL (assuming validators.YoutubeURLValidator returns true for invalid — keep your logic)
	if validators.YoutubeURLValidator(url) {
		return utils.Response{
			Code:    400,
			Message: "URL is invalid",
			Data:    nil,
			Status:  false,
		}
	}

	fmt.Printf("[INFO]: Downloading Video: %s\n", url)

	wd, err := os.Getwd()
	if err != nil {
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: get working dir: %v", err),
			Data:    nil,
			Status:  false,
		}
	}
	outDir := filepath.Join(wd, "video")
	fmt.Printf("[INFO]: creating/using folder: %s\n", outDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: creating folder: %v", err),
			Data:    nil,
			Status:  false,
		}
	}

	client := youtube.Client{}
	video, err := client.GetVideo(url)
	if err != nil {
		return utils.Response{
			Code:    400,
			Message: fmt.Sprintf("[ERROR]: getting video info: %v", err),
			Data:    nil,
			Status:  false,
		}
	}

	if len(video.Formats) == 0 {
		return utils.Response{
			Code:    400,
			Message: "[ERROR]: no available formats",
			Data:    nil,
			Status:  false,
		}
	}

	// Inspect available formats (optional debug)
	fmt.Printf("[DEBUG]: Available formats for video %s:\n", video.ID)
	for i := range video.Formats {
		f := &video.Formats[i]
		fmt.Printf("  itag=%d mime=%s quality=%s audioCh=%d size=%d\n",
			f.ItagNo, f.MimeType, f.QualityLabel, f.AudioChannels, f.ContentLength)
	}

	// Try to find a combined format first (AudioChannels > 0)
	var combined *youtube.Format
	for i := range video.Formats {
		f := &video.Formats[i]
		if f.AudioChannels > 0 && strings.HasPrefix(f.MimeType, "video/") {
			combined = f
			break
		}
	}

	if combined != nil {
		// Download combined format directly
		outPath := filepath.Join(outDir, fmt.Sprintf("%s-%d.%s", video.ID, combined.ItagNo, extensionFromMime(combined.MimeType)))
		fmt.Printf("[INFO]: Found combined format itag=%d, saving to %s\n", combined.ItagNo, outPath)
		if err := downloadFormat(&client, video, combined, outPath); err != nil {
			return utils.Response{
				Code:    500,
				Message: fmt.Sprintf("[ERROR]: downloading combined stream: %v", err),
				Data:    nil,
				Status:  false,
			}
		}
		return utils.Response{
			Code:    200,
			Message: fmt.Sprintf("[INFO]: saved to %s", outPath),
			Data:    outPath,
			Status:  true,
		}
	}

	// No combined format found -> find best video-only and best audio-only
	videoFmt, audioFmt := pickBestVideoAndAudio(video)
	if videoFmt == nil || audioFmt == nil {
		// As a last resort, download first format raw
		fmt.Println("[WARN]: Could not find separate audio+video pair; falling back to first format")
		fallback := &video.Formats[0]
		outPath := filepath.Join(outDir, fmt.Sprintf("%s-%d.%s", video.ID, fallback.ItagNo, extensionFromMime(fallback.MimeType)))
		if err := downloadFormat(&client, video, fallback, outPath); err != nil {
			return utils.Response{
				Code:    500,
				Message: fmt.Sprintf("[ERROR]: fallback download failed: %v", err),
				Data:    nil,
				Status:  false,
			}
		}
		return utils.Response{
			Code:    200,
			Message: fmt.Sprintf("[INFO]: saved fallback to %s", outPath),
			Data:    outPath,
			Status:  true,
		}
	}

	fmt.Printf("[INFO]: Selected video itag=%d (%s), audio itag=%d (%s)\n",
		videoFmt.ItagNo, videoFmt.QualityLabel, audioFmt.ItagNo, audioFmt.QualityLabel)

	// Download both to temp files
	videoTmp := filepath.Join(outDir, fmt.Sprintf("%s-video-%d.%s", video.ID, videoFmt.ItagNo, extensionFromMime(videoFmt.MimeType)))
	audioTmp := filepath.Join(outDir, fmt.Sprintf("%s-audio-%d.%s", video.ID, audioFmt.ItagNo, extensionFromMime(audioFmt.MimeType)))
	finalOut := filepath.Join(outDir, fmt.Sprintf("%s-final.mp4", video.ID))

	fmt.Printf("[INFO]: Downloading video stream to %s\n", videoTmp)
	if err := downloadFormat(&client, video, videoFmt, videoTmp); err != nil {
		cleanupFiles(videoTmp, audioTmp)
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: download video-only stream: %v", err),
			Data:    nil,
			Status:  false,
		}
	}

	fmt.Printf("[INFO]: Downloading audio stream to %s\n", audioTmp)
	if err := downloadFormat(&client, video, audioFmt, audioTmp); err != nil {
		cleanupFiles(videoTmp, audioTmp)
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: download audio-only stream: %v", err),
			Data:    nil,
			Status:  false,
		}
	}

	// Mux with ffmpeg
	fmt.Printf("[INFO]: Muxing to final output %s using ffmpeg\n", finalOut)
	if err := muxWithFFmpeg(videoTmp, audioTmp, finalOut); err != nil {
		cleanupFiles(videoTmp, audioTmp, finalOut)
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: muxing streams: %v", err),
			Data:    nil,
			Status:  false,
		}
	}

	// Cleanup temp pieces
	cleanupFiles(videoTmp, audioTmp)

	return utils.Response{
		Code:    200,
		Message: fmt.Sprintf("[INFO]: saved to %s", finalOut),
		Data:    finalOut,
		Status:  true,
	}
}

// downloadFormat downloads a given youtube.Format to outPath.
func downloadFormat(client *youtube.Client, video *youtube.Video, format *youtube.Format, outPath string) error {
	stream, _, err := client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("GetStream error: %w", err)
	}
	defer stream.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create file error: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, stream)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}
	return nil
}

// pickBestVideoAndAudio finds a video-only format and an audio-only format.
// Strategy: pick highest resolution video (QualityLabel) and highest bitrate audio (approx by MimeType/AudioChannels).
func pickBestVideoAndAudio(video *youtube.Video) (videoFmt *youtube.Format, audioFmt *youtube.Format) {
	// Choose best video-only (video/* with AudioChannels==0), prefer higher QualityLabel (not perfect but pragmatic)
	for i := range video.Formats {
		f := &video.Formats[i]
		if strings.HasPrefix(f.MimeType, "video/") && f.AudioChannels == 0 {
			// pick first as candidate, but prefer one with QualityLabel non-empty and larger
			if videoFmt == nil {
				videoFmt = f
			} else {
				// prefer non-empty quality label over empty, otherwise keep existing
				if f.QualityLabel != "" && (videoFmt.QualityLabel == "" || qualityPrefers(f.QualityLabel, videoFmt.QualityLabel)) {
					videoFmt = f
				}
			}
		}
	}

	// Choose best audio-only (mime starts with audio/ or video/ with AudioChannels>0 but small video size)
	for i := range video.Formats {
		f := &video.Formats[i]
		if strings.HasPrefix(f.MimeType, "audio/") {
			if audioFmt == nil {
				audioFmt = f
			} else {
				// no exact bitrate available; prefer ones with AudioChannels > audioFmt
				if f.AudioChannels > audioFmt.AudioChannels {
					audioFmt = f
				}
			}
		}
		// Sometimes audio-only comes as video/ with audio (e.g., small mux); fallback:
		if strings.HasPrefix(f.MimeType, "video/") && f.AudioChannels > 0 {
			// treat as candidate if audioFmt nil
			if audioFmt == nil {
				audioFmt = f
			}
		}
	}
	return
}

// qualityPrefers returns true if qa looks higher than qb. Very simple heuristic based on numeric part.
func qualityPrefers(qa, qb string) bool {
	// qa and qb might be like "1080p", "720p", "medium", ""
	// parse numbers
	getNum := func(s string) int {
		var num int
		for i := 0; i < len(s); i++ {
			ch := s[i]
			if ch >= '0' && ch <= '9' {
				num = num*10 + int(ch-'0')
			}
		}
		return num
	}
	return getNum(qa) > getNum(qb)
}

// extensionFromMime returns a file extension guess from MIME-like string from youtube.Format.MimeType
func extensionFromMime(mime string) string {
	// sample mime strings: "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\""
	if strings.HasPrefix(mime, "video/mp4") {
		return "mp4"
	}
	if strings.HasPrefix(mime, "video/webm") {
		return "webm"
	}
	if strings.HasPrefix(mime, "audio/mp4") || strings.Contains(mime, "audio/mp4") {
		return "m4a"
	}
	if strings.HasPrefix(mime, "audio/webm") {
		return "webm"
	}
	// fallback
	if strings.Contains(mime, "mp4") {
		return "mp4"
	}
	return "bin"
}

// muxWithFFmpeg runs ffmpeg to merge video+audio into finalPath
func muxWithFFmpeg(videoPath, audioPath, finalPath string) error {
	// Command:
	// ffmpeg -y -i <videoPath> -i <audioPath> -c:v copy -c:a aac -b:a 192k -movflags +faststart <finalPath>
	cmd := exec.Command("ffmpeg", "-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "+faststart",
		finalPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w; output:%s", err, string(out))
	}
	return nil
}

// cleanupFiles removes files if they exist. Non-fatal.
func cleanupFiles(paths ...string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			fmt.Printf("[WARN]: cleanup failed for %s: %v\n", p, err)
		}
	}
}
