# Multi-stage Dockerfile for AINative Code
# Optimized for size and security

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies (if go.mod exists)
RUN if [ -f go.mod ]; then go mod download; fi

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_DATE
ARG CGO_ENABLED=0

# Build the application
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}" \
    -o ainative-code \
    ./cmd/ainative-code

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    bash \
    && addgroup -g 1000 ainative \
    && adduser -D -u 1000 -G ainative ainative

# Copy timezone data and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /build/ainative-code /usr/local/bin/ainative-code

# Set proper permissions
RUN chmod +x /usr/local/bin/ainative-code

# Switch to non-root user
USER ainative

# Set working directory
WORKDIR /home/ainative

# Create config directory
RUN mkdir -p /home/ainative/.config/ainative-code

# Environment variables
ENV PATH="/usr/local/bin:${PATH}"
ENV AINATIVE_CODE_CONFIG_DIR="/home/ainative/.config/ainative-code"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ainative-code version || exit 1

# Labels
LABEL org.opencontainers.image.title="AINative Code"
LABEL org.opencontainers.image.description="AI-Native Development, Natively"
LABEL org.opencontainers.image.vendor="AINative Studio"
LABEL org.opencontainers.image.source="https://github.com/AINative-studio/ainative-code"
LABEL org.opencontainers.image.documentation="https://docs.ainative.studio/code"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"

# Entry point
ENTRYPOINT ["ainative-code"]

# Default command
CMD ["--help"]
