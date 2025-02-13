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

# Install mise
# TODO: I don't think we need these envs here. Or we can remove from the mise_step
ENV MISE_INSTALL_PATH=/usr/local/bin/mise \
    MISE_DATA_DIR=/mise \
    MISE_CONFIG_DIR=/mise \
    MISE_CACHE_DIR=/mise/cache \
    MISE_VERBOSE=1 \
    PATH=/mise/shims:$PATH

# TODO: Install mise from source to get a smaller binary

RUN curl -fsSL https://mise.run | sh
