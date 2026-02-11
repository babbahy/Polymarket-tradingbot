# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -o polymarket-arbitrage ./cmd/trader

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 trader && \
    adduser -D -u 1000 -G trader trader

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/polymarket-arbitrage .

# Copy configuration files
COPY --chown=trader:trader configs/ ./configs/

# Create data directory
RUN mkdir -p /app/data /app/logs && \
    chown -R trader:trader /app/data /app/logs

# Switch to non-root user
USER trader

# Expose metrics and health check ports
EXPOSE 8080 9090

# Default to paper trading mode
ENTRYPOINT ["/app/polymarket-arbitrage"]
CMD ["--mode=paper", "--config=configs/config.yaml"]
