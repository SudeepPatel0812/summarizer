# summarizer

A small Go service to download YouTube videos (uses github.com/kkdai/youtube) and serve a simple HTTP API with Gin.

## Features
- Download YouTube videos (prefers combined audio+video; falls back to DASH download + ffmpeg muxing)
- HTTP API built with Gin
- Saves output files to a local `video/` directory (project working directory)
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
- GET /             — health check (returns "ok")
- POST /download    — download a video

## File locations
- Main server: `main.go`
- Controllers: `app/controllers/`
- Services (downloader): `app/services/video_downloader.go`
- Utilities (request/response types, validators): `app/utils/`
- Downloads: `./video/` (created at runtime in working directory)

## Notes & tips
- If downloads lack audio, ensure ffmpeg is installed and the service is selecting/merging audio+video formats (the code attempts to mux when needed).
- If VS Code terminal doesn't recognize `go` or `ffmpeg`, restart VS Code or launch it from a shell that has the correct PATH.
- Add or adjust ignored files in `.gitignore` (the repo already ignores video outputs and temp files).

## Contributing
Feel free to open issues or submit PRs to improve format selection, parallel downloads, or add a web UI.

## License
Add your preferred license file (e.g. MIT) to the repo.