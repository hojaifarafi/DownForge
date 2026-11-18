<#
.SYNOPSIS
    Installs DownForge as a Windows Service using NSSM (Non-Sucking Service Manager) or sc.exe
.DESCRIPTION
    Run this script in PowerShell as Administrator:
    powershell -ExecutionPolicy Bypass -File .\scripts\install-service-windows.ps1
#>

$ServiceName = "DownForge"
$ProjectDir = (Get-Item $PSScriptRoot).Parent.FullName
$ExePath = Join-Path $ProjectDir "downforge.exe"

Write-Host "🔨 Building DownForge for Windows..." -ForegroundColor Cyan
Set-Location $ProjectDir
go build -o $ExePath .\cmd\downforge

if (-not (Test-Path $ExePath)) {
    Write-Host "❌ Failed to compile downforge.exe" -ForegroundColor Red
    Exit 1
}

Write-Host "⚙️ Configuring Windows Service..." -ForegroundColor Cyan

# Check if nssm is available
if (Get-Command nssm -ErrorAction SilentlyContinue) {
    nssm stop $ServiceName 2>$null
    nssm remove $ServiceName confirm 2>$null
    nssm install $ServiceName "$ExePath" "-port" "8080"
    nssm set $ServiceName AppDirectory "$ProjectDir"
    nssm set $ServiceName Description "DownForge Background Download Manager"
    nssm start $ServiceName
} else {
    # Fallback to standard sc.exe / New-Service
    if (Get-Service $ServiceName -ErrorAction SilentlyContinue) {
        Stop-Service $ServiceName -ErrorAction SilentlyContinue
        sc.exe delete $ServiceName
    }
    New-Service -Name $ServiceName -BinaryPathName "`"$ExePath`" -port 8080" -DisplayName "DownForge Service" -StartupType Automatic
    Start-Service $ServiceName
}

Write-Host "✅ DownForge service installed successfully!" -ForegroundColor Green
Write-Host "   Access web UI: http://localhost:8080" -ForegroundColor Cyan
