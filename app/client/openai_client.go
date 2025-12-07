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

// Summarize textual data that is received from the audio file
func SummarizeText(text string) (string, error) {
	resp, err := openAI.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant that summarizes text concisely."),
			openai.UserMessage(fmt.Sprintf("Please summarize the following text:\n\n%s", text)),
		},
		MaxTokens: openai.Int(500), // Optional: limit response length
	})
	if err != nil {
		return "", fmt.Errorf("chat completion error: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response generated")
}

// TranscribeHandler is a Gin handler that accepts a filename parameter.
// It fetches the file from the "videos" folder and transcribes it.
func Summarizer(audioPath string) utils.Response {

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

	text, err = SummarizeText(text)
	if err != nil {
		return utils.Response{
			Code:    500,
			Message: fmt.Sprintf("[ERROR]: summarization failed: %v", err),
			Data:    nil,
			Status:  false,
		}
	}

	// Success
	return utils.Response{
		Code:    200,
		Message: "Summary Generation Successful",
		Data:    map[string]any{"Summary": text},
		Status:  true,
	}
}
