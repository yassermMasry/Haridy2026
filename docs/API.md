# API Reference

Base path: `/api/v1`

Authentication uses `Authorization: Bearer <jwt>`. Tenant scope is resolved from `X-Tenant`, subdomain, or the demo tenant fallback.

## Auth

- `POST /auth/login` returns `token` and `refresh_token`.
- `POST /auth/refresh` returns a fresh access token.

## Resources

- `GET /items`
- `GET /sales`
- `GET /purchases`
- `GET /customers`
- `GET /suppliers`
- `GET /treasury`
- `GET /reports`
- `GET /financial/statements`
- `GET /financial/trial-balance`
- `POST /mobile/devices`
- `POST /einvoices/:invoice_id`

All list endpoints are tenant filtered and capped to protect production latency. Add explicit pagination before exposing larger exports.
