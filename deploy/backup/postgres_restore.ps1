param(
  [Parameter(Mandatory = $true)][string]$BackupPath,
  [string]$DatabaseUrl = $env:DATABASE_URL,
  [string]$AgeIdentity = $env:BACKUP_AGE_IDENTITY
)

$ErrorActionPreference = "Stop"
if (-not $DatabaseUrl) { throw "DATABASE_URL is required" }

$restoreFile = $BackupPath
if ($BackupPath.EndsWith(".age")) {
  if (-not $AgeIdentity) { throw "BACKUP_AGE_IDENTITY is required for encrypted backups" }
  $restoreFile = [System.IO.Path]::GetTempFileName()
  age -d -i "$AgeIdentity" -o "$restoreFile" "$BackupPath"
}

pg_restore --clean --if-exists --no-owner --no-acl --dbname="$DatabaseUrl" "$restoreFile"
Write-Host "Restore completed from $BackupPath"
