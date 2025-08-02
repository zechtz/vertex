# Vertex

A comprehensive Spring Boot microservice management platform providing automated service startup, environment management, monitoring, and build compatibility fixes with cross-platform support.

## 🚀 Quick Installation

### Windows

1. **Build the application:**
   ```cmd
   go build -o vertex.exe
   ```

2. **Install as Windows service (Run as Administrator):**
   ```powershell
   powershell -ExecutionPolicy Bypass -File install.ps1
   ```

3. **Access the web interface:** http://localhost:8080

### Linux/macOS

1. **Build the application:**
   ```bash
   go build -o vertex
   ```

2. **Install as system service:**
   ```bash
   sudo ./install.sh
   sudo systemctl start vertex
   ```

3. **Access the web interface:** http://localhost:8080

### Cross-Platform Development

For development or testing on any platform:

```bash
# Build and run directly
go run main.go

# Or with custom settings
go run main.go -port 9090 -data-dir ./data
```

## 📦 Platform-Specific Data Locations

Vertex automatically stores data in platform-appropriate locations:

| Platform | Default Location | Example Path |
|----------|------------------|--------------|
| **Windows** | `%APPDATA%\Vertex` | `C:\Users\Username\AppData\Roaming\Vertex\vertex.db` |
| **macOS** | `~/Library/Application Support/Vertex` | `/Users/Username/Library/Application Support/Vertex/vertex.db` |
| **Linux** | `~/.local/share/vertex` | `/home/username/.local/share/vertex/vertex.db` |

### Custom Data Directory

Override the default location:

```bash
# Command line flag
vertex -data-dir /custom/path

# Environment variable  
export VERTEX_DATA_DIR=/custom/path
vertex
```

> 📖 **For detailed installation instructions, troubleshooting, and platform-specific guides, see [INSTALLATION.md](INSTALLATION.md)**

## 📋 Features

- **Service Management**: Start, stop, and restart services individually or in bulk
- **Real-time Monitoring**: Live status updates and health checks
- **Log Management**: View, filter, and export service logs
- **Environment Configuration**: Manage global and service-specific environment variables
- **Configuration Profiles**: Save and apply different service configurations
- **Library Installation**: Environment-aware Maven library installation with UI preview
- **Maven Wrapper Management**: Automatic creation/updating of Maven wrappers during service startup
- **Automated Build Fixes**: Automatic Lombok compatibility checking and fixing
- **Cross-Platform Support**: Native Windows, macOS, and Linux installation with proper data directories
- **Dark Mode**: Full dark theme support across all UI components
- **File Management**: Edit service configuration files directly from the web interface

## 📚 Environment-Aware Library Installation

### Smart GitLab CI Library Management

Vertex includes an intelligent library installation system that parses `.gitlab-ci.yml` files and provides environment-specific Maven library installation:

#### **Key Features**

- ✅ **GitLab CI Parsing**: Automatically detects Maven install commands in CI files
- ✅ **Environment Detection**: Groups libraries by environment (dev, staging, production, etc.)
- ✅ **Interactive UI**: Preview libraries before installation with checkbox selection
- ✅ **Confirmation Dialog**: Review installation details before proceeding
- ✅ **Progress Tracking**: Real-time installation progress with detailed logging
- ✅ **Error Handling**: Graceful error reporting and recovery
- ✅ **Dark Mode Support**: Consistent theming across all modal views

#### **How It Works**

1. **Library Detection**: Click "Install Libraries" on any service card
2. **Environment Preview**: View libraries grouped by environment with metadata
3. **Selective Installation**: Choose which environments to install libraries for
4. **Confirmation**: Review installation summary before proceeding
5. **Progress Monitoring**: Watch real-time installation with Maven output

#### **Supported CI Patterns**

The system recognizes common GitLab CI job patterns:
- `maven-build-dev`, `maven-build-staging`, `maven-build-prod`
- Jobs ending with `-dev`, `-staging`, `-live`, `-prod`, `-training`
- Custom environment detection from job names and branch configurations

### Maven Wrapper Auto-Management

- ✅ **Automatic Creation**: Creates Maven wrappers (`mvnw`) for services that don't have them
- ✅ **Version Management**: Uses Maven 3.9.6 for new wrapper generation
- ✅ **Smart Updates**: Only updates wrappers older than 30 days
- ✅ **Startup Integration**: Wrapper creation happens during service startup
- ✅ **Non-blocking**: Service startup continues even if wrapper creation fails

