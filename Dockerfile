# Multi-stage build for Go application
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG BUILD_VERSION=latest
ARG GIT_HASH=unknown
ARG GIT_BRANCH=unknown
ARG GIT_STAT=clean
ARG GIT_TAG=latest
ARG BUILD_DATE=0

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X github.com/unbxd/go-starter/cmd/ldflags.Module=github.com/unbxd/go-starter \
              -X github.com/unbxd/go-starter/cmd/ldflags.GitHash=${GIT_HASH} \
              -X github.com/unbxd/go-starter/cmd/ldflags.GitBranch=${GIT_BRANCH} \
              -X github.com/unbxd/go-starter/cmd/ldflags.GitStat=${GIT_STAT} \
              -X github.com/unbxd/go-starter/cmd/ldflags.GitTag=${GIT_TAG} \
              -X github.com/unbxd/go-starter/cmd/ldflags.BuildDate=${BUILD_DATE} \
              -X github.com/unbxd/go-starter/cmd/ldflags.Version=${BUILD_VERSION}" \
    -o gostarter.bin ./cmd/app

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /service

# Copy binary from builder
COPY --from=builder /build/gostarter.bin /service/gostarter.bin

# Copy scripts if needed
COPY scripts/* /service/scripts/ 2>/dev/null || true

# Make binary executable
RUN chmod +x /service/gostarter.bin

# Expose default port
EXPOSE 6060

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:6060/pong || exit 1

# Run the application
CMD ["/service/gostarter.bin", "start"]
