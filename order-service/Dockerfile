# Build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# Install necessary dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project to ensure all dependencies (including e-commerce/order-service/orderhand) are available
COPY . .

# Set the working directory to the order-service folder
WORKDIR /app/order-service

# Build the binary using order.go instead of main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/order-service/order-service ./order.go

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the built binary from the builder stage
COPY --from=builder /app/order-service/order-service /app/order-service/order-service

# Ensure the binary has execution permissions
RUN chmod +x /app/order-service/order-service

# Expose the service port (matches r.Run(":8083") in your order.go)
EXPOSE 8083

# Set timezone (optional)
ENV TZ=UTC

# Run the binary (pointing directly to the executable file)
ENTRYPOINT ["/app/order-service/order-service"]
