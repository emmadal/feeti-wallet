FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Define build arguments
ARG PORT
ARG GIN_MODE
ARG NATS_URL
ARG DATABASE_URL

# Set environment variables for build
ENV PORT=$PORT \
    GIN_MODE=$GIN_MODE \
    DATABASE_URL=$DATABASE_URL \
    NATS_URL=$NATS_URL

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o feeti-wallet-service .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy only the necessary binary from the builder stage
COPY --from=builder /app/feeti-wallet-service /app/

# Expose the port for the application
EXPOSE 4000

# Run the Go binary
CMD ["/app/feeti-wallet-service"]
