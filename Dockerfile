FROM golang:1.22-alpine AS builder

# Add build metadata
LABEL maintainer="dehydrated-api-go"
LABEL description="Dehydrated API Go - Multi-stage build stage"

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /build/dehydrated-api-go ./cmd/api

FROM alpine:3.19 AS python-builder

# Install Python build dependencies
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

FROM alpine:3.19

# Add runtime metadata
LABEL maintainer="dehydrated-api-go"
LABEL description="Dehydrated API Go - Runtime container"
LABEL version="1.0.0"

WORKDIR /app

# Install runtime dependencies only
RUN apk add --no-cache \
      bash \
      curl \
      openssl \
      ca-certificates \
      tzdata \
      dcron \
      yq \
      python3

# Copy Python virtual environment from builder stage
COPY --from=python-builder /opt/venv /opt/venv

# Set Python environment
ENV PATH="/opt/venv/bin:$PATH"

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create necessary directories and set ownership
RUN mkdir -p /app/scripts \
    && mkdir -p /app/config \
    && mkdir -p /data/dehydrated \
    && mkdir -p /tmp \
    && chown -R appuser:appgroup /app \
    && chown -R appuser:appgroup /data \
    && chown -R appuser:appgroup /tmp

# Copy binaries, scripts and config
COPY --from=builder /build/dehydrated-api-go /app/
COPY --from=builder /build/scripts/ /app/scripts/

# Set proper permissions
RUN chmod +x /app/scripts/* \
    && chown -R appuser:appgroup /app/scripts

# Switch to non-root user
USER appuser

# Set environment variables
ENV PORT=3000
ENV BASE_DIR=/data/dehydrated
ENV ENABLE_WATCHER=false
ENV ENABLE_OPENSSL_PLUGIN=false
# Format: {"<plugin_name>":{"enabled":<true|false>,"path":"</path/to/plugin|empty>"}}
ENV EXTERNAL_PLUGINS="{\"openssl\":{\"enabled\":false}}"
# Format: {"level":"<level>","encoding":"<console|json>","outputPath": "</path/to/logfile>"}
ENV LOGGING="{\"level\":\"error\",\"encoding\":\"console\",\"outputPath\": \"\"}"

# Expose port
EXPOSE 3000

# Define volume for persistent data
VOLUME ["/data/dehydrated"]

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /app/scripts/healthcheck.sh

# Set the entrypoint to the startup script
ENTRYPOINT ["/app/scripts/entrypoint.sh"]
