# Stage 1: Build the Go binary
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Define build arguments
ARG PORT
ARG NATS_URL
ARG GIN_MODE
ARG DATABASE_URL
ARG NATS_MAX_RECONNECTS
ARG NATS_RECONNECT_WAIT

# Set environment variables for build
ENV PORT=$PORT \
    NATS_URL=$NATS_URL \
    GIN_MODE=$GIN_MODE \
    DATABASE_URL=$DATABASE_URL \
    NATS_MAX_RECONNECTS=$NATS_MAX_RECONNECTS \
    NATS_RECONNECT_WAIT=$NATS_RECONNECT_WAIT

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o feeti-wallet .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy only the necessary binary from the builder stage
COPY --from=builder /app/feeti-wallet /app/

# Expose the port for the application
EXPOSE 3000

# Run the Go binary
CMD ["/app/feeti-wallet"]
