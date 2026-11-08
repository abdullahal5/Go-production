# -------------------------------------------------------------
# Stage 1: Build stage
# -------------------------------------------------------------
FROM golang:alpine AS builder

# Install required build tools and CA certificates
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build statically linked binary with optimizations (strip debug symbols)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/migrate ./cmd/migrate

# -------------------------------------------------------------
# Stage 2: Final lightweight runtime stage
# -------------------------------------------------------------
FROM alpine:latest

# Install CA certificates for secure TLS/SSL database connections (e.g. Neon, AWS RDS)
# and tzdata for standard timezone handling.
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binaries and migration files from the builder stage
COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations

# Run as non-root user for security best practices
USER appuser

# Expose HTTP port
EXPOSE 8080

# Run API binary
ENTRYPOINT ["/app/api"]

