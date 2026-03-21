# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a static binary for Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server/main.go

# Final stage
FROM alpine:latest

# Add certificates for HTTPS support
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /bin/server .

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["./server"]
