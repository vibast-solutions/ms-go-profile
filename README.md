# Profile Microservice

`github.com/vibast-solutions/ms-go-profile`

Profile microservice for managing user profiles via HTTP and gRPC interfaces.

## Requirements

- Go 1.25+

## Build

```bash
# Download dependencies
go mod tidy

# Build native binary
make build

# Cross-compile for Linux
make build-linux-arm64
make build-linux-amd64

# Cross-compile for macOS
make build-darwin-arm64
make build-darwin-amd64

# Build all targets
make build-all

# Clean build artifacts
make clean
```

## Run

```bash
# Run directly
go run main.go serve

# Or run the built binary
./build/profile-service serve
```

The service starts:
- HTTP server on 0.0.0.0:8080
- gRPC server on 0.0.0.0:9090

## Configuration

Set environment variables or use defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| HTTP_HOST | 0.0.0.0 | HTTP server bind address |
| HTTP_PORT | 8080 | HTTP server port |
| GRPC_HOST | 0.0.0.0 | gRPC server bind address |
| GRPC_PORT | 9090 | gRPC server port |
| MYSQL_DSN | (required) | MySQL connection string |
| LOG_LEVEL | info | Log level (trace, debug, info, warn, error, fatal, panic) |
| MYSQL_MAX_OPEN_CONNS | 10 | Max open DB connections |
| MYSQL_MAX_IDLE_CONNS | 5 | Max idle DB connections |
| MYSQL_CONN_MAX_LIFETIME_MINUTES | 30 | Max connection lifetime in minutes |

## Health Check

- `GET /health` returns `{ "status": "ok" }`

## HTTP APIs

### Profiles

- `POST /profiles`
- `GET /profiles/:id`
- `GET /profiles/user/:user_id`
- `PUT /profiles/:id`
- `DELETE /profiles/:id`

### Contacts

- `POST /contacts`
- `GET /contacts/:id`
- `PUT /contacts/:id`
- `DELETE /contacts/:id`
- `GET /contacts?profile_id=<id>&page=<n>&page_size=<n>`

### Addresses

- `POST /addresses`
- `GET /addresses/:id`
- `PUT /addresses/:id`
- `DELETE /addresses/:id`
- `GET /addresses?profile_id=<id>&page=<n>&page_size=<n>`

Address request fields:
- Mandatory: `street_name`, `streen_no`, `city`, `county`, `country`, `profile_id`
- Optional: `postal_code`, `building`, `apartment`, `additional_data` (max 512), `type`

## gRPC

Generate protobuf/grpc files:

```bash
./scripts/gen_proto.sh
```

Service methods:

- Profile: `CreateProfile`, `GetProfile`, `GetProfileByUserID`, `UpdateProfile`, `DeleteProfile`
- Contact: `CreateContact`, `GetContact`, `UpdateContact`, `DeleteContact`, `ListContacts`
- Address: `CreateAddress`, `GetAddress`, `UpdateAddress`, `DeleteAddress`, `ListAddresses`

## E2E Tests

Profile includes Docker Compose based e2e tests in `profile/e2e` for profiles, contacts, and addresses.

Host port bindings are intentionally non-default:
- MySQL: `23306` -> container `3306`
- HTTP: `28080` -> container `8080`
- gRPC: `29090` -> container `9090`

Run e2e tests:

```bash
cd e2e
docker compose up -d --build
cd ..
PROFILE_HTTP_URL=http://localhost:28080 PROFILE_GRPC_ADDR=localhost:29090 go test ./e2e -v -tags e2e
```

Or use:

```bash
./e2e/run.sh
```
