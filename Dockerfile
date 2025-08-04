# Multi-stage build für optimierte amd64 Binary
FROM golang:1.21-alpine AS builder

# Build dependencies für CGO installieren
RUN apk add --no-cache gcc musl-dev

# Arbeitsverzeichnis setzen
WORKDIR /app

# Go Modules kopieren für besseres Caching
COPY go.mod go.sum ./
RUN go mod download

# Source Code kopieren
COPY . .

# Binary für amd64 bauen
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o host-monitor .

# Final Stage - minimal Image
FROM alpine:latest

# CA Zertifikate für HTTPS-Verbindungen
RUN apk --no-cache add ca-certificates

# Arbeitsverzeichnis erstellen
WORKDIR /app

# Binary aus Builder Stage kopieren
COPY --from=builder /app/host-monitor .

# Executable-Rechte setzen
RUN chmod +x host-monitor

# Health Check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep -f host-monitor || exit 1

# Entrypoint
ENTRYPOINT ["./host-monitor"]