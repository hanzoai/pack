# The one Dockerfile in the fabric that pack does not build: pack cannot pack
# itself. The native /v1/runner uses dockerfile.v0 to bootstrap this image, then
# every other repo builds with frontend=gateway.v0 + source=ghcr.io/hanzoai/pack.
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /pack .

FROM gcr.io/distroless/static-debian12
COPY --from=build /pack /pack
ENTRYPOINT ["/pack"]
