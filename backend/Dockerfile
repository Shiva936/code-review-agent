# syntax=docker/dockerfile:1

##
# Multi-stage build for a small final image.
# This Dockerfile is located in `backend/`, so the recommended build command is:
#   docker build -t self-improving-code-review-bot:local .
#
# Backend uses mattn/go-sqlite3 (CGO), so the builder stage must include build tooling.
##

FROM golang:1.22-alpine AS builder

RUN apk add --no-cache build-base ca-certificates

WORKDIR /app

# Copy backend sources (includes vendored dependencies).
COPY . ./ 

ENV CGO_ENABLED=1

# Build the backend binary.
RUN go build -mod=vendor -o /out/self-improving-bot .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && mkdir -p /data

WORKDIR /app

COPY --from=builder /out/self-improving-bot /app/self-improving-bot

EXPOSE 8080

ENV PORT=8080

VOLUME ["/data"]

CMD ["/app/self-improving-bot"]

