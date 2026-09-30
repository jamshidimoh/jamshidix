param(
  [Parameter(Mandatory=$true)][string]$ServerIp,
  [Parameter(Mandatory=$true)][string]$Uuid,
  [Parameter(Mandatory=$true)][string]$RealityPublicKey,
  [Parameter(Mandatory=$true)][ValidatePattern('^[0-9a-fA-F]{1,8}$')][string]$ShortId,
  [Parameter(Mandatory=$true)][ValidatePattern('^[A-Za-z0-9.-]+$')][string]$HandshakeHost,
  [string]$Template = "$PSScriptRoot\..\..\config\client.config.template.json",
  [string]$Output = "$env:ProgramData\Jamshidix\client.json"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $Template)) { throw "Template not found: $Template" }

$templateText = Get-Content -Raw -Path $Template

$replacements = @{
  "REPLACE_WITH_SERVER_IP" = $ServerIp
  "REPLACE_WITH_UUID" = $Uuid
  "REPLACE_WITH_REALITY_PUBLIC_KEY" = $RealityPublicKey
  "REPLACE_WITH_SHORT_ID" = $ShortId
  "REPLACE_WITH_HANDSHAKE_HOST" = $HandshakeHost
}

foreach ($item in $replacements.GetEnumerator()) {
  $templateText = $templateText.Replace($item.Key, $item.Value)
}

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Output) | Out-Null
$templateText | Set-Content -Path $Output -Encoding UTF8

$exe = Join-Path $env:ProgramFiles "Jamshidix\sing-box.exe"
if (Test-Path $exe) {
  & $exe check -c $Output
  if ($LASTEXITCODE -ne 0) {
    Remove-Item $Output -Force
    throw "Generated client config failed sing-box validation."
  }
}

Write-Host "Client configuration created: $Output"
