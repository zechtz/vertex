# Built on Alpine so the binary links against musl, matching the runtime stage
# below. A cgo binary built on Debian links glibc and cannot exec on Alpine -
# the loader it names, /lib64/ld-linux-*.so.2, does not exist there, and the
# container fails at startup with "not found".
FROM golang:1.23.10-alpine AS builder

# cgo toolchain: required because mattn/go-sqlite3 is a cgo package. It compiles
# its own SQLite, so no sqlite development headers are needed here.
RUN apk add --no-cache gcc musl-dev

# go-sqlite3 v1.14.17 predates musl dropping the LFS64 aliases, so its bundled
# SQLite fails to compile with "pread64 undeclared" / "unknown type off64_t".
# This restores those declarations. Removable once go-sqlite3 reaches v1.14.22+,
# which supports musl directly.
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

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
# Use --no-scripts to avoid trigger issues in QEMU ARM64 builds
RUN apk --no-cache --no-scripts add ca-certificates sqlite && \
    update-ca-certificates 2>/dev/null || true

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
