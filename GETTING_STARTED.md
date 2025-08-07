Welcome to Vertex! This guide will walk you through setting up and using Vertex to manage your microservices from initial setup to full service orchestration.

## 📋 Prerequisites

- **Java 17+** (for Spring Boot services)
- **Maven 3.6+** or **Gradle** (for building services)
- **Node.js 18+** and **npm** (if building from source)
- **Services Directory** with your microservice projects

## 🚀 Quick Start

### 1. Installation & Launch

#### Option A: Docker (Recommended)

**Quick Start:**
```bash
# Run with default settings
docker run -d \
  --name vertex \
  -p 8080:8080 \
  -v vertex-data:/app/data \
  zechtz/vertex:latest

# Access the web interface
open http://localhost:8080
```

**Production Setup with Docker Compose:**

Create a `docker-compose.yml` file:
```yaml
version: '3.8'
services:
  vertex:
    image: zechtz/vertex:latest
    container_name: vertex
    ports:
      - "8080:8080"
    volumes:
      - vertex-data:/app/data
      - ./projects:/projects  # Mount your projects directory
    environment:
      - JAVA_HOME=/usr/lib/jvm/default-jvm
      - VERTEX_DATA_DIR=/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  vertex-data:
```

Start with: `docker-compose up -d`

#### Option B: Download Pre-built Binary

```bash
# Download latest release for your platform
wget https://github.com/zechtz/vertex/releases/latest/download/vertex-linux-amd64

# Make executable
chmod +x vertex-linux-amd64

# Install as service with domain (optional)
./vertex-linux-amd64 domain vertex.local

# Or run directly
./vertex-linux-amd64
```

#### Option C: Build from Source

```bash
# Clone repository
git clone https://github.com/zechtz/vertex.git
cd vertex

# Build frontend
cd web && npm install && npm run build && cd ..

# Build backend
go build -o vertex

# Run
./vertex
```

### 2. Access Web Interface

Open your browser and navigate to:

**For Docker installations:**
```
http://localhost:8080
```

**For binary installations:**
```
http://localhost:54321
```

**With domain setup (nginx proxy):**
```
https://vertex.dev        # .dev domains auto-enable HTTPS
http://vertex.local       # Custom domains
```

---

## 🔐 Step 1: Account Setup

### Sign Up / Login

1. **First Time Users**:
   - Click **"Sign Up"** on the login page
   - Fill in your details:
     - **Username**: Choose a unique username (min 3 characters)
     - **Email**: Your valid email address
     - **Password**: Secure password (min 6 characters)
   - Click **"Create Account"**

2. **Returning Users**:
   - Enter your **email** and **password**
   - Click **"Sign In"**

> 💡 **Tip**: Vertex uses JWT tokens for authentication. Your session will remain active across browser restarts.

---

## 📁 Step 2: Create Your First Profile

Profiles allow you to group and manage related services together (e.g., dev, staging, production).

### Creating a Profile

1. **Navigate to Profiles**:
   - Click **"Profiles"** in the sidebar
   - Click **"Create New Profile"**

2. **Basic Information**:

   ```
   Profile Name: My Development Environment
   Description: Development services for my microservice project
   Projects Directory: /path/to/your/services
   ```

3. **Profile Configuration**:
   - **Projects Directory**: Path where your service projects are located
   - **Environment Variables**: Add global variables (optional for now)

4. **Click "Create Profile"**

### Setting Active Profile

- Your new profile will automatically become active
- Switch profiles anytime using the profile dropdown in the top bar

---

## 🔍 Step 3: Auto-Discovery of Services

Vertex can automatically discover Maven and Gradle projects in your workspace.

### Running Auto-Discovery

1. **In Profile Creation/Edit**:
   - Click **"Auto-Discovery"** button
   - Vertex scans your Projects Directory for:
     - `pom.xml` files (Maven projects)
     - `build.gradle` files (Gradle projects)
     - `application.properties` files for port detection

2. **Review Discovered Services**:
   - Check/uncheck services to include
   - Vertex automatically detects:
     - **Service names** (from directory names)
     - **Ports** (from application.properties)
     - **Build systems** (Maven/Gradle)

3. **Manual Service Addition** (if needed):
   - Click **"Add Custom Service"**
   - Fill in service details manually

### Example Auto-Discovery Result

```
✅ eureka-server (Port: 8800, Maven)
✅ config-server (Port: 8801, Maven)
✅ api-gateway (Port: 8080, Maven)
✅ user-service (Port: 8081, Maven)
✅ product-service (Port: 8082, Gradle)
```

---

## ⚡ Step 4: Configure Startup Order

Proper startup order ensures dependencies are available when services start.

### Setting Service Order

1. **Navigate to Configurations**:
   - Go to **"Configurations"** in the sidebar
   - Click **"New Configuration"** in your profile
   - Add a **"Configuration Name"**

2. **Drag & Drop Ordering**:
   - **Registry Services** (like Eureka): Order 1-2
   - **Config Services**: Order 3-4
   - **Infrastructure** (Gateway, Auth): Order 5-10
   - **Business Services**: Order 11+

