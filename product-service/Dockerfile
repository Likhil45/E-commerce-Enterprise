# Build stage
FROM golang:1.23.2-alpine AS builder  

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy the go.mod and go.sum from the monorepo root
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the entire project (monorepo) to ensure dependencies are available
COPY . .

# Change to the write-db-service directory
WORKDIR /app/product-service

# Build the binary correctly
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/product-service/product-service ./main.go

# Final stage: Create a minimal image
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage (Ensure correct path)
COPY --from=builder /app/product-service/product-service /app/product-service

# Ensure the binary has execute permissions
RUN chmod +x /app/product-service

# Expose the service port
EXPOSE 8081

# Set timezone (optional)
ENV TZ=UTC

# Run the binary
ENTRYPOINT ["/app/product-service"]
