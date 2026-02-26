# Video Conference Backend Service

This is the backend service for the Video Conference application. It handles user authentication, room management, roles, and integration with LiveKit for video conferencing. It is built with Go using the Gin framework.

## Tech Stack
- **Go** (1.25)
- **Frameworks/Routers:** Gin
- **Database:** PostgreSQL (via GORM)
- **Caching/State:** Redis
- **Storage:** MinIO
- **Video Conferencing:** LiveKit Server SDK
- **Authentication:** JWT, OIDC (Google)

## Prerequisites
- [Go](https://golang.org/doc/install) 1.25+
- PostgreSQL
- Redis
- MinIO
- LiveKit Server

## Environment Variables
Create a `.env` file in the root of the service based on your setup. A typical `.env` looks like:

```env
# APP
PORT=8080
JWT_SECRET=your_super_secret_key

# DATABASE
DB_DSN=host=localhost port=5432 user=postgres password=admin.admin dbname=meet_db sslmode=disable TimeZone=Asia/Jakarta

DB_HOST=192.168.100.130
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=admin.admin
DB_NAME=meet_db

# LIVEKIT
LIVEKIT_SERVER_URL=http://livekit.example.com/
LIVEKIT_API_KEY=your_livekit_api_key
LIVEKIT_API_SECRET=your_livekit_api_secret

# MINIO / S3
MINIO_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=admin
MINIO_SECRET_KEY=admin.admin
MINIO_BUCKET=livekit
MINIO_REGION=us-east-1

# REDIS
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=admin.admin
```

## Running Locally

1. Install dependencies:
   ```bash
   go mod tidy
   ```
2. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

*Note: The service uses GORM auto-migration, so tables will be created automatically on startup.*

## Running with Docker

You can run the application via Docker if you have set up a Dockerfile for it.

```bash
# Build the image
docker build -t video-conference-be .

# Run the container (make sure to link it to your DB/Redis networks as needed)
docker run -d -p 8080:8080 --env-file .env --name vc-backend video-conference-be
```

## Architecture / Structure

- `cmd/server/main.go`: Entry point of the application.
- `config/`: Configurations and environment loading.
- `internal/`:
  - `app/`: Application business logic (`service`, `repository`, `http` handlers/routers).
  - `domain/`: Core struct models and domain logic.
- `middleware/`: Gin middleware for Auth, CORS, Logging, etc.
- `pkg/`: Shared utility packages, DB initialization, utilities.
