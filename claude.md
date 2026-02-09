# Profile Microservice - Claude Context

## Overview
Profile CRUD microservice. **This is the canonical reference implementation** for the project's architecture patterns. See `development.md` at the repo root for the full coding guide.

## Technology Stack
- **Framework**: Echo (HTTP), gRPC
- **CLI**: Cobra
- **Database**: MySQL with raw `database/sql`
- **Logging**: Logrus (JSON, structured)
- **Configuration**: Environment-based with optional `.env` support

## Module
- **Path**: `github.com/vibast-solutions/ms-go-profile`

## Directory Structure
```
profile/
├── main.go
├── Makefile
├── cmd/
│   ├── root.go             # Cobra root command
│   ├── serve.go            # HTTP + gRPC server setup, dependency wiring
│   ├── logging.go          # Logrus JSON configuration
│   └── version.go          # Version command (ldflags)
├── config/
│   ├── config.go           # Environment variable loading
│   └── config_test.go
├── app/
│   ├── entity/
│   │   └── profile.go      # Profile domain struct (pure data, no tags)
│   ├── repository/
│   │   ├── profile.go      # MySQL CRUD, DBTX interface, sentinel errors
│   │   └── profile_test.go
│   ├── service/
│   │   ├── profile.go      # Business logic, interface deps, sentinel errors
│   │   └── profile_test.go
│   ├── controller/
│   │   ├── profile.go      # Echo HTTP handlers
│   │   └── profile_test.go
│   ├── grpc/
│   │   ├── server.go       # gRPC handlers
│   │   ├── server_test.go
│   │   ├── interceptor.go  # RequestID, Logging, Recovery interceptors
│   │   └── interceptor_test.go
│   ├── types/
│   │   ├── profile.go      # Request parsing (NewXxxFromContext) + Validate()
│   │   ├── profile_test.go
│   │   ├── profile.pb.go   # Generated protobuf types (reused as service DTOs)
│   │   └── profile_grpc.pb.go
│   ├── dto/
│   │   └── response.go     # ErrorResponse, DeleteResponse
│   └── factory/
│       ├── logger.go       # NewModuleLogger, LoggerWithContext
│       └── logger_test.go
├── proto/
│   └── profile.proto
└── scripts/
    └── gen_proto.sh
```

## API Endpoints (HTTP)
| Method | Path | Description |
|--------|------|-------------|
| POST | /profiles | Create profile |
| GET | /profiles/:id | Get profile by ID |
| GET | /profiles/user/:user_id | Get profile by user ID |
| PUT | /profiles/:id | Update profile |
| DELETE | /profiles/:id | Delete profile |
| GET | /health | Health check |

## gRPC Service
Same CRUD operations on port 9090. See `proto/profile.proto`.

## Configuration (Environment Variables)
- `HTTP_HOST` / `HTTP_PORT` (default: 0.0.0.0:8080)
- `GRPC_HOST` / `GRPC_PORT` (default: 0.0.0.0:9090)
- `MYSQL_DSN` (required)
- `MYSQL_MAX_OPEN_CONNS` (default: 10), `MYSQL_MAX_IDLE_CONNS` (default: 5), `MYSQL_CONN_MAX_LIFETIME_MINUTES` (default: 30)
- `LOG_LEVEL` (default: info)

## Key Patterns Demonstrated
- gRPC protobuf types reused as service-layer DTOs (no duplicate request structs)
- Service depends on interfaces defined in service package (duck typing)
- Request parsing: `NewXxxRequestFromContext()` factories in types/ package
- Module logger + per-request context enrichment via factory/ package
- Request ID middleware: `rest-{uuid}` (HTTP) / `grpc-{uuid}` (gRPC)
- Three-layer error mapping: repository → service → transport
- Repository returns `(nil, nil)` for not-found, service decides if it's an error

## Build
- `make build` — native binary to `build/profile-service`
- `make build-all` — cross-compile for linux/darwin arm64/amd64
- Version and commit hash injected via `-ldflags`
