#requires -RunAsAdministrator
param(
  [string]$Version = "1.14.1"
)

$ErrorActionPreference = "Stop"

if (-not [Environment]::Is64BitOperatingSystem) {
  throw "32-bit Windows is unsupported."
}

$base = "https://github.com/SagerNet/sing-box/releases/download/v$Version"
$archive = "sing-box-$Version-windows-amd64.zip"
$expectedSha256 = "5197f16d492d93202dc623622149a6ed040f8eca263128f91d603f2b901baa89"

$tmp = Join-Path $env:TEMP "jamshidix-$Version"
$install = Join-Path $env:ProgramFiles "Jamshidix"
$programData = Join-Path $env:ProgramData "Jamshidix"

New-Item -ItemType Directory -Force -Path $tmp | Out-Null

$archivePath = Join-Path $tmp $archive
Invoke-WebRequest -Uri "$base/$archive" -OutFile $archivePath

$actualSha256 = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLowerInvariant()
if ($actualSha256 -ne $expectedSha256) {
  throw "sing-box archive SHA256 mismatch. Expected $expectedSha256, got $actualSha256."
}

Expand-Archive -Path $archivePath -DestinationPath $tmp -Force

$exe = Get-ChildItem -Path $tmp -Recurse -Filter "sing-box.exe" | Select-Object -First 1
if (-not $exe) { throw "sing-box.exe not found." }

New-Item -ItemType Directory -Force -Path $install | Out-Null
New-Item -ItemType Directory -Force -Path $programData | Out-Null
Copy-Item $exe.FullName (Join-Path $install "sing-box.exe") -Force

Write-Host "Installed sing-box $Version."
Write-Host "Binary: $(Join-Path $install 'sing-box.exe')"
Write-Host "Config directory: $programData"
Write-Host "Next: render client.json, validate it, then enable the TUN tunnel as Administrator."
