# Vertex Installation Guide

## Quick Installation

### Option 1: System-wide Installation (Recommended)

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

### Option 2: Manual Installation

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

### Option 3: Development/Testing

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

## File Locations

| Installation Type | Database Location | Binary Location |
|-------------------|-------------------|-----------------|
| System Service | `/var/lib/vertex/vertex.db` | `/usr/local/bin/vertex` |
| Manual | `$VERTEX_DATA_DIR/vertex.db` | User-defined |
| Development | `./vertex.db` | Current directory |

## Service Management

### Using systemd (after install.sh)

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

### Permission Issues

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