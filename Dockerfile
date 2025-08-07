# Multi-stage Dockerfile for vertex
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM golang:1.23-bullseye AS backend-builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/web/dist ./web/dist

# Build the binary
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex

FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy the binary
COPY --from=backend-builder /app/vertex .

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
