# Optimized single-stage build for vertex
FROM golang:1.23-bullseye AS builder

# Install build dependencies in one layer
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (frontend should be pre-built by CI)
COPY . .

# Verify frontend dist exists (required for go:embed)
RUN ls -la web/dist/ || (echo "ERROR: web/dist directory missing! Frontend must be built before Docker build." && exit 1)

# Build the binary with optimizations
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex

FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy the binary
COPY --from=builder /app/vertex .

# Create directory for database
RUN mkdir -p /app/data

# Expose port (default 54321, configurable via PORT env var)
EXPOSE 54321

# Set environment variables
ENV DB_PATH=/app/data/vertex.db
ENV PORT=54321

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/ || exit 1

# Run the application with port from environment
CMD ["sh", "-c", "./vertex --port $PORT"]
