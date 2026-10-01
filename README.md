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

2. **Make it executable and bypass macOS security (macOS only):**

   ```bash
   chmod +x vertex-*

   # macOS: Remove quarantine to bypass "malware" warning
   xattr -d com.apple.quarantine vertex-darwin-*
   ```

3. **Install as a user service:**

   ```bash
   # 🚀 ONE-COMMAND INSTALLATION (modern syntax - recommended)
   ./vertex-linux-amd64 domain vertex.local      # Auto-installs with nginx!
   ./vertex-darwin-arm64 domain myapp.local      # macOS example
   # vertex-windows-amd64.exe domain myapp.local  (Windows example)

   # 🚀 ONE-COMMAND INSTALLATION (traditional syntax - alternative)
   ./vertex-linux-amd64 --domain vertex.local    # Auto-installs with nginx!
   ./vertex-darwin-arm64 --domain myapp.local    # macOS example
   # vertex-windows-amd64.exe --domain myapp.local  (Windows example)

   # Basic installation options
   ./vertex-linux-amd64 install                  # Basic (modern)
   ./vertex-linux-amd64 --install                # Basic (traditional)
   ./vertex-linux-amd64 install nginx domain vertex.local  # Full (modern)
   ./vertex-linux-amd64 --install --nginx --domain vertex.local  # Full (traditional)
   ```

#### Option 2: Docker

Vertex starts your services inside its own container, so the image carries what they need: **JDK 17 and JDK 21** (21 is the default), Maven, Git and bash. Your projects are mounted in from the host.

```bash
docker run -d \
  --name vertex \
  -p 54321:54321 \
  -p 8080-8099:8080-8099 \
  -v vertex-data:/app/data \
  -v "$HOME/projects:$HOME/projects" \
  -v "$HOME/.m2:/root/.m2" \
  zechtz/vertex:latest

# Access the web interface
open http://localhost:54321
```

What each mount and port is for:

| Option | Why |
| --- | --- |
| `-p 54321:54321` | The Vertex web interface |
| `-p 8080-8099:8080-8099` | Your services' ports. They run inside the container, so publish the range they use |
| `-v vertex-data:/app/data` | Vertex's database and logs, kept across container upgrades |
| `-v "$HOME/projects:$HOME/projects"` | Your projects, mounted at **the same path** as on the host, so directories stored in Vertex and paths in IntelliJ run configurations stay valid |
| `-v "$HOME/.m2:/root/.m2"` | Your Maven `settings.xml` (private repositories and their credentials) and dependency cache, so builds don't download everything again |

**Docker Compose:**

```yaml
services:
  vertex:
    image: zechtz/vertex:latest
    container_name: vertex
    ports:
      - "54321:54321"
      - "8080-8099:8080-8099" # your services' ports
    volumes:
      - vertex-data:/app/data
      - ${HOME}/projects:${HOME}/projects # same path as on the host
      - ${HOME}/.m2:/root/.m2
    restart: unless-stopped

volumes:
  vertex-data:
```

##### Java in the container

A service runs on a JDK **inside the container**. It cannot use the JDK installed on your machine: on macOS, Docker runs Linux, which cannot run a macOS JDK.

| JDK | `JAVA_HOME` |
| --- | --- |
| 21 (default) | `/opt/java/21` |
| 17 | `/opt/java/17` |

- A service with no `JAVA_HOME` runs on JDK 21.
- To run a service on JDK 17, set its `JAVA_HOME` environment variable to `/opt/java/17`. The JDK picker lists the JDKs in the container.
- A `JAVA_HOME` copied from your machine, such as `/Library/Java/JavaVirtualMachines/temurin-17.jdk/Contents/Home`, does not exist in the container, and a service using it will not start. The same applies to a profile's Java Home override.
- The IntelliJ import resolves a run configuration's JDK against the JDKs in the container. A configuration that names a JDK version the container doesn't have is reported, not guessed.

**Need another Java version?** Build a small image on top of Vertex's:

```dockerfile
FROM zechtz/vertex:latest
COPY --from=eclipse-temurin:11-jdk /opt/java/openjdk /opt/java/11
```

Then set `JAVA_HOME=/opt/java/11` on the services that need it.

##### Networking

