#!/usr/bin/env pwsh

Write-Host -ForegroundColor Green "Build TunnelSSH"
$TopLevel = Split-Path -Path $PSScriptRoot
$env:GO111MODULE="on"
$ps = Start-Process -FilePath "go" -WorkingDirectory "$env:TEMP" -ArgumentList "get github.com/balibuild/bali/cmd/bali" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}
## remove
Remove-Item Env:GO111MODULE

$ps = Start-Process -FilePath "bali" -WorkingDirectory $TopLevel -ArgumentList "-z" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Write-Host -ForegroundColor Green "build tunnelssh success"
