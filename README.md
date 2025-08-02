# Vertex Service Manager

A powerful service management platform that provides a web-based interface for managing multiple services across different profiles. Vertex automatically handles Java environment detection, Maven/Gradle builds, service dependencies, and provides real-time monitoring.

## ✨ Features

- 🚀 **Multi-Profile Management** - Organize services into different profiles (dev, staging, production)
- ☕ **Automatic Java Detection** - Works with ASDF, SDKMAN, Homebrew, and system Java installations
- 🔧 **Build System Support** - Automatic detection and support for Maven and Gradle projects
- 📊 **Real-time Monitoring** - Live logs, health checks, and resource metrics
- 🌐 **Web Interface** - Modern React-based dashboard for service management
- 🔒 **User Authentication** - Secure JWT-based authentication system
- 📱 **Responsive Design** - Works on desktop and mobile devices

## 🛠️ Installation

### Prerequisites

- **Java 11+** (OpenJDK or Oracle)
- **Go 1.19+** (for building from source)
- **Node.js 16+** (for frontend development)

### Quick Install

1. **Build the application:**
   ```bash
   go build -o vertex
   ```

2. **Install as a user service:**
   ```bash
   # No sudo required - installs as current user
   ./install.sh
   ```

3. **Access the web interface:**
   Open your browser and navigate to: http://localhost:8080

