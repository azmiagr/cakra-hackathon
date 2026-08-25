# Cakra Backend API

The REST API for Cakra. It handles user authentication, XLSX sales-data uploads, inventory analysis through Cakra AI, analysis credits, and result history.

## Prerequisites

- Git
- Docker Engine and Docker Compose v2 for the recommended setup
- Go `1.25.5` only when running the backend directly on the host
- Access to MariaDB, Supabase Storage, SMTP, and Cakra AI

## Run with Docker

1. Clone the repository and enter the project directory.

   ```bash
   git clone https://github.com/azmiagr/cakra-hackathon.git
   cd cakra-hackathon
   ```

2. Create a local configuration file.

   ```bash
   cp .env.example .env
   ```

3. Fill in all placeholders in `.env`. The following values are required before Compose can start the `app` service:

   ```env
   DB_NAME=cakra-hackathon
   DB_USER=cakra
   DB_PASSWORD=replace-with-a-strong-password
   DB_ROOT_PASSWORD=replace-with-a-separate-strong-root-password

   JWT_SECRET_KEY=replace-with-a-long-random-secret

   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=you@example.com
   SMTP_PASSWORD=your-smtp-app-password

   SUPABASE_URL=https://your-project.supabase.co
   SUPABASE_TOKEN=your-supabase-service-role-token
   SUPABASE_BUCKET=your-bucket

   AI_CALLBACK_SECRET=replace-with-a-long-random-secret
   AI_BASE_URL=https://chakrai.akademicompetition.id
   AI_REQUEST_TIMEOUT_SECONDS=60
   AI_API_KEY=

   ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
   ```

   `AI_API_KEY` may remain empty while the Cakra AI deployment does not require the `X-API-Key` header. Never commit `.env`.

4. Build and start the database and API.

   ```bash
   docker compose up -d --build
   ```

5. Confirm that both services are healthy.

   ```bash
   docker compose ps
   curl http://localhost:8082/healthz
   ```

   By default, the API is bound to `127.0.0.1`. Use a reverse proxy such as Nginx or Caddy for public HTTPS access.

6. Inspect logs if startup fails.

   ```bash
   docker compose logs --tail=100 app
   docker compose logs --tail=100 db
   ```

To stop the services without deleting MariaDB data:

```bash
docker compose down
```

## Run with Go Locally

Use this option when MariaDB and all external dependencies are already available on the host.

1. Copy `.env.example` to `.env`, then update the local database connection:

   ```env
   DB_HOST=localhost
   DB_PORT=3306
   ADDRESS=localhost
   PORT=8082
   AI_BASE_URL=https://chakrai.akademicompetition.id
   ```

2. Start the application:

   ```bash
   go run ./cmd/app
   ```

   The application runs migrations and seeds during startup. The health check is available at `http://localhost:8082/healthz`.

## Important Configuration

| Variable | Purpose |
| --- | --- |
| `DB_*` | MariaDB connection settings. In Docker, the application automatically uses host `db` and port `3306`. |
| `APP_HOST_PORT` | Host port for the Docker API; defaults to `8082`. |
| `JWT_SECRET_KEY` | Secret used to sign access tokens. Use a long, random value. |
| `SMTP_*` | Credentials used to send registration and password-reset OTP emails. |
| `SUPABASE_*` | Object-storage configuration for uploaded XLSX files. |
| `AI_BASE_URL` | Cakra AI endpoint. Use the public HTTPS URL for development, or `http://cakra-ai:7860` when both containers are on the same Docker network. |
| `AI_REQUEST_TIMEOUT_SECONDS` | Timeout for `POST /predict`; defaults to `60`. |
| `AI_API_KEY` | Optional API key if the AI service enables authentication. |
| `AI_CALLBACK_SECRET` | Secret for the legacy internal callback endpoint; still required for compatibility. |
| `ALLOWED_ORIGINS` | Comma-separated frontend origins, for example `http://localhost:3000,https://app.example.com`. Each value must exactly match the browser origin. |

After changing `.env` in a Docker deployment, recreate the API container to apply the new environment:

```bash
docker compose up -d --force-recreate app
```

## Cakra AI

When an analysis is created, the backend creates a session and reserves one credit first. After the database transaction commits, it sends `POST /predict` to Cakra AI.

- `COMPLETED`: the recommendation is saved and one credit is debited.
- `INSUFFICIENT_DATA`: historical data is insufficient; the reserved credit is released.
- `AI_FAILED`: the AI service is unreachable or returned an invalid response; the reserved credit is released.

To confirm that the API container can reach Cakra AI:

```bash
docker compose exec app sh -c 'wget -q -O- "$AI_BASE_URL/health"'
```

The AI contract is documented in [docs/API_CONTRACT.md](docs/API_CONTRACT.md).

## Analysis Upload Format

The upload endpoint accepts an `.xlsx` file containing one SKU per file. The first sheet must include these headers:

```text
tanggal,jumlah_terjual,nama_sku,harga_satuan
```

Data rules:

- `tanggal`: `YYYY-MM-DD` format, with no duplicate dates.
- `jumlah_terjual`: integer greater than or equal to `0`.
- `nama_sku`: required and identical in every row.
- `harga_satuan`: number greater than or equal to `0`.

For regular demand patterns, the AI needs at least 90 days of historical data before it can generate a forecast.

## API Flow Overview

All analysis endpoints require this header after login:

```http
Authorization: Bearer <access_token>
```

1. `POST /api/v1/auth/register` starts registration and returns `X-Session-Token`.
2. `POST /api/v1/auth/register/verify-otp` sends the previous session token and returns a replacement token in the same header.
3. `POST /api/v1/auth/register/password` sends the token returned after OTP verification.
4. `POST /api/v1/auth/login` returns an access token.
5. `POST /api/v1/analysis/upload` uploads a valid XLSX file.
6. `POST /api/v1/analysis/sessions/:uploadID` runs synchronous analysis through AI.
7. `GET /api/v1/analysis/sessions/:sessionID` retrieves analysis-result details.
8. `GET /api/v1/analysis/history` retrieves completed analysis history.

Example analysis-session request:

```bash
curl -X POST http://localhost:8082/api/v1/analysis/sessions/<upload_id> \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "category_name": "Makanan & Minuman",
    "current_stock": 40,
    "lead_time_days": 3
  }'
```

## Validation and Development

```bash
go test ./...
go vet ./...
go build ./...
```

Main project structure:

```text
cmd/app/                 application composition root and startup
internal/handler/rest/   Gin HTTP handlers and routes
internal/service/        business workflows
internal/repository/     GORM queries and database access
pkg/ai/                  Cakra AI HTTP adapter
pkg/config/              environment loading and validation
```

## VPS Deployment

The GitHub Actions workflow in `.github/workflows/deploy.yml` builds a GHCR image and runs `deploy.sh` over SSH. Configure these repository secrets:

```text
VPS_HOST
VPS_USERNAME
VPS_SSH_KEY
VPS_APP_DIR
CR_PAT
```

`VPS_APP_DIR` must be the absolute path to the repository on the VPS, such as `/home/deploy/project/ui/cakra-hackathon`; do not use `~`. Create the production `.env` directly on the VPS and never store it in Git.
