#!/usr/bin/env pwsh

Write-Host -ForegroundColor Green "Build TunnelSSH"
$TopLevel = Split-Path -Path $PSScriptRoot
$ps = Start-Process -FilePath "go" -WorkingDirectory "$env:TEMP" -ArgumentList "install github.com/balibuild/bali/v2/cmd/bali@latest" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

$ps = Start-Process -FilePath "bali" -WorkingDirectory $TopLevel -ArgumentList "-z" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

$hash = Get-FileHash "*.zip" -Algorithm SHA256
$hashtext=$hash.Algorithm + ":" + $hash.Hash.ToLower()
Write-Host -ForegroundColor Green "$hashtext"
Write-Host -ForegroundColor Green "build tunnelssh success"
