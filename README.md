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
- 🚀 **One-Command Installation** - `./vertex --domain myapp.local` installs everything automatically
- 🌐 **Nginx Integration** - Optional nginx proxy for clean domain-based access
- 🔒 **HTTPS Support** - Automatic locally-trusted certificates with mkcert (.dev domains auto-enable HTTPS)
- 🔧 **Build Wrapper Management** - Generate and repair Maven/Gradle wrapper files

## 🛠️ Installation

### Prerequisites

- **Java 11+** (OpenJDK or Oracle)
- **Go 1.19+** (for building from source)
- **Node.js 16+** (for frontend development)
- **nginx** (optional - automatically installed when using `--nginx` flag)
- **mkcert** (optional - automatically installed when using `--https` flag)

### Quick Install

#### Option 1: Download Pre-built Binary

1. **Download the binary for your platform:**
   - Download from [GitHub Releases](https://github.com/zechtz/vertex/releases)
   - Choose the appropriate binary for your system:
     - `vertex-linux-amd64` (Linux 64-bit)
     - `vertex-darwin-amd64` (macOS Intel)
     - `vertex-darwin-arm64` (macOS Apple Silicon)
     - `vertex-windows-amd64.exe` (Windows 64-bit)

2. **Make it executable (Linux/macOS only):**
   ```bash
   chmod +x vertex-*
   ```

3. **Install as a user service:**
   ```bash
   # 🚀 ONE-COMMAND INSTALLATION (recommended)
   ./vertex-linux-amd64 --domain vertex.local    # Auto-installs with nginx!
   ./vertex-darwin-arm64 --domain myapp.local    # macOS example
   # vertex-windows-amd64.exe --domain myapp.local  (Windows example)
   
   # Traditional installation (still works)
   ./vertex-linux-amd64 --install      # Basic installation
   ./vertex-linux-amd64 --install --nginx --domain vertex.local  # Explicit
   ```

#### Option 2: Build from Source

1. **Build the application:**
   ```bash
   go build -o vertex
   ```

2. **Install as a user service:**
   ```bash
   # 🚀 ONE-COMMAND INSTALLATION (recommended)
   ./vertex --domain myapp.local        # Auto-installs with nginx!
   
   # Traditional installation (still works)
   ./vertex --install                   # Basic installation
   ./vertex --install --nginx --domain myapp.local  # Explicit
   ```

3. **Access the web interface:**
   - **With HTTPS domain**: https://vertex.dev (when using `--domain vertex.dev`)
   - **With HTTP domain**: http://myapp.local (when using `--domain myapp.local`)
   - **Direct access**: http://localhost:54321

> 📖 **For detailed usage instructions and tutorials, see our [Getting Started Guide](https://github.com/zechtz/vertex/wiki/Getting-Started-with-Vertex-Service-Manager)** on the wiki.

## 🚀 Usage

### 🌐 Nginx Proxy Configuration

Vertex includes optional nginx integration for clean domain-based access without port numbers.

#### Quick Setup
```bash
# 🚀 ONE-COMMAND INSTALLATION (recommended)
./vertex --domain vertex.dev

# Access via clean domain with HTTPS (auto-enabled for .dev domains)
open https://vertex.dev
```

#### Custom Domain
```bash
# One-command installation with custom domain (HTTP)
./vertex --domain myapp.local

# One-command installation with HTTPS
./vertex --domain myapp.local --https

# Access your custom domain
open https://myapp.local    # With HTTPS
open http://myapp.local     # HTTP only
```

#### Advanced Configuration
```bash
# Traditional explicit installation (all options available)
./vertex --install \
  --nginx \                    # Enable nginx proxy
  --https \                    # Enable HTTPS with locally-trusted certificates
  --domain myproject.local \   # Custom domain name
  --port 54321                 # Vertex service port (default: 54321)

# One-command with custom port
./vertex --domain myproject.local --port 8080

# Force HTTPS for any domain
./vertex --domain myproject.local --https
```

#### What Nginx Setup Does
- ✅ **Automatically installs nginx** on macOS (brew), Linux (apt/yum/etc), Windows (choco/winget)
- ✅ **Creates proxy configuration** from port 80/443 to Vertex service
- ✅ **Manages /etc/hosts entries** for local domain resolution
- ✅ **Handles permissions** and log directory creation
- ✅ **Starts nginx service** automatically
- 🔒 **HTTPS Support** - Automatically installs mkcert and generates locally-trusted certificates
- 🔒 **Auto-HTTPS for .dev domains** - Google-owned .dev domains automatically enable HTTPS (HSTS required)
- 🔐 **HTTP to HTTPS redirect** - Automatic redirects when HTTPS is enabled
- 🛡️ **Modern SSL configuration** - TLS 1.2+, HTTP/2, secure ciphers, security headers

#### Access Methods
| Method | URL | Use Case |
|--------|-----|----------|
| **Nginx Proxy (HTTPS)** | `https://vertex.dev` | Secure domain access (.dev domains auto-enable HTTPS) |
| **Nginx Proxy (HTTP)** | `http://myproject.local` | Clean domain access for non-.dev domains |
| **Direct Access** | `http://localhost:54321` | Development, bypassing nginx |

#### HTTPS with mkcert

Vertex uses [mkcert](https://mkcert.dev) to generate locally-trusted SSL certificates. This provides real HTTPS with valid certificates that browsers trust.

```bash
# Check certificate status
ls -la ~/.vertex/ssl/

# Manually generate certificates
mkcert -install                          # Install local CA
mkcert -cert-file ~/.vertex/ssl/mydomain.local.pem \
       -key-file ~/.vertex/ssl/mydomain.local-key.pem \
       mydomain.local

# View certificate details  
openssl x509 -in ~/.vertex/ssl/vertex.dev.pem -text -noout
```

**Special .dev Domain Handling:**
- Google owns the `.dev` TLD and requires HTTPS via HSTS preloading
- Vertex automatically enables HTTPS for any `.dev` domain
- Certificates are generated and installed automatically
- No browser security warnings with locally-trusted certificates

**Certificate Management:**
- Certificates stored in `~/.vertex/ssl/`
- Valid for the local CA installed by mkcert
- Automatically trusted by browsers and curl
- Use `mkcert -uninstall` to remove the local CA if needed

#### Troubleshooting Nginx
```bash
# Check nginx status
brew services list | grep nginx           # macOS
systemctl status nginx                   # Linux

# View nginx logs
tail -f /opt/homebrew/var/log/nginx/error.log    # macOS
tail -f /var/log/nginx/error.log                 # Linux

# Test configuration
nginx -t

# Restart nginx
brew services restart nginx              # macOS
sudo systemctl restart nginx            # Linux

# Check HTTPS certificate
curl -v https://vertex.dev
openssl s_client -connect vertex.dev:443 -servername vertex.dev
```

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

You can run Vertex on a different port (default is 54321):

#### Option 1: Direct execution
```bash
./vertex --port 9090
```

#### Option 2: Modify service configuration

**macOS:**
1. Stop the service: `launchctl stop com.vertex.manager`
2. Edit the plist file: `~/Library/LaunchAgents/com.vertex.manager.plist`
3. Change the port argument from `54321` to your desired port
4. Reload: `launchctl unload ~/Library/LaunchAgents/com.vertex.manager.plist && launchctl load ~/Library/LaunchAgents/com.vertex.manager.plist`

**Linux:**
1. Stop the service: `systemctl --user stop vertex`
2. Edit the service file: `~/.config/systemd/user/vertex.service`
3. Change the `--port 54321` argument to your desired port
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

### Command Line Options

Vertex supports these command line flags:

```bash
./vertex --help
```

| Flag | Default | Description |
|------|---------|-------------|
| `--domain` | vertex.dev | **🚀 Smart install**: Domain name for nginx proxy (auto-installs when specified) |
| `--install` | - | Install Vertex as a user service |
| `--uninstall` | - | Uninstall Vertex service and data |
| `--update` | - | Update the Vertex binary and restart the service |
| `--nginx` | false | Configure nginx proxy for domain access |
| `--https` | false | Enable HTTPS with locally-trusted certificates (auto-enabled for .dev domains) |
| `--port` | 54321 | Port to run the server on |
| `--data-dir` | ~/.vertex | Directory to store application data |
| `--version` | - | Show version information |

#### Examples
```bash
# 🚀 ONE-COMMAND INSTALLATION (recommended)
./vertex --domain myapp.local        # HTTP installation
./vertex --domain vertex.dev         # HTTPS auto-enabled for .dev domains

# Traditional installations (still supported)
./vertex --install                   # Basic installation
./vertex --install --nginx          # With nginx proxy
./vertex --install --nginx --https --domain myapp.local  # With HTTPS
./vertex --install --nginx --domain myapp.local --port 8080  # Full explicit

# Force HTTPS for any domain
./vertex --domain myproject.local --https

# Temporary run (no installation)
./vertex --port 9090

# Custom data directory
./vertex --data-dir /tmp/vertex-test --port 8080
```

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
   lsof -i :54321
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
curl http://localhost:54321/api/system/java-diagnostics
```

This will show:
- Detected Java installations
- PATH configuration
- Available vs working Java versions

## 🔄 Updating

Vertex includes a built-in updater to simplify the process of updating the binary and restarting the service. This is especially useful during development.

### Recommended Update Method

1. **Build the new binary:**
   ```bash
   go build -o vertex
   ```

2. **Run the updater:**
   ```bash
   ./vertex --update
   ```

This command will:
- Stop the running Vertex service.
- Replace the existing binary with the new one.
- Restart the service.

### Manual Update Method

To update Vertex manually:

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

```bash
# Self-uninstalling - works on all platforms!
./vertex --uninstall
```

Or manually:

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

**Windows:**
```bash
schtasks /delete /tn "VertexServiceManager" /f
rm ~/.local/bin/vertex.exe
rm ~/.local/bin/vertex-service.bat
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
./vertex --port 54321

# With custom data directory
VERTEX_DATA_DIR=/tmp/vertex-dev ./vertex --port 9090

# Run with nginx proxy in development
./vertex --install --nginx --domain dev.local
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