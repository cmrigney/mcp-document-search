# Build stage
FROM golang:1.25-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled (required for SQLite)
RUN CGO_ENABLED=1 go build -o doc-search ./cmd/doc-search

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/doc-search .

# Create directory for database
RUN mkdir -p /data

# Run as non-root user
RUN useradd -m -u 1000 mcp
RUN chown -R mcp:mcp /app /data
USER mcp

# Set default database path
ENV DB_PATH=/data/doc_search.db

ENTRYPOINT ["/app/doc-search"]
