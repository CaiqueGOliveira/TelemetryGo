# syntax=docker/dockerfile:1

# ---------- build stage ----------
FROM golang:1.27-alpine AS build

WORKDIR /app

ENV CGO_ENABLED=0 \
    GOFLAGS=-mod=mod \
    GOCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /out/telemetrygo ./src

# ---------- runtime stage ----------
FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

COPY --from=build /out/telemetrygo /usr/local/bin/telemetrygo

USER app

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/telemetrygo"]