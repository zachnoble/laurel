# syntax=docker/dockerfile:1
# Dev image for any SvelteKit app: --build-arg APP=<name>. The repo is
# mounted at /repo; node_modules stay the image's (Linux binaries).
FROM node:26-alpine

WORKDIR /repo
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
RUN npm install -g corepack && corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm fetch --store-dir=/pnpm/store

COPY apps apps
COPY libraries/ts libraries/ts
COPY rpc/generated/typescript rpc/generated/typescript

ARG APP
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    test -n "$APP" && \
    pnpm install --offline --frozen-lockfile --filter "${APP}..." --store-dir=/pnpm/store

ENV APP=${APP}
WORKDIR /repo/apps/${APP}
CMD ["pnpm", "dev"]
