#requires -RunAsAdministrator
$ErrorActionPreference = "Stop"

Unregister-ScheduledTask -TaskName "Jamshidix" -Confirm:$false -ErrorAction SilentlyContinue
Write-Host "Jamshidix autostart task removed."
