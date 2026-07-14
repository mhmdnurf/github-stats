FROM golang:1.26.5-alpine3.24 AS build

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN --mount=type=cache,target=/root/.cache/go-build \
    go test ./...

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/github-stats \
    ./cmd/server

FROM alpine:3.24.1 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=build \
    /out/github-stats \
    /usr/local/bin/github-stats

ENV HTTP_ADDRESS=:9000

EXPOSE 9000

USER app:app

HEALTHCHECK \
    --interval=30s \
    --timeout=3s \
    --start-period=5s \
    --retries=3 \
    CMD wget -q -O /dev/null \
        "http://127.0.0.1${HTTP_ADDRESS}/healthz" \
        || exit 1

ENTRYPOINT ["/usr/local/bin/github-stats"]