# Build the ikuai-tools-service image.
#
# The service depends on ikuai-api and ikuai_exporter via local `replace`
# directives (../ikuai-api, ../ikuai_exporter). Build this Dockerfile with the
# monorepo root as the context so those siblings resolve:
#
#   docker build -t ikuai-tools-service -f ikuai-tools-service/Dockerfile .
#
# (run from the directory containing ikuai-api/, ikuai_exporter/, ikuai-tools-service/)

# Match the Go version pinned in ikuai-tools-service/go.mod (go 1.24.0).
FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

# Copy the three repos side-by-side so the local replace directives resolve.
COPY ikuai-api/        ./ikuai-api/
COPY ikuai_exporter/   ./ikuai_exporter/
COPY ikuai-tools-service/ ./ikuai-tools-service/

WORKDIR /src/ikuai-tools-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

# --- runtime image ---
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /out/server ./server
# Ship the example config so the container runs out of the box; mount a real
# config.yaml at /app/configs/config.yaml (or set CONFIG_PATH) in production.
COPY ikuai-tools-service/configs/config.yaml.example ./configs/config.yaml.example

# API port (default 9997) and the embedded Prometheus exporter port (9100).
EXPOSE 9997 9100

# The service reads CONFIG_PATH to find its config; default to the shipped example.
ENV CONFIG_PATH=/app/configs/config.yaml.example

ENTRYPOINT ["./server"]
