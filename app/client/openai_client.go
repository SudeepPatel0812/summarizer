package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"summarizer/app/utils"
)

var openAI *openai.Client

func init() {
    _ = godotenv.Load()

    apiKey, ok := os.LookupEnv("OPEN_AI_API_KEY")
    if !ok || strings.TrimSpace(apiKey) == "" {
        log.Fatal("OPEN_AI_API_KEY not set")
    }
	// assign to package-level variable (no :=)
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	openAI = &client
}

// TranscribeAudioFile uses the OpenAI SDK (Whisper) to transcribe an audio file.
// Returns the transcript text.
func TranscribeAudioFile(ctx context.Context, audioPath string) (string, error) {
	f, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("open audio file: %w", err)
	}
	defer f.Close()

	// Create transcription request via SDK
	resp, err := openAI.Audio.Transcriptions.New(
		ctx,
		openai.AudioTranscriptionNewParams{
			Model: openai.AudioModel("whisper-1"),
			File:  f, // Pass *os.File directly, no wrapper needed
			// Optional fields:
			// ResponseFormat: openai.AudioResponseFormatJSON,
			// Temperature: openai.Float(0.0),
			// Language:    openai.String("en"),
		},
	)
	if err != nil {
		return "", fmt.Errorf("whisper transcription error: %w", err)
	}

	return resp.Text, nil
}

// TranscribeHandler is a Gin handler that accepts a filename parameter.
// It fetches the file from the "videos" folder and transcribes it.
func Transcribe(audioPath string) utils.Response {

	ctx := context.Background()

	// Transcribe
	text, err := TranscribeAudioFile(ctx, audioPath)
	if err != nil {
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: transcription failed: %v", err),
			Data:    nil,
			Status:  false,
		}
	}
	// cleanup extracted audio after transcription
	defer func() { _ = os.Remove(audioPath) }()

	// Success
	return utils.Response{
		Code:    200,
		Message: "Transcription successful",
		Data:    map[string]any{"transcript": text},
		Status:  true,
	}
}
