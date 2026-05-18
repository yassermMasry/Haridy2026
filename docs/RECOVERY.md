# Recovery Procedures

## Backup

Run encrypted PostgreSQL backups:

```powershell
$env:DATABASE_URL="postgres://user:pass@host:5432/haridy?sslmode=require"
$env:BACKUP_S3_URI="s3://company-haridy-backups/prod"
$env:BACKUP_AGE_PUBLIC_KEY="age1..."
.\deploy\backup\postgres_backup.ps1
```

## Verify

```powershell
.\deploy\backup\verify_backup.ps1 -BackupPath .\storage\backups\haridy-YYYYMMDD-HHMMSS.dump
```

## Restore

Restore only into an explicitly selected target database:

```powershell
$env:DATABASE_URL="postgres://user:pass@restore-host:5432/haridy_restore?sslmode=require"
.\deploy\backup\postgres_restore.ps1 -BackupPath .\storage\backups\haridy-YYYYMMDD-HHMMSS.dump
```

## Recovery Targets

- RPO: 24 hours until a shorter backup schedule is configured.
- RTO: 2 hours for database restore plus smoke testing.
- Always run `/healthz`, `/readyz`, login, sale creation, purchase creation, and trial balance checks after restore.
