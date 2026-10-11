FROM golang:1.26-alpine AS builder

ENV CGO_ENANLED=0

RUN go install -tags=no_sqlite3 \
    github.com/pressly/goose/v3/cmd/goose@v3.28.0

FROM alpine:3.24

COPY --from=builder /go/bin/goose /usr/local/bin/goose

WORKDIR /migrations

ENTRYPOINT ["goose"]