When you're done, just save and apply the configuration and it will become the default configuration

### Recommended Startup Order

```
1. eureka-server      (Service Registry)
2. config-server      (Configuration)
3. api-gateway        (Gateway)
4. auth-service       (Authentication)
5. user-service       (Business Logic)
6. product-service    (Business Logic)
7. notification-service
```

### Auto-Generated Order

- Vertex can suggest order based on common microservice patterns
- Click **"Auto-Generate Order"** for automatic ordering

---

## 🔗 Step 5: Define Service Dependencies

Dependencies ensure services wait for required services to be healthy before starting.

### Adding Dependencies

1. **In Dependencies View**:
   - Select a service from the list
   - Click **"Add Dependency"**

2. **Configure Dependency**:

   ```
   Service: user-service
   Depends On: eureka-server, config-server
   Wait Time: 30 seconds
   Health Check: Required
   ```

3. **Dependency Types**:
   - **Hard Dependency**: Service won't start without it
   - **Soft Dependency**: Service starts but waits if available
   - **Health Check**: Verifies dependency is healthy

### Example Dependencies

```
api-gateway depends on:
  ├── eureka-server (health check required)
  └── config-server (health check required)

user-service depends on:
  ├── eureka-server (health check required)
  ├── config-server (health check required)
  └── database-service (hard dependency)
```

---

## 🌍 Step 6: Environment Variables Management

Manage global and service-specific environment variables.

### Global Environment Variables

1. **Access Global Settings**:
   - Click **"Settings"** → **"Global Configuration"**
   - Or click **"Environment Variables"** in sidebar

2. **Add Global Variables**:

   ```
   DB_HOST=localhost
   DB_PORT=5432
   REDIS_HOST=localhost
   EUREKA_URL=http://localhost:8800/eureka
   ACTIVE_PROFILE=dev
   ```

3. **Variable Categories**:
   - **Database**: DB_HOST, DB_PORT, DB_USER
   - **Cache**: REDIS_HOST, REDIS_PORT
   - **Config**: CONFIG_SERVER_URL
   - **Network**: SERVICE URLs and ports

### Service-Specific Variables

1. **Edit Service**:
   - Click on any service card
   - Go to **"Environment Variables"** tab

2. **Add Service Variables**:
   ```
   SERVICE_PORT=8081
   SERVICE_NAME=user-service
   DATABASE_URL=jdbc:postgresql://localhost:5432/users
   ```

### Variable Precedence

- **Service-specific** variables override global variables
- **Environment Variables** are automatically passed to services
- **Spring Profiles**: `ACTIVE_PROFILE` sets `SPRING_PROFILES_ACTIVE`

### Export/Import Variables

- **Export**: Copy variables to clipboard for sharing
- **Bulk Import**: Import from text or environment files

---

## 🎯 Step 7: Service Operations

### Starting Services

#### Individual Services

1. **Start Single Service**:
   - Click **"Start"** button on service card
   - Watch real-time status updates
   - Health status shows: Starting → Running → Healthy

#### Bulk Operations

1. **Start All Services**:
   - Click **"Start All"** in top toolbar
   - Services start in configured dependency order
   - Watch progress in real-time

2. **Profile-Aware Starting**:
   - Only services in active profile start
   - Respects dependency chains
   - Automatic health checking

### Stopping Services

1. **Stop Individual Service**:
   - Click **"Stop"** on service card
   - Graceful shutdown with SIGTERM
   - Force kill if needed

2. **Stop All Services**:
   - Click **"Stop All"** in toolbar
   - Reverse dependency order
   - Prevents cascade failures

### Service Status Indicators

| Status        | Color     | Meaning                    |
| ------------- | --------- | -------------------------- |
| **Healthy**   | 🟢 Green  | Running and responding     |
| **Starting**  | 🟡 Yellow | Boot process active        |
| **Unhealthy** | 🔴 Red    | Running but not responding |
| **Stopped**   | ⚪ Gray   | Not running                |

---

## 📊 Step 8: Monitoring & Logs

### Real-Time Logs

1. **View Service Logs**:
   - Click **"Logs"** on any service card
   - Real-time log streaming
   - Color-coded by log level (INFO, WARN, ERROR)

2. **Log Features**:
   - **Search**: Filter logs by keyword
   - **Level Filter**: Show only ERROR/WARN logs
   - **Auto-scroll**: Follow latest logs
   - **Export**: Download logs as text file

### System Metrics

1. **Access Metrics**:
   - Click **"System Metrics"** in sidebar
   - View system-wide service health

2. **Available Metrics**:
   - **CPU Usage**: Per service resource usage
   - **Memory**: RAM consumption by service
   - **Network**: Request/response metrics
   - **Uptime**: Service availability time

### Service Topology

1. **Visualize Dependencies**:
   - Click **"Topology"** in sidebar
   - Interactive dependency graph
   - Service health visualization
   - Connection status between services

---

## 🔧 Step 9: Advanced Features

### Maven Library Installation

1. **Install Service Libraries**:
   - Click **"Install Libraries"** on service card
   - Vertex parses `.gitlab-ci.yml` for Maven dependencies
   - Select environments to install libraries for
   - Monitor installation progress

