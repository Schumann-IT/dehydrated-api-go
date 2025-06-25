# Stage 1: Download and extract the snapshot binary
FROM alpine:3.19 AS downloader

WORKDIR /build

COPY dist/ /build/

RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
      x86_64) ARCH="x86_64" ;; \
      aarch64) ARCH="arm64" ;; \
      armv7l) ARCH="armv7" ;; \
      i386 | i686) ARCH="386" ;; \
      *) echo "Unsupported arch: $ARCH" && exit 1 ;; \
    esac && \
    OS=$(uname -s) && \
    tar -zxf dehydrated-api-go_${OS}_${ARCH}.tar.gz

RUN ./dehydrated-api-go -version

# Stage 2: Create the final minimal image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -s /bin/sh -u 1000 -G appuser appuser

# Create necessary directories
RUN mkdir -p /app /data/dehydrated /app/config && \
    chown -R appuser:appuser /app /data

# Copy the binary from the downloader stage
COPY --from=downloader /build/dehydrated-api-go /app/dehydrated-api-go

# Set ownership
RUN chown appuser:appuser /app/dehydrated-api-go

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Expose the default port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Default command
ENTRYPOINT ["/app/dehydrated-api-go"]

# Default arguments (can be overridden)
CMD ["--config", "/app/config/config.yaml"]