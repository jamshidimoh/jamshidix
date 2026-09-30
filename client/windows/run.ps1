param(
  [string]$Config = "$PSScriptRoot\client.json"
)

$ErrorActionPreference = "Stop"
$exe = Join-Path $env:ProgramFiles "Jamshidix\sing-box.exe"

if (-not (Test-Path $exe)) { throw "Run install.ps1 first." }
if (-not (Test-Path $Config)) { throw "Client config not found: $Config" }

& $exe check -c $Config
if ($LASTEXITCODE -ne 0) { throw "Invalid sing-box configuration." }

& $exe run -c $Config
