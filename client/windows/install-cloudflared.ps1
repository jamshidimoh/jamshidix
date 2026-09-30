#requires -RunAsAdministrator
param(
  [string]$Version = "2026.9.3"
)

$ErrorActionPreference = "Stop"

$url = "https://github.com/cloudflare/cloudflared/releases/download/$Version/cloudflared-windows-amd64.exe"
$expectedSha256 = "2837888cc0f5d58f15b6dc478376de90b4d3ba5241c7947455d1e0a0df429712"
$installDir = Join-Path $env:ProgramFiles "Cloudflared"
$exe = Join-Path $installDir "cloudflared.exe"
$tmp = Join-Path $env:TEMP "cloudflared-$Version.exe"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Invoke-WebRequest -Uri $url -OutFile $tmp

$actualSha256 = (Get-FileHash -Algorithm SHA256 -Path $tmp).Hash.ToLowerInvariant()
if ($actualSha256 -ne $expectedSha256) {
  Remove-Item $tmp -Force -ErrorAction SilentlyContinue
  throw "cloudflared SHA256 mismatch. Expected $expectedSha256, got $actualSha256."
}

Copy-Item $tmp $exe -Force
Remove-Item $tmp -Force
& $exe --version
Write-Host "cloudflared $Version installed and hash verified."
Write-Host "Tunnel activation is separate; do not commit a tunnel token."
