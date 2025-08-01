# Vertex

A comprehensive Spring Boot microservice management platform providing automated service startup, environment management, monitoring, and build compatibility fixes.

## 📦 Installation

### Option 1: Download Binary (Recommended)

1. **Download the latest binary** from the [Releases](https://github.com/zechtz/service-manager/releases) page
2. **Make it executable**:
   ```bash
   chmod +x vertex
   ```
3. **Add to PATH** (optional but recommended):

   ```bash
   # Move to a directory in your PATH
   sudo mv vertex /usr/local/bin/

   # Or create a symlink
   sudo ln -s /path/to/vertex /usr/local/bin/vertex
   ```

4. **Verify installation**:
   ```bash
   vertex --version
   ```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/your-org/vertex.git
cd vertex

# Build the backend
go build -o vertex

# Build the frontend
cd web
npm install
npm run build
cd ..
```

## 🚀 Quick Start

1. **Start the service manager** (automatic setup):

   ```bash
   # If installed to PATH
   vertex

   # Or run directly
   ./vertex

   # Optionally you can specify a port
   ./vertex --port 3000
   ```

2. **Access the web interface**: Open `http://localhost:8080` in your browser

> **Note**: Environment setup is now automatic! The service manager will automatically detect and configure missing environment variables when started. No manual setup script is needed when using the binary.

## 📋 Features

- **Service Management**: Start, stop, and restart services individually or in bulk
- **Real-time Monitoring**: Live status updates and health checks
- **Log Management**: View, filter, and export service logs
- **Environment Configuration**: Manage global and service-specific environment variables
- **Configuration Profiles**: Save and apply different service configurations
- **Automated Build Fixes**: Automatic Lombok compatibility checking and fixing
- **Automatic Environment Setup**: Built-in environment variable detection and configuration
- **File Management**: Edit service configuration files directly from the web interface

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

```bash
# Build Go backend
go build -o vertex

# Build React frontend
cd web
npm install
npm run build
cd ..
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

## 📊 API Endpoints

### Service Management

- `GET /api/services` - List all services
- `POST /api/services/{name}/start` - Start a service
- `POST /api/services/{name}/stop` - Stop a service
- `POST /api/services/{name}/restart` - Restart a service
- `POST /api/services/start-all` - Start all services
- `POST /api/services/stop-all` - Stop all services

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

### Database Connection Issues

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Verify database exists
psql -h localhost -U postgres -l
```

### Maven Compilation Errors

- Use the automatic "Fix Lombok" feature
- Ensure compatible Maven and Lombok versions
- Check that `./mvnw` wrapper exists in service directories

### Environment Variables Not Loading

```bash
# Check if variables are set
echo $DB_HOST
echo $ACTIVE_PROFILE

# Reload environment
source ./common_env_settings.sh
```

## 📝 License

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

**Need more help?** Check the [Environment Setup Guide](ENVIRONMENT_SETUP.md) for detailed configuration instructions.
