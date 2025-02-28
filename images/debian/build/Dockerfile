FROM buildpack-deps:bookworm-scm

# Install any system deps that might be needed to build a language from source
# Python: https://github.com/pyenv/pyenv/wiki#suggested-build-environment
RUN apt-get update && apt-get install -y \
    build-essential \
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    liblzma-dev

# Mise config
ENV MISE_INSTALL_PATH=/usr/local/bin/mise \
    MISE_DATA_DIR=/mise \
    MISE_CONFIG_DIR=/mise \
    MISE_CACHE_DIR=/mise/cache \
    DEBIAN_FRONTEND=noninteractive \
    PATH=/mise/shims:$PATH

# Copy mise binary from builder
ARG MISE_IMAGE=ghcr.io/railwayapp/railpack-mise:latest
COPY --from=${MISE_IMAGE} /usr/local/bin/mise /usr/local/bin/
RUN chmod +x /usr/local/bin/mise