## 🔧 Maven & Lombok Compatibility

### Automatic Lombok Fix System

The service manager includes an **automated Lombok compatibility checker** that resolves common Maven compilation issues:

#### **Automatic Detection & Fixing**

- ✅ **On Service Startup**: Automatically checks and fixes Lombok configuration before starting services
- ✅ **Manual Fix**: Use the "Fix Lombok" button in the web interface to fix all services at once
- ✅ **Smart Detection**: Only processes services that actually use Lombok
- ✅ **Safe Backups**: Creates backups before making changes, with automatic restoration on failure

#### **What It Fixes**

The system automatically adds the required Maven compiler plugin configuration for Lombok:

```xml
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-compiler-plugin</artifactId>
    <version>3.11.0</version>
    <configuration>
        <source>17</source>
        <target>17</target>
        <annotationProcessorPaths>
            <path>
                <groupId>org.projectlombok</groupId>
                <artifactId>lombok</artifactId>
                <version>1.18.30</version>
            </path>
        </annotationProcessorPaths>
    </configuration>
</plugin>
```

### **Compatible Maven & Lombok Versions**

For optimal compatibility, ensure your Java projects use these versions:

| Component                 | Recommended Version | Notes                                   |
| ------------------------- | ------------------- | --------------------------------------- |
| **Maven**                 | 3.6.3 (via wrapper) | Uses project's `./mvnw` automatically   |
| **Lombok**                | 1.18.30             | Compatible with Java 17 and Maven 3.6.3 |
| **Java**                  | 17                  | Required for Spring Boot 2.7.x          |
| **Maven Compiler Plugin** | 3.11.0              | Supports annotation processing          |

## 🌍 **Automatic Environment Management**

### **Built-in Environment Setup**

The service manager includes a comprehensive environment management system that eliminates the need for manual setup scripts:

#### **Automatic Detection & Setup**

- ✅ **Startup Detection**: Automatically checks environment variables when the application starts
- ✅ **Smart Sync**: Loads from existing `common_env_settings.sh` and `env_vars.fish` files if available
- ✅ **Manual Sync**: "Sync Environment" button in web interface for manual synchronization

````

#### **Binary Distribution Benefits**

- **📦 Self-contained**: Single binary (~13MB) with embedded React web interface and SQLite database
- **🚀 Zero configuration**: No setup scripts, config files, or external dependencies needed
- **🌍 Cross-platform**: Works on Linux, macOS, and Windows
- **🔄 Backward compatible**: Generates traditional config files for existing workflows
- **📱 Portable**: Copy and run anywhere - perfect for development teams
- **🗄️ Built-in database**: SQLite database included for environment persistence
- **⚡ No external web server**: Frontend assets embedded directly in the binary

### **Manual pom.xml Configuration**

If you need to manually configure a new Java service, ensure your `pom.xml` includes:

