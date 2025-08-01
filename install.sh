#!/bin/bash

# Vertex Installation Script
set -e

BINARY_NAME="vertex"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/vertex"
SERVICE_FILE="/etc/systemd/system/vertex.service"
USER="vertex"
GROUP="vertex"

echo "🚀 Installing Vertex Service Manager..."

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root (use sudo)"
   exit 1
fi

# Create system user if it doesn't exist
if ! id "$USER" &>/dev/null; then
    echo "👤 Creating system user: $USER"
    useradd --system --home "$DATA_DIR" --shell /bin/false "$USER"
fi

# Create data directory
echo "📁 Creating data directory: $DATA_DIR"
mkdir -p "$DATA_DIR"
chown "$USER:$GROUP" "$DATA_DIR"
chmod 755 "$DATA_DIR"

# Copy binary
if [[ -f "./$BINARY_NAME" ]]; then
    echo "📦 Installing binary to $INSTALL_DIR/$BINARY_NAME"
    cp "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod 755 "$INSTALL_DIR/$BINARY_NAME"
else
    echo "❌ Binary $BINARY_NAME not found in current directory"
    echo "Please build the application first with: go build -o $BINARY_NAME"
    exit 1
fi

# Create systemd service file
echo "🔧 Creating systemd service file"
cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Vertex Service Manager
Documentation=https://github.com/zechtz/vertex
After=network.target
Wants=network.target

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$DATA_DIR
Environment=VERTEX_DATA_DIR=$DATA_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME -port 8080
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=vertex

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$DATA_DIR

[Install]
WantedBy=multi-user.target
EOF

# Set permissions for service file
chmod 644 "$SERVICE_FILE"

# Reload systemd and enable service
echo "🔄 Reloading systemd and enabling vertex service"
systemctl daemon-reload
systemctl enable vertex

echo "✅ Installation completed successfully!"
echo ""
echo "📋 Next steps:"
echo "   • Start the service: sudo systemctl start vertex"
echo "   • Check status: sudo systemctl status vertex"
echo "   • View logs: sudo journalctl -u vertex -f"
echo "   • Access web interface: http://localhost:8080"
echo ""
echo "📂 Data directory: $DATA_DIR"
echo "🗄️  Database location: $DATA_DIR/vertex.db"
echo "👤 Service runs as user: $USER"