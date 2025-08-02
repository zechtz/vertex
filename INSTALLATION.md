# Vertex Installation Guide

## Platform-Specific Quick Installation

### Windows

#### Option 1: Automatic Installation (Recommended)

1. **Build the application:**
   ```cmd
   go build -o vertex.exe
   ```

2. **Install as Windows service (Run as Administrator):**
   ```powershell
   powershell -ExecutionPolicy Bypass -File install.ps1
   ```

3. **Manage the service:**
   ```powershell
   # Start service
   Start-Service -Name "Vertex"
   
   # Check status
   Get-Service -Name "Vertex"
   
   # Stop service
   Stop-Service -Name "Vertex"
   ```

4. **Access the web interface:**
   Open http://localhost:8080 in your browser

#### Option 2: Manual Windows Installation

1. **Build and install:**
   ```cmd
   go build -o vertex.exe
   mkdir "C:\Program Files\Vertex"
   copy vertex.exe "C:\Program Files\Vertex\"
   ```

2. **Set data directory:**
   ```cmd
   mkdir "C:\ProgramData\Vertex"
   setx VERTEX_DATA_DIR "C:\ProgramData\Vertex" /M
   ```

3. **Run manually:**
   ```cmd
   "C:\Program Files\Vertex\vertex.exe" -port 8080
   ```

### Linux/Unix

#### Option 1: System-wide Installation (Recommended)

1. **Build the application:**
   ```bash
   go build -o vertex
   ```

2. **Install as a system service:**
   ```bash
   sudo ./install.sh
   ```

3. **Start the service:**
   ```bash
   sudo systemctl start vertex
   sudo systemctl status vertex
   ```

4. **Access the web interface:**
   Open http://localhost:8080 in your browser

#### Option 2: Manual Linux Installation

1. **Build and copy binary:**
   ```bash
   go build -o vertex
   sudo cp vertex /usr/local/bin/
   ```

2. **Create data directory:**
   ```bash
   sudo mkdir -p /var/lib/vertex
   sudo chown $USER:$USER /var/lib/vertex
   ```

3. **Run with custom data directory:**
   ```bash
   vertex -data-dir /var/lib/vertex
   ```

## Cross-Platform Development/Testing

For development or testing, you can run directly:

```bash
go run main.go
```

Or with a custom data directory:

```bash
go run main.go -data-dir ./data
```

## Configuration Options

### Command Line Flags

- `-port`: Server port (default: 8080)
- `-data-dir`: Data directory path (overrides VERTEX_DATA_DIR)
- `-version`: Show version information

### Environment Variables

- `VERTEX_DATA_DIR`: Directory for database and application data

### Examples

```bash
# Run on custom port with specific data directory
vertex -port 9090 -data-dir /home/user/vertex-data

# Use environment variable
export VERTEX_DATA_DIR=/var/lib/vertex
vertex -port 8080

# System service (using install.sh)
sudo systemctl start vertex
```

## Default Data Directory Locations

When no custom path is specified, Vertex automatically chooses platform-appropriate locations:

| Platform | Default Data Directory | Example Database Path |
|----------|------------------------|----------------------|
| **Windows** | `%APPDATA%\Vertex` | `C:\Users\Username\AppData\Roaming\Vertex\vertex.db` |
| **macOS** | `~/Library/Application Support/Vertex` | `/Users/Username/Library/Application Support/Vertex/vertex.db` |
| **Linux** | `~/.local/share/vertex` | `/home/username/.local/share/vertex/vertex.db` |

## Installation File Locations

### Windows
| Installation Type | Database Location | Binary Location |
|-------------------|-------------------|-----------------|
| Windows Service | `C:\ProgramData\Vertex\vertex.db` | `C:\Program Files\Vertex\vertex.exe` |
| Manual | `%VERTEX_DATA_DIR%\vertex.db` | User-defined |
| Development | `%APPDATA%\Vertex\vertex.db` | Current directory |

### Linux/Unix
| Installation Type | Database Location | Binary Location |
|-------------------|-------------------|-----------------|
| System Service | `/var/lib/vertex/vertex.db` | `/usr/local/bin/vertex` |
| Manual | `$VERTEX_DATA_DIR/vertex.db` | User-defined |
| Development | `~/.local/share/vertex/vertex.db` | Current directory |

## Service Management

### Windows Service Management (after install.ps1)

```powershell
# Start service
Start-Service -Name "Vertex"

# Stop service
Stop-Service -Name "Vertex"

# Restart service
Restart-Service -Name "Vertex"

# Check status
Get-Service -Name "Vertex"

# View service configuration
Get-WmiObject Win32_Service -Filter "Name='Vertex'"

# View logs in Event Viewer
Get-EventLog -LogName Application -Source "Vertex" -Newest 20

# Uninstall service and application
powershell -ExecutionPolicy Bypass -File uninstall.ps1
```

### Linux systemd Management (after install.sh)

```bash
# Start service
sudo systemctl start vertex

# Stop service
sudo systemctl stop vertex

# Restart service
sudo systemctl restart vertex

# Enable on boot
sudo systemctl enable vertex

# View logs
sudo journalctl -u vertex -f

# Check status
sudo systemctl status vertex
```

### Manual Process Management

```bash
# Start in background
nohup vertex -data-dir /var/lib/vertex > /var/log/vertex.log 2>&1 &

# Find and stop process
pkill -f vertex
```

## Troubleshooting

### Windows-Specific Issues

#### Permission Issues
```powershell
# Fix data directory permissions
icacls "C:\ProgramData\Vertex" /grant Users:F /T

# Run PowerShell as Administrator
# Right-click PowerShell -> "Run as Administrator"

# Check if service is running
Get-Service -Name "Vertex"

# Restart service if stuck
Stop-Service -Name "Vertex" -Force
Start-Service -Name "Vertex"
```

#### Windows Firewall
```powershell
# Manually add firewall rule
New-NetFirewallRule -DisplayName "Vertex Service" -Direction Inbound -Port 8080 -Protocol TCP -Action Allow

# Or use Windows Firewall GUI:
# Windows Security -> Firewall & network protection -> Allow an app through firewall
```

#### Service Installation Issues
```powershell
# If install.ps1 fails, try manual service creation:
sc.exe create Vertex binPath= "C:\Program Files\Vertex\vertex.exe -port 8080" start= auto

# Check Windows Event Log for errors:
Get-EventLog -LogName System -Source "Service Control Manager" | Where-Object {$_.Message -like "*Vertex*"}
```

### Linux Permission Issues

If you encounter permission issues:

```bash
# Fix data directory permissions
sudo chown -R vertex:vertex /var/lib/vertex
sudo chmod -R 755 /var/lib/vertex

# Fix binary permissions
sudo chmod 755 /usr/local/bin/vertex
```

### Database Issues

```bash
# Check database location
ls -la $VERTEX_DATA_DIR/

# Reset database (WARNING: This deletes all data)
sudo systemctl stop vertex
sudo rm /var/lib/vertex/vertex.db
sudo systemctl start vertex
```

### Port Conflicts

```bash
# Check what's using port 8080
sudo lsof -i :8080

# Use different port
vertex -port 9090
```