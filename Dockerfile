FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o /build/dehydrated-api-go ./cmd/api

FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
      bash \
      curl \
      openssl \
      ca-certificates \
      tzdata \
      dcron \
      yq

# Install python build tools
RUN apk add --no-cache \
      python3 \
      py3-pip \
      gcc \
      musl-dev \
      python3-dev \
      libffi-dev \
      openssl-dev \
      cargo

# Create and activate a virtual environment for Azure CLI
RUN python3 -m venv /opt/venv \
  && . /opt/venv/bin/activate \
  && pip install --upgrade pip \
  && pip install --no-cache-dir azure-cli \
  && deactivate

ENV PATH="/opt/venv/bin:$PATH"

RUN mkdir -p /app/scripts \
    && mkdir -p /app/config

# copy binaries, scripts and config
COPY --from=builder /build/dehydrated-api-go /app/
COPY --from=builder /build/scripts/ /app/scripts/

RUN chmod +x /app/scripts/*

# Set environment variables
ENV PORT=3000
ENV BASE_DIR=/data/dehydrated
ENV ENABLE_WATCHER=false
ENV ENABLE_OPENSSL_PLUGIN=false
# Format: {"<plugin_name>":{"enabled":<true|false>,"path":"</path/to/plugin|empty>"}}
ENV EXTERNAL_PLUGINS="{\"openssl\":{\"enabled\":false}}"
# Format: {"level":"<level>","encoding":"<console|json>","outputPath": "</path/to/logfile>"}
ENV EXTERNAL_PLUGINS="{\"level\":\"error\",\"encoding\":\"console\",\"outputPath\": \"\"}"

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /app/scripts/healthcheck.sh

# Set the entrypoint to the startup script
ENTRYPOINT ["/app/scripts/entrypoint.sh"]
