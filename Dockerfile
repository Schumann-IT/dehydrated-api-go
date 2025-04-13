FROM alpine:3.19 AS downloader

ARG GOOS="linux"
ARG GOARCH="arm64"
ARG ARCH="arm64"
ARG VERSION="v0.0.1"

WORKDIR /build

RUN apk add --no-cache curl

RUN mkdir -p /downloads

# download dehydrated
RUN curl -o /downloads/dehydrated https://raw.githubusercontent.com/dehydrated-io/dehydrated/refs/heads/master/dehydrated

# download dehydrated-api
RUN curl -L -o /build/dehydrated-api-go_${GOOS}_${ARCH}.tar.gz https://github.com/Schumann-IT/dehydrated-api-go/releases/download/${VERSION}/dehydrated-api-go_${GOOS}_${ARCH}.tar.gz
RUN tar -zxf /build/dehydrated-api-go_${GOOS}_${ARCH}.tar.gz
RUN cp /build/dehydrated-api-go /downloads/dehydrated-api-go

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

# copy binaries
COPY --from=downloader /downloads/dehydrated /app/scripts/
COPY --from=downloader /downloads/dehydrated-api-go /app/

# install default configs
COPY examples/config/config.yaml /app/config/
COPY examples/config/dehydrated /app/config/

# install scripts
COPY scripts/update-api-config.sh /app/scripts/
COPY scripts/update-dehydrated-config.sh /app/scripts/
COPY scripts/configure-cron.sh /app/scripts/
COPY scripts/start-crond.sh /app/scripts/
COPY scripts/renew-certs.sh /app/scripts/
COPY scripts/start-api.sh /app/scripts/
COPY scripts/healthcheck.sh /app/scripts/
COPY scripts/dehydrated-hook-azure.sh /app/scripts/

RUN chmod +x /app/scripts/*

# Set environment variables
ENV PORT=3000
ENV ENABLE_WATCHER=false
ENV ENABLE_OPENSSL_PLUGIN=true
# Format: 0 */12 * * *
ENV CRON_SCHEDULE=""
# Format: {"plugin_name":{"enabled":true,"path":"/path/to/plugin"}}
ENV EXTERNAL_PLUGINS="{}"
# dehydrated config settings @see https://github.com/dehydrated-io/dehydrated/blob/master/docs/examples/config
# each setting needs to be prefixed with DEHYDRATED_
# some settings are manged internally and cannot be changed: BASEDIR, CERTDIR, DOMAINS_TXT
# ENV DEHYDRATED_CA="letsencrypt-test"

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /app/scripts/healthcheck.sh

# Expose the API port
EXPOSE ${PORT}

# Set the entrypoint to the startup script
ENTRYPOINT ["/app/scripts/start-api.sh"]
