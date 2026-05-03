# Use the official Golang image to create a build artifact.
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the local code to the container
COPY . .

# Build the Go application
# CGO_ENABLED=0 creates a statically linked binary (better for alpine/scratch)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Use a lean alpine image for the final stage
FROM alpine:latest

WORKDIR /app

# Add CA certificates for HTTPS requests (needed to talk to Supabase)
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose port 8080
EXPOSE 8080

# Run the web service on container startup.
CMD ["./server"]