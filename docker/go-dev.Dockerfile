# syntax=docker/dockerfile:1
# One dev image for every Go service (compose.projects.yml). The repo is
# mounted at /src; `dev-go <name>` hot-reloads services/<name> with air.
FROM golang:1.27-alpine

ARG TARGETARCH
ARG AIR_VERSION=1.67.4

ADD --chmod=755 https://github.com/air-verse/air/releases/download/v${AIR_VERSION}/air_${AIR_VERSION}_linux_${TARGETARCH} /usr/local/bin/air
COPY docker/air.toml /etc/air.toml
COPY --chmod=755 docker/dev-go.sh /usr/local/bin/dev-go

WORKDIR /src
ENV PORT=8080
EXPOSE 8080
