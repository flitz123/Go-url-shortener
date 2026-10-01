$ErrorActionPreference = "Stop"

if (-not $env:DATABASE_URL) {
    $env:DATABASE_URL = "postgres://user:password@localhost:5432/urlshort?sslmode=disable"
}
if (-not $env:REDIS_ADDR) {
    $env:REDIS_ADDR = "localhost:6379"
}

foreach ($port in @(5432, 6379)) {
    $available = Test-NetConnection -ComputerName localhost -Port $port -InformationLevel Quiet -WarningAction SilentlyContinue
    if (-not $available) {
        throw "Required local service is not listening on port $port. Start PostgreSQL (5432) and Redis (6379), or set DATABASE_URL and REDIS_ADDR."
    }
}

go run ./cmd/server
exit $LASTEXITCODE