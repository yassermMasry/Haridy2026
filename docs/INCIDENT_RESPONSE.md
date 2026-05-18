# Incident Response

## Severity

- SEV1: outage, data loss, confirmed breach, or accounting corruption.
- SEV2: degraded POS/API, failed backups, tenant isolation risk, or repeated reconciliation findings.
- SEV3: isolated errors without customer impact.

## First 15 Minutes

1. Assign incident lead.
2. Freeze risky deploys.
3. Capture current `/metrics`, application logs, database status, and recent `security_events`.
4. Communicate impact and next update time.

## Security Events

Review `security_events` for brute-force blocks, login failures, suspicious IPs, and unusual user agents. Rotate `APP_SECRETS` and `JWT_SECRETS` by prepending the new secret and keeping the old value later in the comma-separated list until sessions and tokens age out.

## Accounting Integrity

Run reconciliation jobs after any interrupted posting flow. Treat unbalanced entries, duplicate references, negative stock, or negative treasury balances as SEV1 until explained.
