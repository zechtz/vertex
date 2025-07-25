# NeST Service Manager - Environment Setup

This document explains how to set up and use the simplified environment configuration for NeST microservices.

## Quick Start

1. **Run the setup script** (one-time setup):
   ```bash
   ./setup_environment.sh
   ```

2. **Start the service manager**:
   ```bash
   ./nest-up
   ```

## Environment Configuration

### Common Environment Variables

All services share these common environment variables defined in `common_env_settings.sh`:

- **Database Settings**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`
- **RabbitMQ Settings**: `RABBIT_HOSTNAME`, `RABBIT_PORT`, `RABBIT_USERNAME`, `RABBIT_PASSWORD`
- **Redis Settings**: `REDIS_HOST`, `REDIS_USER`, `REDIS_PASS`
- **Service Discovery**: `DISCOVERY_SERVER`, `DEFAULT_ZONE`
- **Configuration Server**: `CONFIG_SERVER`, `CONFIG_USERNAME`, `CONFIG_PASSWORD`
- **Common Profile**: `ACTIVE_PROFILE` (dev/prod/test)

### Service-Specific Variables

Each service automatically gets its own:
- **Database Name**: `DB_NAME` (e.g., `nest_uaa`, `nest_app`, `nest_contract`, `nest_dsms`)
- **Service Port**: `SERVICE_PORT` (e.g., `8803` for UAA, `8805` for App)

## How It Works

1. **Global Variables**: Common environment variables are loaded from the database
2. **Service Detection**: The service manager detects which service is starting
3. **Automatic Configuration**: Each service gets the appropriate `DB_NAME` and `SERVICE_PORT`
4. **Environment Injection**: All variables are passed to the Maven process

## Service Mapping

| Service Name | Database | Port |
|--------------|----------|------|
| nest-uaa | nest_uaa | 8803 |
| nest-app | nest_app | 8805 |
| nest-contract | nest_contract | 8818 |
| nest-dsms | nest_dsms | 8812 |
| nest-gateway | - | 8802 |
| nest-config-server | - | 8801 |
| nest-registry-server | - | 8800 |

## Configuration Management

You can manage environment variables through:

1. **Web Interface**: Use the "Global Environment Variables" modal in the service manager
2. **Direct File Edit**: Modify `common_env_settings.sh` and reload
3. **Fish Shell**: Use `env_vars.fish` if you prefer Fish shell

## Automatic Loading

The setup script automatically adds environment loading to your shell profile (`~/.zshrc` or `~/.bashrc`), so variables are available in all new terminal sessions.

## Maven & Lombok Compatibility

The service manager includes automatic Lombok compatibility checking. If you encounter compilation errors like "cannot find symbol: method setId()", the system will automatically fix the issue when starting services.

**Manual Fix**: If needed, click the "Fix Lombok" button in the web interface.

**Compatible Versions**:
- Maven: 3.6.3 (automatically uses `./mvnw` wrapper)
- Lombok: 1.18.30
- Java: 17

For detailed information about Lombok configuration, see the main [README.md](README.md) file.

## Troubleshooting

### Environment Variables Not Working
```bash
# Check if variables are loaded
echo $DB_HOST
echo $ACTIVE_PROFILE

# Manually reload if needed
source ./common_env_settings.sh
```

### Service Can't Connect to Database
1. Verify PostgreSQL is running
2. Check database credentials: `DB_USER=postgres`, `DB_PASS=P057gr35`
3. Ensure service-specific database exists (e.g., `nest_uaa`)

### RabbitMQ Connection Issues
1. Verify RabbitMQ is running on localhost:5672
2. Check credentials: `RABBIT_USERNAME=rabbitmq`, `RABBIT_PASSWORD=R@bb17mq`

## Files

- `common_env_settings.sh` - Bash environment variables (primary)
- `env_vars.fish` - Fish shell environment variables (alternative)
- `setup_environment.sh` - One-time setup script
- `ENVIRONMENT_SETUP.md` - This documentation