> 📖 **For detailed usage instructions and tutorials, see our [Getting Started Guide](https://github.com/zechtz/vertex/wiki/Getting-Started-with-Vertex-Service-Manager)** on the wiki.

## 🚀 Usage

### Service Management

The service **starts automatically** after installation using:
- **macOS**: LaunchAgent (user-level service)
- **Linux**: systemd user service

#### Start/Stop Service

**macOS:**
```bash
# Start
launchctl start com.vertex.manager

# Stop
launchctl stop com.vertex.manager

# Restart
launchctl stop com.vertex.manager && launchctl start com.vertex.manager
```

**Linux:**
```bash
# Start
systemctl --user start vertex

# Stop
systemctl --user stop vertex

# Restart
systemctl --user restart vertex

# Check status
systemctl --user status vertex
```

### Custom Port Configuration

You can run Vertex on a different port:

#### Option 1: Direct execution
```bash
./vertex --port 9090
```

#### Option 2: Modify service configuration

**macOS:**
1. Stop the service: `launchctl stop com.vertex.manager`
2. Edit the plist file: `~/Library/LaunchAgents/com.vertex.manager.plist`
3. Change the port argument from `8080` to your desired port
4. Reload: `launchctl unload ~/Library/LaunchAgents/com.vertex.manager.plist && launchctl load ~/Library/LaunchAgents/com.vertex.manager.plist`

**Linux:**
1. Stop the service: `systemctl --user stop vertex`
2. Edit the service file: `~/.config/systemd/user/vertex.service`
3. Change the `-port 8080` argument to your desired port
4. Reload: `systemctl --user daemon-reload && systemctl --user start vertex`

### Viewing Logs

#### Real-time Log Monitoring

**macOS:**
```bash
# Main application logs
tail -f ~/.vertex/vertex.stderr.log

# Startup logs
tail -f ~/.vertex/vertex.stdout.log
```

**Linux:**
```bash
# All logs
journalctl --user -u vertex -f

# Recent logs
journalctl --user -u vertex --since="1 hour ago"
```

#### Log Locations

| Platform | Location |
|----------|----------|
| **macOS** | `~/.vertex/vertex.stderr.log`<br>`~/.vertex/vertex.stdout.log` |
| **Linux** | `journalctl --user -u vertex` |
| **Database** | `~/.vertex/vertex.db` |
| **Config** | `~/.vertex/` |

## 📂 Directory Structure

```
~/.vertex/                     # User data directory
├── vertex.db                  # SQLite database
├── vertex.stderr.log          # Application logs (macOS)
├── vertex.stdout.log          # Startup logs (macOS)
└── env_vars.fish             # Environment variables (optional)

~/.local/bin/vertex            # Binary location (user installation)
```

## 🔧 Configuration

### Environment Variables

Vertex supports these environment variables:

- `VERTEX_DATA_DIR` - Override data directory (default: `~/.vertex`)
- `JWT_SECRET` - Custom JWT secret for authentication
- `JAVA_HOME` - Override Java installation path

### Profile Management

1. **Create a Profile** - Navigate to the Profiles section in the web interface
2. **Add Services** - Define your services with their directories and configurations
3. **Set Projects Directory** - Each profile can have its own root directory for services
4. **Start Profile** - Use the profile management interface to start all services in a profile

## ☕ Java Environment

Vertex automatically detects Java installations in this order:

1. **JAVA_HOME** environment variable
2. **Java in PATH** (with validation)
3. **User-specific installations:**
   - ASDF: `~/.asdf/installs/java/`
   - SDKMAN: `~/.sdkman/candidates/java/`
4. **System installations:**
   - macOS: Homebrew, system locations
   - Linux: OpenJDK packages
   - Windows: Program Files

### Supported Java Managers

- ✅ **ASDF** - `asdf install java openjdk-17`
- ✅ **SDKMAN** - `sdk install java 17.0.1-open`
- ✅ **Homebrew** - `brew install openjdk`
- ✅ **System packages** - `apt install openjdk-17-jdk`

## 🐛 Troubleshooting

### Service Won't Start

1. **Check logs:**
   ```bash
   # macOS
   tail -n 100 ~/.vertex/vertex.stderr.log
   
   # Linux  
   journalctl --user -u vertex --lines=100
   ```

2. **Verify Java installation:**
   ```bash
   java -version
   echo $JAVA_HOME
   ```

3. **Check port availability:**
   ```bash
   lsof -i :8080
   ```

### Permission Issues

Since Vertex runs as your user account, it should have access to all your project files. If you encounter permission issues:

1. **Verify directory ownership:**
   ```bash
   ls -la /path/to/your/projects
   ```

2. **Check build directory permissions:**
   ```bash
   ls -la /path/to/project/target  # Maven
   ls -la /path/to/project/build   # Gradle
   ```

### Java Detection Issues

Run the built-in diagnostics:
```bash
curl http://localhost:8080/api/system/java-diagnostics
```

This will show:
- Detected Java installations
- PATH configuration
- Available vs working Java versions

## 🔄 Updating

To update Vertex:

1. **Stop the service:**
   ```bash
   # macOS
   launchctl stop com.vertex.manager
   
   # Linux
   systemctl --user stop vertex
   ```

2. **Build new version:**
   ```bash
   git pull
   go build -o vertex
   ```

3. **Reinstall:**
   ```bash
   ./install.sh
   ```

## 🗑️ Uninstalling

To completely remove Vertex:

**macOS:**
```bash
launchctl stop com.vertex.manager
launchctl unload ~/Library/LaunchAgents/com.vertex.manager.plist
rm ~/Library/LaunchAgents/com.vertex.manager.plist
rm ~/.local/bin/vertex
rm -rf ~/.vertex
```

**Linux:**
```bash
systemctl --user stop vertex
systemctl --user disable vertex
rm ~/.config/systemd/user/vertex.service
systemctl --user daemon-reload
rm ~/.local/bin/vertex
rm -rf ~/.vertex
```

## 📝 Development

### Building from Source

```bash
# Backend
go build -o vertex

# Frontend (if modified)
cd frontend
npm install
npm run build
```

### Running in Development Mode

```bash
# Run without installing
./vertex --port 8080

# With custom data directory
VERTEX_DATA_DIR=/tmp/vertex-dev ./vertex --port 9090
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section above
- Review the logs for error messages

---

**Happy service managing! 🚀**