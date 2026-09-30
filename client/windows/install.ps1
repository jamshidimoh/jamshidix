param(
  [string]$Version = "1.14.2"
)

$ErrorActionPreference = "Stop"

$base = "https://github.com/SagerNet/sing-box/releases/download/v$Version"
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { throw "32-bit Windows is unsupported." }

$archive = "sing-box-$Version-windows-$arch.zip"
$tmp = Join-Path $env:TEMP "jamshidix-$Version"

New-Item -ItemType Directory -Force -Path $tmp | Out-Null
Invoke-WebRequest -Uri "$base/$archive" -OutFile (Join-Path $tmp $archive)
Expand-Archive -Path (Join-Path $tmp $archive) -DestinationPath $tmp -Force

$exe = Get-ChildItem -Path $tmp -Recurse -Filter "sing-box.exe" | Select-Object -First 1
if (-not $exe) { throw "sing-box.exe not found." }

$install = Join-Path $env:ProgramFiles "Jamshidix"
New-Item -ItemType Directory -Force -Path $install | Out-Null
Copy-Item $exe.FullName (Join-Path $install "sing-box.exe") -Force

Write-Host "Installed sing-box $Version at $install"
Write-Host "Create client.json from the project template before starting."
