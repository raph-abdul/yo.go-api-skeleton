# Copyright 2025 Raph Abdul
# Licensed under the Apache License, Version 2.0.
# Visit http://www.apache.org/licenses/LICENSE-2.0 for details.

# / Dockerfile

# --- Builder Stage ---
# Use an appropriate Go version. Alpine images are smaller.
FROM golang:1.24-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Install build dependencies + curl (for downloading migrate)
RUN apk add --no-cache git build-base postgresql-dev curl

# --- Add migrate CLI installation ---
ARG MIGRATE_VERSION=v4.17.1
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate \
    && migrate -version # Verify installation
# --- End migrate CLI installation ---

# Copy Go module files first to leverage Docker cache
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download
# Verify dependencies (optional, but good practice)
RUN go mod verify

# Copy the entire project source code
COPY . .

# Build the application binary (keeping your existing setup)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG APP_VERSION="0.0.1-dev"
ARG BUILD_COMMIT="unknown"
ARG BUILD_DATE="unknown"
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build \
    -ldflags="-s -w \
    -X youGo/cmd/api.AppVersion=${APP_VERSION} \
    -X youGo/cmd/api.BuildCommit=${BUILD_COMMIT} \
    -X youGo/cmd/api.BuildDate=${BUILD_DATE}" \
    -o /app/bin/you-go-server \
    ./cmd/api

# --- Final Stage ---
# Use a minimal base image. Alpine is small. Distroless is even smaller/more secure.
FROM alpine:latest
# FROM gcr.io/distroless/static-debian11 # Alternative: requires fully static binary from builder

# Install necessary runtime dependencies
# ca-certificates: for making HTTPS calls from the app
# tzdata: for timezone support (if your app uses timezones)
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Create a non-root user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup  # Keep this line

# Copy necessary files from the builder stage and host context
# Copy the compiled application binary
COPY --from=builder /app/bin/you-go-server /app/bin/you-go-server

# Copy the application configuration files
COPY --from=builder /app/configs /app/configs/

# Copy the installed migrate CLI binary
COPY --from=builder /usr/local/bin/migrate /app/bin/migrate

# Copy the SQL migration files from the host
COPY migrations/ /app/migrations/

# Copy the migration script from the host
COPY scripts/migrate.sh /app/scripts/migrate.sh

# Ensure binaries and script are executable
RUN chmod +x /app/bin/you-go-server \
    && chmod +x /app/bin/migrate \
    && chmod +x /app/scripts/migrate.sh

# Change ownership of the application directory/files to the non-root user
# Ensures the application doesn't run as root
RUN chown -R appuser:appgroup /app # This already covers /app/bin, /app/scripts, /app/migrations

# Switch to the non-root user
USER appuser

# Expose the port the application listens on (should match config)
EXPOSE 8080

# --- START: Added PATH environment variable ---
# Add /app/bin (where migrate now lives) to the PATH
ENV PATH="/app/bin:${PATH}"
# --- END: Added PATH environment variable ---

# Define the default command to run when the container starts
CMD ["/app/bin/you-go-server"]