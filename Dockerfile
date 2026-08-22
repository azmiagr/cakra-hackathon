# syntax=docker/dockerfile:1
FROM golang:1.25.5-alpine3.22 AS builder

WORKDIR /src
RUN apk add --no-cache build-base libwebp-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/cakra-api ./cmd/app

FROM alpine:3.22

RUN apk add --no-cache ca-certificates libwebp tzdata \
    && addgroup -S cakra \
    && adduser -S -G cakra -u 10001 cakra

WORKDIR /app
COPY --from=builder /out/cakra-api /app/cakra-api

USER cakra
EXPOSE 8082
ENTRYPOINT ["./cakra-hackathon"]
