# SocialForge

A powerful Opinion Control System (OCS) built with Go, gRPC, Ent, and Auth0 for automated social media posting and influencer management.

## Features

- Automated Posting System
- Influencer Management
- Scheduled Post Configuration
- Real-time Status Monitoring
- Secure Authentication with Auth0

## Tech Stack

- Go
- gRPC
- Ent (Entity Framework)
- Auth0
- PostgreSQL

## Project Structure

```
.
├── cmd/            # Application entrypoints
├── internal/       # Private application code
├── proto/         # Protocol buffer definitions
└── pkg/           # Public libraries
```

## Getting Started

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

## API Documentation

The API is defined using Protocol Buffers and served via gRPC. See the `proto/` directory for detailed API specifications.