# Vertex starts the services it manages inside this container, so the image
# carries what they need to run: JDKs, Maven, bash and git. The JDKs are glibc
# builds, so the image is Debian rather than Alpine, and the binary is built on
# Debian to link against the same C library.
FROM golang:1.23.10-bookworm AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (frontend should be pre-built by CI)
COPY . .

# Verify frontend dist exists (required for go:embed)
RUN ls -la web/dist/ || (echo "ERROR: web/dist directory missing! Frontend must be built before Docker build." && exit 1)

# cgo is required because mattn/go-sqlite3 is a cgo package
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex

# Each JDK is taken whole from the official Eclipse Temurin image. A service
# picks one by setting JAVA_HOME to /opt/java/<version>. To offer another
# version, add a stage here and a COPY below - or extend the published image:
#   FROM zechtz/vertex
#   COPY --from=eclipse-temurin:11-jdk /opt/java/openjdk /opt/java/11
FROM eclipse-temurin:17-jdk AS jdk17
FROM eclipse-temurin:21-jdk AS jdk21

# Maven is only needed to generate a wrapper for a project that has none;
# services themselves start through ./mvnw or ./gradlew.
FROM maven:3.9-eclipse-temurin-21 AS maven

FROM debian:bookworm-slim

# bash runs service start commands, git the branch tools, lsof the port
# cleanup before a start, and curl the health check.
RUN apt-get update && \
    apt-get install -y --no-install-recommends bash ca-certificates curl git lsof sqlite3 && \
    rm -rf /var/lib/apt/lists/*

# Projects are mounted from the host and owned by a host user, which git
# otherwise refuses to read as root ("detected dubious ownership").
RUN git config --system --add safe.directory '*'

COPY --from=jdk17 /opt/java/openjdk /opt/java/17
COPY --from=jdk21 /opt/java/openjdk /opt/java/21
COPY --from=maven /usr/share/maven /opt/maven
RUN ln -s /opt/maven/bin/mvn /usr/local/bin/mvn

# The JDK a service gets when it sets no JAVA_HOME of its own
ENV JAVA_HOME=/opt/java/21
ENV PATH=/opt/java/21/bin:$PATH

WORKDIR /app

# Copy the binary
COPY --from=builder /app/vertex .

# Vertex keeps its database and logs in VERTEX_DATA_DIR; mount a volume here
ENV VERTEX_DATA_DIR=/app/data
RUN mkdir -p /app/data

# Expose port (default 54321, configurable via PORT env var)
EXPOSE 54321
ENV PORT=54321

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -fsS -o /dev/null http://localhost:${PORT}/ || exit 1

# Run the application with port from environment
CMD ["sh", "-c", "./vertex --port $PORT"]
