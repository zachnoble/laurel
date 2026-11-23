# syntax=docker/dockerfile:1
# Production image for any Go service: docker build --build-arg SERVICE=<name> -f docker/go-service.Dockerfile .
#
# Reproducible: the same source gives the same image digest (no VCS stamping,
# -trimpath, pinned base images, and CI builds with SOURCE_DATE_EPOCH=0 and
# rewrite-timestamp), so deploys skip services whose image didn't change.
FROM golang:1.27-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY libraries/go libraries/go
COPY rpc/generated/go rpc/generated/go
COPY services services

ARG SERVICE
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    test -n "$SERVICE" && \
    CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/server "./services/$SERVICE/cmd/http"

# Static, CA certificates included, no shell or package manager.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/server ./server

EXPOSE 8080
ENTRYPOINT ["./server"]