```xml
<properties>
    <java.version>17</java.version>
</properties>

<dependencies>
    <dependency>
        <groupId>org.projectlombok</groupId>
        <artifactId>lombok</artifactId>
        <version>1.18.30</version>
        <scope>provided</scope>
    </dependency>
</dependencies>

<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-compiler-plugin</artifactId>
            <version>3.11.0</version>
            <configuration>
                <source>17</source>
                <target>17</target>
                <annotationProcessorPaths>
                    <path>
                        <groupId>org.projectlombok</groupId>
                        <artifactId>lombok</artifactId>
                        <version>1.18.30</version>
                    </path>
                </annotationProcessorPaths>
            </configuration>
        </plugin>
    </plugins>
</build>
````

### **Troubleshooting Build Issues**

If you encounter compilation errors like:

```
cannot find symbol: method setId(long)
cannot find symbol: method setUuid(java.lang.String)
```

**Solution 1 - Automatic Fix:**

1. Click the "Fix Lombok" button in the web interface
2. Or simply start the affected service (it will auto-fix)

**Solution 2 - Manual Fix:**

1. Add the Maven compiler plugin configuration shown above
2. Ensure Lombok version matches your dependency
3. Clean and rebuild: `./mvnw clean compile`

## 📁 Project Structure

```
vertex/
├── internal/
│   ├── config/         # Configuration management
│   ├── database/       # SQLite database operations
│   ├── handlers/       # HTTP request handlers
│   ├── models/         # Data models
│   └── services/       # Core service management logic
│       ├── lombok.go   # Lombok compatibility checker
│       ├── manager.go  # Service manager
│       └── operations.go # Service operations
├── web/                # React frontend
│   ├── src/
│   │   ├── components/ # UI components
│   │   └── types.ts    # TypeScript definitions
│   ├── dist/           # Built frontend assets
│   └── embed.go        # Embeds dist/* into Go binary
├── README.md           # This file
├── ENVIRONMENT_SETUP.md # Environment configuration guide
└── vertex             # Main executable
```

## 🚀 Automated Releases

The project includes CI/CD pipelines for both GitHub Actions and GitLab CI that automatically build and release binaries when you push a version tag.

### Creating a Release

1. **Tag a version**:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Automated builds** will create binaries for:
   - **Linux**: x64, ARM64
   - **macOS**: x64, ARM64 (Intel + Apple Silicon)
   - **Windows**: x64
   - **Docker**: Multi-platform container images

3. **GitHub/GitLab releases** are created automatically with:
   - Pre-built binaries for all platforms
   - SHA256 checksums for verification
   - Release notes with download instructions
   - Docker images (optional)

### CI/CD Features

- ✅ **Automated testing** on all platforms
- ✅ **Frontend build** with npm/Node.js
- ✅ **Cross-compilation** with CGO support
- ✅ **Release creation** with assets and notes
- ✅ **Docker images** pushed to registries
- ✅ **Checksum generation** for security
- ✅ **Version injection** from git tags

## 🔨 Development

### Prerequisites

- Go 1.19+
- Node.js 16+
- PostgreSQL
- RabbitMQ
- Redis

### Building from Source

#### Quick Build

```bash
# Use the automated build script (recommended)
./build.sh

# Or build manually:
cd web && npm install && npm run build && cd ..
go build -o vertex
```

#### Cross-Platform Build

```bash
# Build for all platforms
./build.sh

# Generated files:
# - vertex (current platform)
# - vertex-windows-amd64.exe
# - vertex-linux-amd64  
# - vertex-darwin-amd64
# - vertex-darwin-arm64
```

### Running in Development Mode

#### Option 1: Separate Frontend/Backend (Hot Reload)

```bash
# Start backend
go run main.go

# Start frontend development server (separate terminal)
cd web
npm run dev
# Frontend will be available at http://localhost:5173
# Backend API at http://localhost:8080
```

#### Option 2: Full Production Mode (Embedded UI)

```bash
# Build frontend first
cd web && npm run build && cd ..

# Start with embedded UI
go run main.go
# Full app available at http://localhost:8080
```

### Frontend Embedding Architecture

The application uses Go's `embed` package to include the React frontend directly in the binary:

```go
// web/embed.go
package web

import "embed"

//go:embed dist/*
var EmbeddedUI embed.FS
```

This approach provides:

- **Single Binary Distribution**: No need to ship frontend files separately
- **Simplified Deployment**: Just copy and run the binary
- **Version Consistency**: Frontend and backend are always in sync
- **No Web Server Required**: Go serves static files directly from memory

### Building for Distribution

#### Quick Build (Recommended)

```bash
# Use the build script (includes version info)
./build.sh

# Or build manually
cd web && npm run build && cd ..
CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex
```

#### Release Build (All Platforms)

```bash
# Build for all supported platforms
./release.sh v1.0.0

# This creates:
# - vertex-linux-amd64
# - vertex-linux-arm64
# - vertex-darwin-amd64
# - vertex-darwin-arm64
# - vertex-windows-amd64.exe
# - checksums.txt
```

> **Note**: The frontend must be built before the Go binary, as the `web/embed.go` file uses `//go:embed dist/*` to embed the React build artifacts directly into the binary.

#### Cross-Platform Builds

```bash
# Linux 64-bit
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex-linux-amd64

# macOS 64-bit
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex-darwin-amd64

# Windows 64-bit
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o vertex-windows-amd64.exe
```

> **Note**: CGO is required for SQLite support. For cross-compilation, you may need platform-specific CGO toolchains.

## 🔧 Configuration

### Environment Variables

All services share common environment variables defined in `common_env_settings.sh`:

