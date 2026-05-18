param(
  [string]$OutputDir = "storage/backups",
  [string]$DatabaseUrl = $env:DATABASE_URL,
  [string]$S3Uri = $env:BACKUP_S3_URI,
  [string]$AgePublicKey = $env:BACKUP_AGE_PUBLIC_KEY
)

$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$dump = Join-Path $OutputDir "haridy-$stamp.dump"
$encrypted = "$dump.age"

if (-not $DatabaseUrl) { throw "DATABASE_URL is required" }
pg_dump --format=custom --no-owner --no-acl --dbname="$DatabaseUrl" --file="$dump"
pg_restore --list "$dump" | Out-Null

if ($AgePublicKey) {
  age -r "$AgePublicKey" -o "$encrypted" "$dump"
  Remove-Item -LiteralPath "$dump"
  $artifact = $encrypted
} else {
  $artifact = $dump
}

$checksum = Get-FileHash -Algorithm SHA256 -LiteralPath "$artifact"
"$($checksum.Hash)  $artifact" | Set-Content -Path "$artifact.sha256"

if ($S3Uri) {
  aws s3 cp "$artifact" "$S3Uri/"
  aws s3 cp "$artifact.sha256" "$S3Uri/"
}

Write-Host "Backup created: $artifact"
