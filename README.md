# Linkline URL shortener

A small same-origin URL shortener with a responsive dashboard, optional custom aliases, click counts, PostgreSQL persistence, and Redis redirect caching.

## Run locally

### Windows without WSL or Docker

Install Go, PostgreSQL, and a Redis-compatible Windows service (for example, Memurai). Start PostgreSQL and Redis so they listen on `localhost:5432` and `localhost:6379`. In PostgreSQL, create the development role and database:

```sql
CREATE USER user WITH PASSWORD 'password';
CREATE DATABASE urlshort OWNER user;
```

Then run the app in PowerShell:

```powershell
.\run-windows.ps1
```

The script uses local defaults and checks that both service ports are reachable. Set `DATABASE_URL` or `REDIS_ADDR` in the current PowerShell session first if your installation uses different credentials or ports. Open [http://localhost:8080](http://localhost:8080).

### Docker Compose

Alternatively, start the full containerized stack:

```sh
docker compose up --build
```

Open [http://localhost:8080](http://localhost:8080). Compose starts PostgreSQL and Redis, waits for both health checks, and the application creates the `urls` table on startup. Link history is kept in the current browser; click counts are read from PostgreSQL. Use `docker compose down` to stop the stack. The named `postgres_data` volume keeps stored links between restarts.

The Go server defaults to native Windows/local service addresses. Compose explicitly overrides them with its internal service names.

## API

Create a short link:

```http
POST /api/urls
Content-Type: application/json

{"url":"https://example.com/a-long-page","custom_alias":"launch"}
```

`custom_alias` is optional and may contain 3-32 letters, digits, hyphens, or underscores. The response includes `code`, `original`, `short_url`, and `clicks`. A duplicate alias returns `409 Conflict`; invalid URLs return `400 Bad Request`.

Read a link and its click count with `GET /api/urls/{code}`. Visiting `/{code}` redirects to the destination and records a visit. `GET /healthz` is the application health endpoint; `/metrics` exposes Prometheus metrics.