FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -o /out/william-server ./cmd/william-server
RUN go build -o /out/admin-server ./cmd/admin-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates iproute2 wireguard-tools

WORKDIR /app

COPY --from=builder /out/william-server /app/william-server
COPY --from=builder /out/admin-server /app/admin-server
COPY --from=builder /app/db/migrations /app/db/migrations

EXPOSE 8080 8081
