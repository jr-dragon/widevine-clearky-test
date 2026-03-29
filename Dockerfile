FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server .

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Install Shaka Packager
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then PKG_ARCH="linux-x64"; \
    elif [ "$ARCH" = "arm64" ]; then PKG_ARCH="linux-arm64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    curl -fsSL "https://github.com/shaka-project/shaka-packager/releases/latest/download/packager-${PKG_ARCH}" \
      -o /usr/local/bin/shaka-packager && \
    chmod +x /usr/local/bin/shaka-packager

COPY --from=builder /server /server
COPY web /web

ENV VIDEOS_DIR=/data/videos
ENV KEYS_PATH=/data/keys.json
ENV WEB_DIR=/web
ENV ADDR=:8080
ENV SHAKA_PACKAGER_BIN=shaka-packager

EXPOSE 8080
VOLUME ["/data"]

CMD ["/server"]