### Lombok Compatibility

1. **Automatic Lombok Fixes**:
   - Click **"Fix Lombok"** in toolbar
   - Vertex updates `pom.xml` with proper Lombok configuration
   - Fixes compilation issues automatically

### Port Management

1. **Automatic Port Cleanup**:
   - Vertex cleans up processes on service ports before starting
   - Prevents "port already in use" errors
   - Graceful process termination

### Configuration Files

1. **Edit Service Files**:
   - Click **"Edit Files"** on service card
   - Edit `application.properties`, `application.yml`
   - Direct file editing with syntax highlighting

---

## 🎛️ Step 10: Configuration Management

### Backup & Restore

1. **Export Configuration**:
   - Profiles, services, dependencies, and environment variables
   - JSON format for easy backup
   - Share configurations between team members

2. **Import Configuration**:
   - Restore from backup files
   - Merge with existing configurations
   - Validate before applying

### Team Collaboration

1. **Shared Profiles**:
   - Export profiles for team use
   - Consistent development environments
   - Version-controlled configurations

---

## 🚨 Troubleshooting

### Common Issues

#### Service Won't Start

1. **Check Dependencies**: Ensure required services are running
2. **Port Conflicts**: Use port cleanup feature  
3. **Environment Variables**: Verify all required variables are set
4. **Java Version**: Ensure Java 17+ is available
5. **Docker Volume Permissions**: Check if mounted directories are accessible

#### Docker-Specific Issues

1. **Container Won't Start**:
   ```bash
   # Check container logs
   docker logs vertex
   
   # Check container status
   docker ps -a
   ```

2. **Port Already in Use**:
   ```bash
   # Find processes using port 8080
   lsof -i :8080
   
   # Stop conflicting containers
   docker stop $(docker ps -q --filter "publish=8080")
   ```

3. **Volume Mount Issues**:
   ```bash
   # Verify volume exists
   docker volume ls | grep vertex-data
   
   # Check volume permissions
   docker run --rm -v vertex-data:/data alpine ls -la /data
   ```

4. **Java Not Found in Container**:
   ```bash
   # Check Java installation in container
   docker exec vertex java -version
   ```

#### Compilation Errors

1. **Lombok Issues**: Use "Fix Lombok" feature
2. **Maven Wrapper**: Vertex auto-creates missing wrappers  
3. **Dependencies**: Check if libraries are properly installed
4. **Docker Build Context**: Ensure projects are accessible within container

#### Connection Issues

1. **Eureka Registration**: Verify registry server is running
2. **Network Configuration**: Check service URLs and ports
3. **Health Checks**: Monitor service health endpoints
4. **Docker Networking**: Services in containers may need different host references

### Getting Help

- **Logs**: Check service logs for detailed error messages
- **Debug Mode**: Enable debug logging in global configuration
- **GitHub Issues**: Report bugs and request features
- **Documentation**: Refer to README.md for detailed setup

---

## 🎉 Next Steps

Now that you have Vertex set up and running:

1. **Explore Advanced Features**:
   - Set up multiple profiles for different environments
   - Create complex dependency chains
   - Use metrics for performance monitoring

2. **Automation**:
   - Set up automatic service health monitoring
   - Configure alerts for service failures
   - Implement CI/CD integration

3. **Team Setup**:
   - Share profiles with team members
   - Standardize development environments
   - Document service dependencies

4. **Maintenance**:
   - Regular updates for security and features
   - Backup your configurations and data
   - Monitor resource usage and optimize

**Welcome to efficient microservice management with Vertex!** 🚀

---

## 🔄 Update & Maintenance

### Updating Vertex

**For Docker installations:**
```bash
# Using docker-compose
docker-compose pull
docker-compose up -d

# Manual update
docker pull zechtz/vertex:latest
docker stop vertex
docker rm vertex
docker run -d --name vertex -p 8080:8080 -v vertex-data:/app/data zechtz/vertex:latest
```

**For binary installations:**
```bash
# Using built-in updater
./vertex update

# Manual update
wget https://github.com/zechtz/vertex/releases/latest/download/vertex-linux-amd64
chmod +x vertex-linux-amd64
# Stop current service and replace binary
```

### Backup Your Data

**Docker:**
```bash
# Backup database and configuration
docker run --rm -v vertex-data:/data -v $(pwd):/backup alpine tar czf /backup/vertex-backup.tar.gz /data
```

**Binary installation:**
```bash
# Backup user data directory
tar czf vertex-backup.tar.gz ~/.vertex/
```

### Uninstalling

**Docker:**
```bash
# Remove everything (WARNING: This deletes all data!)
docker-compose down -v
docker rmi zechtz/vertex:latest
```

**Binary installation:**
```bash
# Self-uninstall
./vertex uninstall
```

---

## 📚 Additional Resources

- [Installation Guide](INSTALLATION.md) - Platform-specific installation
- [README.md](README.md) - Complete feature documentation
- [Environment Setup](ENVIRONMENT_SETUP.md) - Detailed configuration guide
- [GitHub Repository](https://github.com/zechtz/vertex) - Source code and issues
