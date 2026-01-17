# Step 1: Build the binary
FROM golang:1.21-alpine AS builder

# Install git for any external dependencies
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum (if you have them) and download dependencies
COPY go.mod ./
# RUN go mod download # Uncomment if you have external dependencies

# Copy the source code
COPY . .

# Build the application
# We build a static binary so it can run in a tiny "scratch" image
RUN CGO_ENABLED=0 GOOS=linux go build -o dns-server ./cmd/server/main.go

# Step 2: Final thin image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/dns-server .

# Copy your initial config file
COPY config.json .

# Expose the DNS port (UDP 53)
# Note: Inside the container, we might still use 8083 or switch to 53
EXPOSE 8083/udp

# Run the server
CMD ["./dns-server"]