- **Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`
- **RabbitMQ**: `RABBIT_HOSTNAME`, `RABBIT_PORT`, `RABBIT_USERNAME`, `RABBIT_PASSWORD`
- **Redis**: `REDIS_HOST`, `REDIS_USER`, `REDIS_PASS`
- **Service Discovery**: `DISCOVERY_SERVER`, `DEFAULT_ZONE`
- **Configuration Server**: `CONFIG_SERVER`, `CONFIG_USERNAME`, `CONFIG_PASSWORD`
- **Profile**: `ACTIVE_PROFILE` (dev/prod/test)

### Global Configuration

Use the web interface to manage:

- **Projects Directory**: Location of your Java services
- **Java Home Override**: Custom Java installation path
- **Environment Variables**: Global and service-specific variables

## 🔧 Service Management

### Windows Service Commands

```powershell
# Start service
Start-Service -Name "Vertex"

# Stop service  
Stop-Service -Name "Vertex"

# Check status
Get-Service -Name "Vertex"

# View logs
Get-EventLog -LogName Application -Source "Vertex" -Newest 20

# Uninstall
powershell -ExecutionPolicy Bypass -File uninstall.ps1
```

### Linux/macOS Service Commands

```bash
# Start service
sudo systemctl start vertex

# Stop service
sudo systemctl stop vertex

# Check status
sudo systemctl status vertex

# View logs
sudo journalctl -u vertex -f

# Enable on boot
sudo systemctl enable vertex
```

## 📊 API Endpoints

### Service Management

- `GET /api/services` - List all services
- `POST /api/services/{id}/start` - Start a service by UUID
- `POST /api/services/{id}/stop` - Stop a service by UUID
- `POST /api/services/{id}/restart` - Restart a service by UUID
- `POST /api/services/start-all` - Start all services
- `POST /api/services/stop-all` - Stop all services

### Library Installation

- `GET /api/services/{id}/libraries/preview` - Preview libraries from GitLab CI
- `POST /api/services/{id}/libraries/install` - Install selected libraries by environment
- `POST /api/services/{id}/install-libraries` - Install all libraries (legacy)

### Lombok Compatibility

- `POST /api/services/fix-lombok` - Check and fix Lombok compatibility for all services

### Environment Management

- `POST /api/environment/setup` - Setup default environment variables
- `POST /api/environment/sync` - Sync environment from existing configuration files
- `GET /api/environment/status` - Get current environment status and variables

### Configuration

- `GET /api/configurations` - List configurations
- `POST /api/configurations` - Save configuration
- `POST /api/configurations/{id}/apply` - Apply configuration

### Environment Variables

- `GET /api/env-vars/global` - Get global environment variables
- `PUT /api/env-vars/global` - Update global environment variables
- `GET /api/services/{name}/env-vars` - Get service-specific variables

## 🐛 Troubleshooting

### Service Won't Start

1. Check service logs in the web interface
2. Verify environment variables are set correctly
3. Ensure databases exist and are accessible
4. Try the "Fix Lombok" button if you see compilation errors
5. Check if Maven wrapper was created successfully

### Library Installation Issues

```bash
# Check if .gitlab-ci.yml exists in service directory
ls -la path/to/service/.gitlab-ci.yml

# Verify Maven install commands are properly formatted
grep -n "mvn install:install-file" path/to/service/.gitlab-ci.yml

# Check service directory permissions
ls -la path/to/service/
```

### Maven Wrapper Issues

```bash
# Check if wrapper was created
ls -la path/to/service/mvnw

# Manually create wrapper if needed
cd path/to/service && mvn wrapper:wrapper -Dmaven=3.9.6

# Verify wrapper permissions
chmod +x path/to/service/mvnw
```

### Database Location Issues

```bash
# Check current data directory
echo $VERTEX_DATA_DIR

# Verify database file exists
ls -la ~/.local/share/vertex/vertex.db  # Linux
ls -la ~/Library/Application\ Support/Vertex/vertex.db  # macOS
```

### Windows-Specific Issues

```powershell
# Check service status
Get-Service -Name "Vertex"

# Fix data directory permissions
icacls "C:\ProgramData\Vertex" /grant Users:F /T

# Check Windows Firewall
Get-NetFirewallRule -DisplayName "*Vertex*"

# View Windows Event Log
Get-EventLog -LogName Application -Source "Vertex" -Newest 10
```

### Maven Compilation Errors

- Use the automatic "Fix Lombok" feature
- Ensure compatible Maven and Lombok versions
- Check that `./mvnw` wrapper exists in service directories
- Verify Maven wrapper has execution permissions

## 📝 License

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

**Need more help?** Check the [Environment Setup Guide](ENVIRONMENT_SETUP.md) for detailed configuration instructions.
