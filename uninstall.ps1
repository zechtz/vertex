# Vertex Windows Uninstallation Script
# Run as Administrator: powershell -ExecutionPolicy Bypass -File uninstall.ps1

param(
    [string]$InstallPath = "$env:ProgramFiles\Vertex",
    [string]$DataPath = "$env:ProgramData\Vertex",
    [string]$ServiceName = "Vertex",
    [switch]$KeepData = $false
)

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "❌ This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator', then run:" -ForegroundColor Yellow
    Write-Host "powershell -ExecutionPolicy Bypass -File uninstall.ps1" -ForegroundColor Yellow
    exit 1
}

Write-Host "🗑️ Uninstalling Vertex Service Manager..." -ForegroundColor Yellow

try {
    # Stop and remove service
    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Host "⏹️ Stopping Vertex service..." -ForegroundColor Blue
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        
        Write-Host "🗑️ Removing Vertex service..." -ForegroundColor Blue
        sc.exe delete $ServiceName | Out-Null
        Start-Sleep -Seconds 2
    } else {
        Write-Host "ℹ️ Service not found, skipping service removal" -ForegroundColor Gray
    }

    # Remove firewall rule
    Write-Host "🔥 Removing firewall rule..." -ForegroundColor Blue
    try {
        Remove-NetFirewallRule -DisplayName "Vertex Service Manager" -ErrorAction SilentlyContinue | Out-Null
    } catch {
        Write-Host "ℹ️ Firewall rule not found or could not be removed" -ForegroundColor Gray
    }

    # Remove environment variable
    [System.Environment]::SetEnvironmentVariable("VERTEX_DATA_DIR", $null, [System.EnvironmentVariableTarget]::Machine)

    # Remove installation directory
    if (Test-Path $InstallPath) {
        Write-Host "📁 Removing installation directory: $InstallPath" -ForegroundColor Blue
        Remove-Item -Path $InstallPath -Recurse -Force
    }

    # Remove data directory (unless keeping data)
    if (Test-Path $DataPath) {
        if (-not $KeepData) {
            $confirmation = Read-Host "⚠️ Do you want to delete all application data? (y/N)"
            if ($confirmation -eq 'y' -or $confirmation -eq 'Y') {
                Write-Host "📁 Removing data directory: $DataPath" -ForegroundColor Blue
                Remove-Item -Path $DataPath -Recurse -Force
            } else {
                Write-Host "📁 Keeping data directory: $DataPath" -ForegroundColor Green
            }
        } else {
            Write-Host "📁 Keeping data directory: $DataPath" -ForegroundColor Green
        }
    }

    Write-Host ""
    Write-Host "✅ Uninstallation completed successfully!" -ForegroundColor Green
    
    if (Test-Path $DataPath) {
        Write-Host ""
        Write-Host "📁 Data directory preserved at: $DataPath" -ForegroundColor Cyan
        Write-Host "   Contains database and configuration files"
        Write-Host "   Delete manually if no longer needed"
    }

} catch {
    Write-Host "❌ Uninstallation failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}