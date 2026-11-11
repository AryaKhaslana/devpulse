# Stage 1: Build static binary
FROM golang:alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/devpulse ./cmd/bot/main.go

# Stage 2: Minimal runtime image (< 15MB)
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates and tzdata for TLS and Asia/Jakarta timezone
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/devpulse /app/devpulse

ENV TIMEZONE=Asia/Jakarta
ENV TZ=Asia/Jakarta

ENTRYPOINT ["/app/devpulse"]