- **Services reaching each other** works as usual: they all run in the same container, so `localhost` reaches them.
- **Services reaching something on your machine**, such as a database, need `host.docker.internal` instead of `localhost`. On Linux, also add `--add-host=host.docker.internal:host-gateway` (Compose: `extra_hosts: ["host.docker.internal:host-gateway"]`).
- **On Linux**, `--network host` (Compose: `network_mode: host`) avoids both: services share the host's network, and no ports need publishing.

##### Other notes

- **Git over SSH:** mount your keys read-only with `-v "$HOME/.ssh:/root/.ssh:ro"`.
- **File ownership on Linux:** Vertex runs as root in the container, so build output it writes into your projects (`target/`, `build/`) is owned by root on the host. Docker Desktop on macOS and Windows maps ownership to your user, so this only affects Linux.

##### HTTPS on a domain (like native `./vertex domain vertex.dev`)

Put nginx in front of Vertex. Vertex no longer publishes its own port; the services' ports still need publishing:

```yaml
services:
  vertex:
    image: zechtz/vertex:latest
    container_name: vertex
    expose:
      - "54321"
    ports:
      - "8080-8099:8080-8099" # your services' ports
    volumes:
      - vertex-data:/app/data
      - ${HOME}/projects:${HOME}/projects
      - ${HOME}/.m2:/root/.m2
    restart: unless-stopped
    networks:
      - vertex-network

  nginx:
    image: nginx:alpine
    container_name: vertex-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - vertex
    restart: unless-stopped
    networks:
      - vertex-network

networks:
  vertex-network:
    driver: bridge

volumes:
  vertex-data:
```

**Required nginx.conf:**

```nginx
events {
    worker_connections 1024;
}

http {
    upstream vertex {
        server vertex:54321;
    }

    server {
        listen 80;
        server_name vertex.dev;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl;
        http2 on;
        server_name vertex.dev;

        ssl_certificate /etc/nginx/ssl/vertex.dev.pem;
        ssl_certificate_key /etc/nginx/ssl/vertex.dev-key.pem;

        # Modern SSL configuration
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        location / {
            proxy_pass http://vertex;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket support
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }
}
```

**Setup steps:**

1. **Let your user talk to Docker (Linux).** Without this you get `permission denied while trying to connect to the docker API at unix:///var/run/docker.sock`:

   ```bash
   sudo usermod -aG docker $USER
   # Group changes apply on next login: log out of your desktop session (on
   # Omarchy/Hyprland, exit the whole Hyprland session, not just the terminal)
   # and back in, or run `newgrp docker` in the current terminal
   id   # should now list "docker"
   ```

2. **Install mkcert** (plus NSS tools so browsers can trust the certificates):

   ```bash
   # macOS:  brew install mkcert nss
   # Arch / Omarchy: sudo pacman -S mkcert nss
   # Ubuntu: sudo apt install mkcert libnss3-tools
   ```

3. **Install the local CA and generate certificates.** Create `ssl/` yourself _before_ the first `docker compose up`, otherwise Docker creates it owned by root and mkcert can't write to it:

   ```bash
   # Linux + Chromium/Chrome/Brave (e.g. Omarchy, which ships Chromium as the
   # default browser): these browsers use their own NSS store at
   # ~/.pki/nssdb, and mkcert only installs into it if it already exists
   mkdir -p ~/.pki/nssdb
   [ -f ~/.pki/nssdb/cert9.db ] || certutil -d sql:$HOME/.pki/nssdb -N --empty-password

   mkcert -install

   mkdir -p ssl   # if Docker already created it: sudo chown $USER: ssl
   mkcert -cert-file ssl/vertex.dev.pem -key-file ssl/vertex.dev-key.pem \
          vertex.dev localhost 127.0.0.1
   ```

4. **Point the domain at localhost:**

   ```bash
   echo "127.0.0.1 vertex.dev" | sudo tee -a /etc/hosts
   ```

5. **Start the stack:**

   ```bash
   docker compose up -d
   docker compose logs -f nginx   # should end with "Configuration complete; ready for start up"
   ```

6. **Open https://vertex.dev.** Fully restart your browser after `mkcert -install` so it picks up the new CA.

**Docker troubleshooting:**

