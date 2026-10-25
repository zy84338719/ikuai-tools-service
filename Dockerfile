# Build the ikuai-tools-service image.
#
# All dependencies resolve from the module proxy at published versions
# (ikuai-api v1.3.0, ikuai_exporter v1.0.1, oekit) — a plain single-repo build:
#
#   docker build -t ikuai-tools-service .

FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime image ---
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /out/server ./server
# Ship the example config so the container runs out of the box; mount a real
# config.yaml at /app/configs/config.yaml (or set CONFIG_PATH) in production.
COPY configs/config.yaml.example ./configs/config.yaml.example

# API port (default 9997) and the embedded Prometheus exporter port (9100).
EXPOSE 9997 9100

# The service reads CONFIG_PATH to find its config; default to the shipped example.
ENV CONFIG_PATH=/app/configs/config.yaml.example

ENTRYPOINT ["./server"]
