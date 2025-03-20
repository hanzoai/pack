FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive \
    LD_LIBRARY_PATH=/usr/local/lib:/usr/lib:/lib

RUN apt-get update && apt-get install -y \
    ca-certificates \
    pkg-config \
    openssl \
    && rm -rf /var/lib/apt/lists/*
