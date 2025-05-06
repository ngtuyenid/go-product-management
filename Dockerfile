# Build stage
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Use a small alpine image for the final container
FROM alpine:latest

WORKDIR /

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /api /api
# Copy the .env file
COPY .env /.env

# Expose the API port
EXPOSE 8080

# Run the application
CMD ["/api"] 