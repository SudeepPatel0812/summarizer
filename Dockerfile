FROM golang:1.25.5 AS builder
WORKDIR /src

# copy only go.mod first (do NOT require go.sum)
COPY go.mod ./

# Download modules (will succeed even if go.sum absent)
RUN go mod download

# copy the rest of the project
COPY . .

# build (adjust path to your main package)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/summarizer ./


# final
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/summarizer /usr/local/bin/summarizer

RUN useradd --create-home --shell /bin/bash appuser

RUN mkdir -p /app/video && chown appuser:appuser /app/video

WORKDIR /app

USER appuser

EXPOSE 8080

ENV GIN_MODE=release TZ=UTC

CMD ["/usr/local/bin/summarizer"]
