param(
  [Parameter(Mandatory = $true)][string]$BackupPath,
  [string]$Sha256Path = "$BackupPath.sha256"
)

$ErrorActionPreference = "Stop"
if (Test-Path $Sha256Path) {
  $expected = (Get-Content $Sha256Path -Raw).Split(" ", [System.StringSplitOptions]::RemoveEmptyEntries)[0]
  $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath "$BackupPath").Hash
  if ($expected -ne $actual) { throw "Checksum mismatch" }
}

if (-not $BackupPath.EndsWith(".age")) {
  pg_restore --list "$BackupPath" | Out-Null
}

Write-Host "Backup verification passed: $BackupPath"
