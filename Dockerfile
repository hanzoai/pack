# The one Dockerfile in the fabric that pack does not build: pack cannot pack
# itself. The native /v1/runner uses dockerfile.v0 to bootstrap this image, then
# every other repo builds with frontend=gateway.v0 + source=ghcr.io/hanzoai/pack.
FROM golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pack .

# Pinned by digest, and TAGGED: a reference carrying neither tag nor digest is
# refused at parse time by BuildKit — `object required`, from
# reference.ErrObjectRequired — before it opens a socket, so it reads like an
# unreachable registry and is not one. :latest at the time of pinning.
FROM gcr.io/distroless/static-debian12:latest@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2
COPY --from=build /pack /pack
ENTRYPOINT ["/pack"]
