FROM rust:1-slim-bookworm AS builder

# Compile mise from source with size optimizations
RUN apt-get update && apt-get install -y \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/jdx/mise.git /tmp/mise && \
    cd /tmp/mise && \
    RUSTFLAGS="-C opt-level=z -C link-arg=-s -C codegen-units=1" \
    cargo build --profile serious \
    --config debuginfo=0

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
COPY --from=builder /tmp/mise/target/serious/mise /usr/local/bin/
RUN chmod +x /usr/local/bin/mise
