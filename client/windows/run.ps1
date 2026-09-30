#requires -RunAsAdministrator
param(
  [string]$Config = "$env:ProgramData\Jamshidix\client.json"
)

$ErrorActionPreference = "Stop"

$exe = Join-Path $env:ProgramFiles "Jamshidix\sing-box.exe"

if (-not (Test-Path $exe)) { throw "sing-box.exe is not installed." }
if (-not (Test-Path $Config)) { throw "Client config not found: $Config" }

& $exe check -c $Config
if ($LASTEXITCODE -ne 0) { throw "Invalid client configuration." }

Write-Host "Starting Jamshidix TUN. Press Ctrl+C to stop."
& $exe run -c $Config
