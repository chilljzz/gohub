#syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

ARG GOPROXY=https://proxy.golang.org,direct

ENV GOPROXY=${GOPROXY}

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY . .

RUN mkdir -p /out \
    && CGO_ENABLED=0 \
    COOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w"\
    -o /out/gohub \
    ./cmd/server

FROM alpine:3.24

RUN apk add --no-cache \
    ca-certificates \
    tzdata

RUN addgroup -S gohub \
    && adduser -S -G gohub gohub

COPY --from=builder \
    /out/gohub \
    /app/gohub

USER gohub

EXPOSE 8081

HEALTHCHECK \
    --interval=10s \
    --timeout=3s \
    --start-period=10s \
    --retries=5 \
    CMD wget -qO- \
        http://127.0.0.1:8081/ping \
        >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/gohub"]