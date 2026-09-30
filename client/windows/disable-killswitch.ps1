#requires -RunAsAdministrator
$ErrorActionPreference = "Stop"

$group = "Jamshidix KillSwitch"
$stateFile = Join-Path $env:ProgramData "Jamshidix\firewall-state.json"

Get-NetFirewallRule -DisplayGroup $group -ErrorAction SilentlyContinue | Remove-NetFirewallRule

if (Test-Path $stateFile) {
  $state = Get-Content -Raw $stateFile | ConvertFrom-Json
  foreach ($profile in @($state)) {
    Set-NetFirewallProfile -Name $profile.Name -DefaultOutboundAction $profile.DefaultOutboundAction
  }
} else {
  Set-NetFirewallProfile -Profile Domain,Public,Private -DefaultOutboundAction Allow
}

Remove-Item $stateFile -Force -ErrorAction SilentlyContinue
Write-Host "Jamshidix kill-switch DISABLED."
