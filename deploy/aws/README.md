# AWS Deployment Profile

- ECS Fargate for the Go API container.
- RDS PostgreSQL for tenant data.
- ElastiCache Redis for caching and rate-limit state.
- S3 for backups, invoice PDFs, and exported reports.
- CloudFront + ACM + ALB for HTTPS and custom tenant domains.
- Secrets Manager for `APP_SECRET`, `JWT_SECRET`, database password, and S3 credentials.

Environment profile: `.env.production` with AWS-managed hostnames and secrets injected at runtime.
