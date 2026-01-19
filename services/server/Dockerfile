FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/bufbuild/buf/cmd/buf@v1.63.0
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
RUN go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.18.0

COPY buf.yaml buf.lock buf.gen.yaml ./
COPY proto ./proto
COPY services/server ./services/server

ENV CGO_ENABLED=0

RUN buf generate

RUN go build -o /out/william-server ./services/server/cmd/william-server


FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/william-server /app/william-server
COPY --from=builder /app/db/migrations /app/db/migrations

EXPOSE 8080 8081
