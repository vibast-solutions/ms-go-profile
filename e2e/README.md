E2E Tests (Docker Compose)

Prereqs:
- Docker Desktop (or Docker Engine) running.

Host port bindings are intentionally non-default to avoid collisions:
- MySQL: `23306` -> container `3306`
- Profile HTTP: `28080` -> container `8080`
- Profile gRPC: `29090` -> container `9090`

Run:
1. cd profile/e2e
2. docker compose up -d --build
3. cd ..
4. PROFILE_HTTP_URL=http://localhost:28080 PROFILE_GRPC_ADDR=localhost:29090 go test ./e2e -v -tags e2e

Shortcut:
- profile/e2e/run.sh

Teardown:
- cd profile/e2e
- docker compose down -v
