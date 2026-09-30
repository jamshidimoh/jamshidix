#requires -RunAsAdministrator
$ErrorActionPreference = "Stop"

$group = "Jamshidix KillSwitch"
$program = Join-Path $env:ProgramFiles "Jamshidix\sing-box.exe"
$stateDir = Join-Path $env:ProgramData "Jamshidix"
$stateFile = Join-Path $stateDir "firewall-state.json"

if (-not (Test-Path $program)) { throw "Install sing-box first." }

New-Item -ItemType Directory -Force -Path $stateDir | Out-Null

if (-not (Test-Path $stateFile)) {
  Get-NetFirewallProfile |
    Select-Object Name, DefaultOutboundAction |
    ConvertTo-Json |
    Set-Content -Path $stateFile -Encoding UTF8
}

Get-NetFirewallRule -DisplayGroup $group -ErrorAction SilentlyContinue | Remove-NetFirewallRule

New-NetFirewallRule -DisplayGroup $group -DisplayName "$group - Allow tunnel" -Direction Outbound -Action Allow -Profile Any -InterfaceAlias "JamshidixTunnel"
New-NetFirewallRule -DisplayGroup $group -DisplayName "$group - Allow sing-box" -Direction Outbound -Action Allow -Profile Any -Program $program
New-NetFirewallRule -DisplayGroup $group -DisplayName "$group - Allow DHCP" -Direction Outbound -Action Allow -Profile Any -Protocol UDP -LocalPort 68 -RemotePort 67

Set-NetFirewallProfile -Profile Domain,Public,Private -DefaultOutboundAction Block

Write-Host "Jamshidix kill-switch ENABLED."
