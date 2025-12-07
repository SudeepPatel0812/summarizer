package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"summarizer/app/utils"
)

// ExtractAudioToWav extracts mono 16k WAV audio from an input video file using ffmpeg.
// Returns path to the extracted WAV file in the audio folder.
func ExtractAudioToWav(videoPath string) (string, error) {
	// Create audio directory if it doesn't exist
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create audio directory: %w", err)
	}

	// Get the base filename without extension and create output path
	baseFilename := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	out := filepath.Join(audioDir, fmt.Sprintf("%s.wav", baseFilename))

	// ffmpeg -y -i input -vn -ac 1 -ar 16000 -f wav out
	cmd := exec.Command("ffmpeg", "-y",
		"-i", videoPath,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-f", "wav",
		out,
	)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w (output: %s)", err, string(outBytes))
	}
	return out, nil
}

// isVideoExtension returns true for known video file extensions.
func isVideoExtension(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp4", ".mkv", ".mov", ".webm", ".avi", ".flv", ".wmv", ".mpeg":
		return true
	default:
		return false
	}
}

func AudioExtractor(filename string) utils.Response {

	// Build path to file - only join if filename is not an absolute path
	var videoPath string
	if filepath.IsAbs(filename) {
		// Use absolute path directly
		videoPath = filename
	} else {
		// Relative path - join with videos folder
		videoPath = filepath.Join("videos", filename)
	}

	srcFilename := filepath.Base(videoPath)

	// Check if file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return utils.Response{
			Code:    404,
			Message: fmt.Sprintf("[ERROR]: file not found: %s", videoPath),
			Data:    nil,
			Status:  false,
		}
	}

	if isVideoExtension(srcFilename) {
		// Extract audio
		audioPath, err := ExtractAudioToWav(videoPath)
		if err != nil {
			return utils.Response{
				Code:    500,
				Message: fmt.Sprintf("[ERROR]: extracting audio: %v", err),
				Data:    nil,
				Status:  false,
			}
		}

		return utils.Response{
			Code:    200,
			Message: "Audio extraction successful",
			Data:    map[string]interface{}{"audio_path": audioPath},
			Status:  true,
		}
	}

	return utils.Response{
		Code:    200,
		Message: "File is not a video, no extraction needed",
		Data:    nil,
		Status:  true,
	}

}
