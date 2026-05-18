# Haridy ERP SaaS

Commercial multi-tenant ERP platform built with Go, Gin, GORM, PostgreSQL, Redis, server-rendered HTML, and a separated React frontend starter.

## Commercial SaaS Scope

- Multi-tenant architecture: tenants, plans, subscriptions, tenant users, company settings, subdomain/header tenant resolution.
- Tenant isolation foundation: tenant-aware models and API query filtering for core endpoints.
- ERP modules: inventory, warehouses, sales, purchases, treasury, customers, suppliers, returns, vouchers, accounting, reports.
- Financial statements: trial balance, income/cash/balance summary API, fiscal years, periods, opening balances, closing entries.
- E-invoicing foundation: VAT fields, QR payload, XML export model, invoice branding path.
- Workflow engine foundation: approval workflows, steps, and logs for purchases, expenses, and returns.
- Advanced inventory: serial numbers, batches, expiry dates, FIFO/Average valuation layers, barcode-ready model.
- CRM: activities, reminders, followups, sales pipeline, deals.
- HR basics: employees, attendance, payroll runs.
- Mobile-ready API: JWT auth, refresh tokens, device registration and push token hook.
- Observability: `/healthz`, `/readyz`, `/metrics`, structured logs, trace headers.
- Security: CSP/HSTS headers, secure production cookies, CSRF for web, JWT rotation support, rate limiting, brute force protection, password policy validator, audited security events.
- Production hardening: accounting/inventory reconciliation jobs, duplicate posting protection, non-negative balance guards, encrypted PostgreSQL backup scripts, Prometheus/Grafana starter configs.
- Deployment: Docker, Compose, Nginx, GitHub Actions, AWS and DigitalOcean deployment profiles.
- Commercial polish: landing page, onboarding wizard, demo tenant, license table, system updates table.

## Architecture

```mermaid
flowchart LR
  Web[Browser UI] --> Gin[Gin SaaS ERP]
  React[React App] --> API[/api/v1/]
  Mobile[Flutter App] --> API
  Gin --> Services[Domain Services]
  API --> Services
  Services --> DB[(PostgreSQL)]
  Services --> Redis[(Redis)]
  Jobs[Background Jobs] --> Services
  Gin --> Metrics[/metrics]
  Gin --> Tenant[Tenant Resolver]
```

## Run Locally

```powershell
Copy-Item .env.example .env
go mod tidy
go run ./cmd/server
```

Open:

```text
http://localhost:8080
```

Demo login:

```text
admin / admin123
```

Tenant defaults:

```text
X-Tenant: demo
Subdomain: demo.your-domain.com
```

## React Frontend

```powershell
cd frontend
npm install
npm run dev
```

Set `VITE_API_BASE=http://localhost:8080/api/v1` if serving separately.

## API

```http
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/mobile/devices
GET  /api/v1/items
GET  /api/v1/sales
GET  /api/v1/purchases
GET  /api/v1/customers
GET  /api/v1/suppliers
GET  /api/v1/treasury
GET  /api/v1/reports
GET  /api/v1/financial/statements
GET  /api/v1/financial/trial-balance
POST /api/v1/einvoices/:invoice_id
```

Headers:

```http
Authorization: Bearer <token>
X-Tenant: demo
```

## Docker

```powershell
docker compose up -d --build
```

## CI/CD

GitHub Actions workflow:

```text
.github/workflows/ci.yml
```

It runs Go tests, builds Docker image, installs frontend dependencies, and builds React.

## Cloud Profiles

- AWS: `deploy/aws/README.md`
- DigitalOcean: `deploy/digitalocean/README.md`
- Nginx: `deploy/nginx/default.conf`
- Production env: `.env.production`

## Verification

```powershell
go test ./...
```

## Production Operations

- API reference: `docs/API.md`
- Architecture and deployment diagram: `docs/ARCHITECTURE.md`
- Recovery procedures: `docs/RECOVERY.md`
- Incident response: `docs/INCIDENT_RESPONSE.md`
- Backups: `deploy/backup/postgres_backup.ps1`, `postgres_restore.ps1`, `verify_backup.ps1`
- Monitoring: `deploy/monitoring/prometheus.yml`, `alerts.yml`, `grafana-dashboard.json`

## Notes

This is now a hardened SaaS ERP foundation with automated tests, accounting integrity checks, tenant isolation coverage, security controls, backup runbooks, monitoring scaffolding, and production recovery docs. Before onboarding high-volume customers, complete a real staging load test and jurisdiction-specific tax/e-invoice compliance review.
