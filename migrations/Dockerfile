FROM golang:1.26-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct

ENV GOPROXY=${GOPROXY} \
    CGO_ENABLED=0

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install \
    -tags="no_sqlite3 no_postgres no_clickhouse no_mssql no_libsql no_vertica no_ydb" \
    github.com/pressly/goose/v3/cmd/goose@v3.28.0

FROM alpine:3.24

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/goose /usr/local/bin/goose

WORKDIR /migrations

ENTRYPOINT ["goose"]