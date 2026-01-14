# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# Adjust the build path if your main file is located elsewhere. 
# Based on your project structure, the main entry point seems to be cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Start a new stage from scratch
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/server .

# Copy .env file (Optional: It's often better to pass env vars at runtime, but for simplicity copying here)
COPY --from=builder /app/.env .

# Expose port (adjust if your app uses a different port, e.g. from .env)
# As per main.go, it uses utility.Config.Port. Commonly 8080 or 3000.
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
