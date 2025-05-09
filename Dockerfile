# Use the official Golang image
FROM golang:1.23-alpine AS builder
LABEL authors="999iQ"
WORKDIR /app
# Install dependencies
RUN apk add --no-cache git
# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download
# Copy the source code
COPY . .
# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer ./cmd

# stage 2
FROM alpine:3.19
# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates
WORKDIR /app
# Copy the binary from the builder stage
COPY --from=builder /app/loadbalancer /app/loadbalancer
COPY config.yaml .
EXPOSE 8080
# Set entry point
ENTRYPOINT ["/app/loadbalancer"]