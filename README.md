
![alt text](https://cdn.hashnode.com/res/hashnode/image/upload/v1698324966590/ecc115a4-12ef-4074-a7be-45a009792324.png)

# Summarizer

A small Go service to download YouTube videos (uses github.com/kkdai/youtube) and serve a simple summary of the video over HTTP API with Gin.

## Features
- Download YouTube videos (prefers combined audio+video; falls back to DASH download + ffmpeg muxing)
- HTTP API built with Gin
- Saves output files to a local `video/` `audio/` directory (project working directory)
- Simple JSON request/response model and structured service responses

## Prerequisites
- Go 1.20+ installed and on PATH
- ffmpeg installed and available on PATH (required for muxing audio/video)
- (Optional) VS Code configured so the integrated terminal sees `go` and `ffmpeg` on PATH

## Install & run
From the repo root:
```powershell
# install deps and tidy modules
go get ./...
go mod tidy

# run server
go run .
# server listens on :8080 by default
```

## API
- GET /              — health check (returns "ok")
- POST /summarize    — summarize a video (need to provide URL in body)

## Files
```
├───app
│   ├───client
│   │       openai_client.go
│   ├───controllers
│   │       health_check.go
│   │       summarize.go
│   ├───models
│   │       request.go
│   │       response.go
│   ├───routes
│   │       routes.go
│   ├───services
│   │       audio_extractor.go
│   │       video_downloader.go
│   ├───utils
│   │       clear_directory.go
│   └───validators
│           youtube_url_validator.go
├───audio
├───compose
│       compose.yml
└───video
```
## Usage

- Run docker compose command which will start the service on port 8080. You need to add Open AI API key in the environment variables section of the compose file.

## Notes & tips
- If downloads lack audio, ensure ffmpeg is installed and the service is selecting/merging audio+video formats (the code attempts to mux when needed).
- If VS Code terminal doesn't recognize `go` or `ffmpeg`, restart VS Code or launch it from a shell that has the correct PATH.
- Add or adjust ignored files in `.gitignore` (the repo already ignores video outputs and temp files).

