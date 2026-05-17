# Deployment Guide

## Docker Compose

1. Copy `.env.production` and replace secrets and database passwords.
2. Start the stack:

```bash
docker compose up -d --build
```

3. Open:

```text
http://localhost
```

## Production Checklist

- Use strong `APP_SECRET` and `JWT_SECRET`.
- Put Nginx behind HTTPS or add TLS certificates directly.
- Restrict PostgreSQL and Redis to the private network.
- Use managed backups or replace the backup marker job with `pg_dump`.
- Run SQL migrations with a migration tool in controlled environments.
- Monitor logs, disk usage, database locks, and slow queries.

## Architecture

```mermaid
flowchart LR
  U[Users] --> N[Nginx]
  N --> G[Gin ERP App]
  G --> P[(PostgreSQL)]
  G --> R[(Redis Cache)]
  G --> J[Background Jobs]
  J --> P
  G --> UI[HTML Tailwind UI]
  G --> API[/api/v1 JSON API]
```
