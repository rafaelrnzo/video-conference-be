# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o seeder ./cmd/seeder/main.go

# Start a new stage from scratch
FROM alpine:latest

# --- FIX START ---
# Install ca-certificates AND tzdata (required for TimeZone=Asia/Jakarta)
RUN apk --no-cache add ca-certificates tzdata

# Set the timezone globally inside the container
ENV TZ=Asia/Jakarta
# --- FIX END ---

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/server .
COPY --from=builder /app/seeder .

# Copy .env file
COPY --from=builder /app/.env .

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./server"]