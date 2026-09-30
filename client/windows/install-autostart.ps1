#requires -RunAsAdministrator
param(
  [string]$Config = "$env:ProgramData\Jamshidix\client.json"
)

$ErrorActionPreference = "Stop"

$exe = Join-Path $env:ProgramFiles "Jamshidix\sing-box.exe"

if (-not (Test-Path $exe)) { throw "sing-box.exe not found. Run install.ps1 first." }
if (-not (Test-Path $Config)) { throw "Client config not found: $Config" }

& $exe check -c $Config
if ($LASTEXITCODE -ne 0) { throw "Invalid client configuration." }

$action = New-ScheduledTaskAction -Execute $exe -Argument ('run -c "' + $Config + '"')
$trigger = New-ScheduledTaskTrigger -AtStartup
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest

Register-ScheduledTask -TaskName "Jamshidix" -Action $action -Trigger $trigger -Principal $principal -Force | Out-Null

Write-Host "Jamshidix autostart task installed."
