# Vertex Windows Installation Script
# Run as Administrator: powershell -ExecutionPolicy Bypass -File install.ps1

param(
    [string]$InstallPath = "$env:ProgramFiles\Vertex",
    [string]$DataPath = "$env:ProgramData\Vertex",
    [string]$ServiceName = "Vertex",
    [int]$Port = 8080
)

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "❌ This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator', then run:" -ForegroundColor Yellow
    Write-Host "powershell -ExecutionPolicy Bypass -File install.ps1" -ForegroundColor Yellow
    exit 1
}

Write-Host "🚀 Installing Vertex Service Manager for Windows..." -ForegroundColor Green

# Check if binary exists
if (-not (Test-Path ".\vertex.exe")) {
    Write-Host "❌ vertex.exe not found in current directory" -ForegroundColor Red
    Write-Host "Please build the application first with: go build -o vertex.exe" -ForegroundColor Yellow
    exit 1
}

try {
    # Create installation directory
    Write-Host "📁 Creating installation directory: $InstallPath" -ForegroundColor Blue
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null

    # Create data directory
    Write-Host "📁 Creating data directory: $DataPath" -ForegroundColor Blue
    New-Item -ItemType Directory -Path $DataPath -Force | Out-Null

    # Copy binary
    Write-Host "📦 Installing binary to $InstallPath" -ForegroundColor Blue
    Copy-Item ".\vertex.exe" -Destination "$InstallPath\vertex.exe" -Force

    # Create Windows service using NSSM (if available) or built-in sc command
    $servicePath = "$InstallPath\vertex.exe"
    $serviceArgs = "-port $Port"
    
    # Check if service already exists
    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Host "🔄 Stopping existing service..." -ForegroundColor Yellow
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Write-Host "🗑️ Removing existing service..." -ForegroundColor Yellow
        sc.exe delete $ServiceName | Out-Null
        Start-Sleep -Seconds 2
    }

    # Create new service
    Write-Host "🔧 Creating Windows service..." -ForegroundColor Blue
    sc.exe create $ServiceName binPath= "`"$servicePath`" $serviceArgs" start= auto DisplayName= "Vertex Service Manager" | Out-Null
    
    # Set service description
    sc.exe description $ServiceName "Vertex microservice management platform" | Out-Null
    
    # Set environment variable for the service
    $env:VERTEX_DATA_DIR = $DataPath
    [System.Environment]::SetEnvironmentVariable("VERTEX_DATA_DIR", $DataPath, [System.EnvironmentVariableTarget]::Machine)

    # Set service to restart on failure
    sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/30000 | Out-Null

    # Start the service
    Write-Host "▶️ Starting Vertex service..." -ForegroundColor Blue
    Start-Service -Name $ServiceName

    # Add to Windows Firewall (optional)
    Write-Host "🔥 Configuring Windows Firewall..." -ForegroundColor Blue
    try {
        New-NetFirewallRule -DisplayName "Vertex Service Manager" -Direction Inbound -Port $Port -Protocol TCP -Action Allow -ErrorAction SilentlyContinue | Out-Null
    } catch {
        Write-Host "⚠️ Could not configure firewall automatically. You may need to allow port $Port manually." -ForegroundColor Yellow
    }

    Write-Host ""
    Write-Host "✅ Installation completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Installation Details:" -ForegroundColor Cyan
    Write-Host "   • Service Name: $ServiceName"
    Write-Host "   • Binary Path: $InstallPath\vertex.exe"
    Write-Host "   • Data Directory: $DataPath"
    Write-Host "   • Database: $DataPath\vertex.db"
    Write-Host "   • Port: $Port"
    Write-Host ""
    Write-Host "📋 Service Management:" -ForegroundColor Cyan
    Write-Host "   • Start: Start-Service -Name '$ServiceName'"
    Write-Host "   • Stop: Stop-Service -Name '$ServiceName'"
    Write-Host "   • Status: Get-Service -Name '$ServiceName'"
    Write-Host "   • Logs: Get-EventLog -LogName Application -Source '$ServiceName'"
    Write-Host ""
    Write-Host "🌐 Access the web interface at: http://localhost:$Port" -ForegroundColor Green
    Write-Host ""
    Write-Host "🔧 Service Management UI:" -ForegroundColor Cyan
    Write-Host "   • Services: services.msc"
    Write-Host "   • Event Viewer: eventvwr.msc"

} catch {
    Write-Host "❌ Installation failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}