| Symptom                                                                           | Cause / Fix                                                                                                                                                                                                                                                                                                                       |
| --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `permission denied ... /var/run/docker.sock`                                      | User not in the `docker` group, or the session predates it. See step 1.                                                                                                                                                                                                                                                           |
| `ERROR: failed to save certificate: open ssl/...: permission denied`              | `ssl/` was created by Docker as root. Run `sudo chown $USER: ssl`.                                                                                                                                                                                                                                                                |
| nginx restart loop with `SSL_CTX_set_cipher_list(...) failed ... no cipher match` | Invalid cipher names in `nginx.conf` (e.g. `...-GCM-SHA512`, which doesn't exist). Use the `...-GCM-SHA384` suites shown above.                                                                                                                                                                                                   |
| nginx fails with `cannot load certificate "/etc/nginx/ssl/vertex.dev.pem"`        | Certificates not generated yet. See step 3.                                                                                                                                                                                                                                                                                       |
| `curl https://vertex.dev` works but the browser shows a certificate error         | The mkcert CA isn't in the browser's trust store. On Linux Chromium (e.g. the default browser on a fresh Omarchy install, where `~/.pki/nssdb` doesn't exist yet), create `~/.pki/nssdb` (step 3), re-run `mkcert -install`, and restart the browser. `.dev` is HSTS-preloaded, so browsers won't let you click past the warning. |
| A service fails with `The JAVA_HOME environment variable is not defined correctly` (Maven) or `JAVA_HOME is set to an invalid directory` (Gradle) | Its `JAVA_HOME` (or the profile's Java Home override) is a path from your machine. Set it to a JDK in the container, such as `/opt/java/17` — see [Java in the container](#java-in-the-container).                                                                                                                                |
| `Using fallback JWT secret`                                                       | Fine for local use. Set `JWT_SECRET` under `environment:` for anything real.                                                                                                                                                                                                                                                      |

#### Option 3: Build from Source

1. **Build the application:**

   ```bash
   go build -o vertex
   ```

2. **Install as a user service:**

   ```bash
   # 🚀 ONE-COMMAND INSTALLATION (modern syntax - recommended)
   ./vertex domain myapp.local          # Auto-installs with nginx!

   # 🚀 ONE-COMMAND INSTALLATION (traditional syntax - alternative)
   ./vertex --domain myapp.local        # Auto-installs with nginx!

   # Basic installation options
   ./vertex install                     # Basic (modern)
   ./vertex --install                   # Basic (traditional)
   ./vertex install nginx domain myapp.local  # Full (modern)
   ./vertex --install --nginx --domain myapp.local  # Full (traditional)
   ```

3. **Access the web interface:**
   - **Docker**: http://localhost:54321
   - **With HTTPS domain**: https://vertex.dev (when using `--domain vertex.dev`)
   - **With HTTP domain**: http://myapp.local (when using `--domain myapp.local`)
   - **Direct access**: http://localhost:54321

> 📖 **For detailed usage instructions and tutorials, see our [Getting Started Guide](https://github.com/zechtz/vertex/wiki/Getting-Started-with-Vertex-Service-Manager)** on the wiki.

## 🚀 Usage

### 🌐 Nginx Proxy Configuration

Vertex includes optional nginx integration for clean domain-based access without port numbers.

#### Quick Setup

```bash
# 🚀 ONE-COMMAND INSTALLATION (modern syntax - recommended)
./vertex domain vertex.dev

# 🚀 ONE-COMMAND INSTALLATION (traditional syntax - alternative)
./vertex --domain vertex.dev

# Access via clean domain with HTTPS (auto-enabled for .dev domains)
open https://vertex.dev
```

#### Custom Domain

```bash
# One-command installation with custom domain (modern syntax - recommended)
./vertex domain myapp.local           # HTTP
./vertex domain myapp.local https     # With HTTPS

# One-command installation with custom domain (traditional syntax - alternative)
./vertex --domain myapp.local         # HTTP
./vertex --domain myapp.local --https # With HTTPS

# Access your custom domain
open https://myapp.local    # With HTTPS
open http://myapp.local     # HTTP only
```

#### Advanced Configuration

```bash
# Modern subcommand syntax (recommended)
./vertex install nginx https domain myproject.local port 54321  # Full installation
./vertex domain myproject.local port 8080                       # One-command with custom port
./vertex domain myproject.local https                           # Force HTTPS for any domain

# Traditional flag syntax (alternative)
./vertex --install \
  --nginx \                    # Enable nginx proxy
  --https \                    # Enable HTTPS with locally-trusted certificates
  --domain myproject.local \   # Custom domain name
  --port 54321                 # Vertex service port (default: 54321)

./vertex --domain myproject.local --port 8080                   # One-command with custom port
./vertex --domain myproject.local --https                       # Force HTTPS for any domain
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

| Method                  | URL                      | Use Case                                              |
| ----------------------- | ------------------------ | ----------------------------------------------------- |
| **Nginx Proxy (HTTPS)** | `https://vertex.dev`     | Secure domain access (.dev domains auto-enable HTTPS) |
| **Nginx Proxy (HTTP)**  | `http://myproject.local` | Clean domain access for non-.dev domains              |
| **Direct Access**       | `http://localhost:54321` | Development, bypassing nginx                          |

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
- **Windows**: Scheduled Task

#### Built-in Service Commands (Recommended)

Vertex includes built-in commands that work across all platforms. You can use either the modern subcommand syntax or the traditional flag syntax:

```bash
# Start the service
./vertex start          # (recommended)
./vertex --start        # (alternative)

# Stop the service
./vertex stop           # (recommended)
./vertex --stop         # (alternative)

# Restart the service
./vertex restart        # (recommended)
./vertex --restart      # (alternative)

# Check service status and available URLs
./vertex status         # (recommended)
./vertex --status       # (alternative)

# Show recent logs
./vertex logs           # (recommended)
./vertex --logs         # (alternative)

# Follow logs in real-time (press Ctrl+C to exit)
./vertex logs -f        # (recommended - like tail -f)
./vertex logs --follow  # (explicit long form)
./vertex --logs --follow # (traditional syntax)
```

#### Platform-Specific Commands (Advanced)

**macOS:**

```bash
# Start
launchctl start com.vertex.manager

# Stop
launchctl stop com.vertex.manager

# Check status
launchctl list | grep vertex
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

**Windows:**

```bash
# Start
schtasks /run /tn "VertexServiceManager"

# Stop
schtasks /end /tn "VertexServiceManager"

# Check status
schtasks /query /tn "VertexServiceManager"
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

#### Built-in Log Commands (Recommended)

```bash
# Show recent logs from all sources
./vertex logs           # (recommended)
./vertex --logs         # (alternative)

# Follow logs in real-time (press Ctrl+C to exit)
./vertex logs -f        # (recommended - like tail -f)
./vertex logs --follow  # (explicit long form)
./vertex --logs --follow # (traditional syntax)
```

#### Platform-Specific Log Access

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

**Windows:**

```bash
# View log files directly
type %USERPROFILE%\.vertex\vertex.log
```

#### Log Locations

| Platform     | Location                                                       |
| ------------ | -------------------------------------------------------------- |
| **macOS**    | `~/.vertex/vertex.stderr.log`<br>`~/.vertex/vertex.stdout.log` |
| **Linux**    | `journalctl --user -u vertex`                                  |
| **Database** | `~/.vertex/vertex.db`                                          |
| **Config**   | `~/.vertex/`                                                   |

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

Vertex supports both modern subcommand syntax and traditional flag syntax:

```bash
./vertex --help
```

**Service Management Commands:**

| Subcommand         | Flag              | Description                                      |
| ------------------ | ----------------- | ------------------------------------------------ |
| `vertex start`     | `--start`         | Start the Vertex service                         |
| `vertex stop`      | `--stop`          | Stop the Vertex service                          |
| `vertex restart`   | `--restart`       | Restart the Vertex service                       |
| `vertex status`    | `--status`        | Show service status and availability             |
| `vertex logs`      | `--logs`          | Show service logs                                |
| `vertex logs -f`   | `--logs --follow` | Follow log output (like tail -f)                 |
| `vertex install`   | `--install`       | Install Vertex as a user service                 |
| `vertex uninstall` | `--uninstall`     | Uninstall Vertex service and data                |
| `vertex update`    | `--update`        | Update the Vertex binary and restart the service |
| `vertex version`   | `--version`       | Show version information                         |

**Configuration Commands:**

| Subcommand               | Flag                | Default    | Description                                                                      |
| ------------------------ | ------------------- | ---------- | -------------------------------------------------------------------------------- |
| `vertex domain <name>`   | `--domain <name>`   | vertex.dev | **🚀 Smart install**: Domain name for nginx proxy (auto-installs when specified) |
| `vertex port <number>`   | `--port <number>`   | 54321      | Port to run the server on                                                        |
| `vertex data-dir <path>` | `--data-dir <path>` | ~/.vertex  | Directory to store application data                                              |
| `vertex nginx`           | `--nginx`           | -          | Configure nginx proxy for domain access                                          |
| `vertex https`           | `--https`           | -          | Enable HTTPS with locally-trusted certificates (auto-enabled for .dev domains)   |

#### Examples

```bash
# 🚀 ONE-COMMAND INSTALLATION (modern subcommand syntax - recommended)
./vertex domain myapp.local          # HTTP installation
./vertex domain vertex.dev           # HTTPS auto-enabled for .dev domains

# 🚀 ONE-COMMAND INSTALLATION (traditional flag syntax - alternative)
./vertex --domain myapp.local        # HTTP installation
./vertex --domain vertex.dev         # HTTPS auto-enabled for .dev domains

# Installation Commands (modern subcommand syntax - recommended)
./vertex install                     # Basic installation
./vertex install nginx               # With nginx proxy
./vertex install nginx https domain myapp.local  # With HTTPS
./vertex domain myapp.local port 8080  # Custom domain and port

# Installation Commands (traditional flag syntax - alternative)
./vertex --install                   # Basic installation
./vertex --install --nginx          # With nginx proxy
./vertex --install --nginx --https --domain myapp.local  # With HTTPS
./vertex --install --nginx --domain myapp.local --port 8080  # Full explicit

# Service Management (modern subcommand syntax - recommended)
./vertex start                       # Start the service
./vertex stop                        # Stop the service
./vertex restart                     # Restart the service
./vertex status                      # Show service status and URLs
./vertex logs                        # Show recent logs
./vertex logs -f                     # Follow logs in real-time (like tail -f)
./vertex version                     # Show version
./vertex update                      # Update service

# Service Management (traditional flag syntax - alternative)
./vertex --start                     # Start the service
./vertex --stop                      # Stop the service
./vertex --restart                   # Restart the service
./vertex --status                    # Show service status and URLs
./vertex --logs                      # Show recent logs
./vertex --logs --follow             # Follow logs in real-time
./vertex --version                   # Show version
./vertex --update                    # Update service

# Configuration Commands (modern subcommand syntax - recommended)
./vertex port 9090                   # Run on custom port
./vertex data-dir /tmp/vertex-test   # Custom data directory
./vertex domain myproject.local https  # Force HTTPS for any domain
./vertex nginx                       # Enable nginx proxy
./vertex https                       # Enable HTTPS

# Configuration Commands (traditional flag syntax - alternative)
./vertex --port 9090                 # Run on custom port
./vertex --data-dir /tmp/vertex-test # Custom data directory
./vertex --domain myproject.local --https  # Force HTTPS for any domain
./vertex --nginx                     # Enable nginx proxy
./vertex --https                     # Enable HTTPS

# Combined Examples (modern syntax)
./vertex domain vertex.dev https port 8080  # Full configuration
./vertex install nginx domain myapp.local data-dir /custom/path  # Advanced install
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

### macOS Security Warning ("cannot verify vertex is free of malware")

This is a common macOS Gatekeeper security warning for unsigned binaries.

**Quick Fix:**

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ./vertex-darwin-arm64

# Then run your command
./vertex domain vertex.dev
```

**Alternative Fix:**

1. Try running the command and get the security warning
2. Go to **System Preferences → Security & Privacy → General**
3. Click **"Allow Anyway"** next to the vertex warning
4. Run the command again

**For Developers:**
Consider code signing your releases with an Apple Developer Certificate to eliminate this warning for users.

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
   ./vertex update       # (recommended)
   ./vertex --update     # (alternative)
   ```

This command will:

- Stop the running Vertex service.
- Replace the existing binary with the new one.
- Restart the service.

### Manual Update Method

**For Native Installation:**

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

**For Docker Installation:**

```bash
# Pull latest image and restart
docker-compose pull
docker-compose up -d

# Or with docker run
docker pull zechtz/vertex:latest
docker stop vertex
docker rm vertex
docker run -d --name vertex -p 54321:54321 -v vertex-data:/app/data zechtz/vertex:latest
```

## 🗑️ Uninstalling

**For Docker Installation:**

```bash
# Using docker-compose
docker-compose down -v

# Or manually
docker stop vertex
docker rm vertex
docker volume rm vertex-data  # This removes all data!
docker rmi zechtz/vertex:latest
```

**For Native Installation:**

```bash
# Self-uninstalling - works on all platforms!
./vertex uninstall       # (recommended)
./vertex --uninstall     # (alternative)
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
cd web
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
