#!/bin/bash

# Vertex Installation Script
set -e

BINARY_NAME="vertex"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/vertex"
USER="vertex"
GROUP="vertex"
PLIST_FILE="/Library/LaunchDaemons/com.vertex.manager.plist"

echo "🚀 Installing Vertex Service Manager..."

# Check if running as root
if [[ $EUID -ne 0 ]]; then
	echo "❌ This script must be run as root (use sudo)"
	exit 1
fi

OS="$(uname)"

if [[ "$OS" == "Linux" ]]; then
	# Create system user if it doesn't exist
	if ! id "$USER" &>/dev/null; then
		echo "👤 Creating system user: $USER"
		useradd --system --home "$DATA_DIR" --shell /bin/false "$USER"
	else
		echo "✅ System user '$USER' already exists"
	fi
elif [[ "$OS" == "Darwin" ]]; then
	echo "🍎 macOS detected: running as root user"
	USER="root"
	GROUP="wheel"
else
	echo "❌ Unsupported OS: $OS"
	exit 1
fi

# Create data directory
if [[ ! -d "$DATA_DIR" ]]; then
	echo "📁 Creating data directory: $DATA_DIR"
	mkdir -p "$DATA_DIR"
else
	echo "✅ Data directory already exists: $DATA_DIR"
fi
chown "$USER:$GROUP" "$DATA_DIR"
chmod 755 "$DATA_DIR"

# Check for existing installation and stop service if running
if [[ "$OS" == "Linux" ]]; then
	if systemctl is-active --quiet vertex; then
		echo "🛑 Stopping existing vertex service"
		systemctl stop vertex
	fi
elif [[ "$OS" == "Darwin" ]]; then
	if launchctl list | grep -q "com.vertex.manager"; then
		echo "🛑 Stopping existing vertex service"
		launchctl stop com.vertex.manager &>/dev/null || true
		launchctl unload "$PLIST_FILE" &>/dev/null || true
	fi
fi

# Copy binary
if [[ -f "./$BINARY_NAME" ]]; then
	if [[ -f "$INSTALL_DIR/$BINARY_NAME" ]]; then
		echo "📦 Updating binary at $INSTALL_DIR/$BINARY_NAME"
	else
		echo "📦 Installing binary to $INSTALL_DIR/$BINARY_NAME"
	fi
	cp "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
	chmod 755 "$INSTALL_DIR/$BINARY_NAME"
else
	echo "❌ Binary $BINARY_NAME not found in current directory"
	echo "Please build the application first with: go build -o $BINARY_NAME"
	exit 1
fi

if [[ "$OS" == "Linux" ]]; then
	# Create systemd service file
	SERVICE_FILE="/etc/systemd/system/vertex.service"
	if [[ ! -f "$SERVICE_FILE" ]]; then
		echo "🔧 Creating systemd service file"
		cat >"$SERVICE_FILE" <<EOF
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
		chmod 644 "$SERVICE_FILE"
		echo "🔄 Reloading systemd and enabling vertex service"
		systemctl daemon-reload
		systemctl enable vertex
	else
		echo "✅ Systemd service file already exists"
		echo "🔄 Reloading systemd daemon"
		systemctl daemon-reload
	fi

	echo "✅ Installation completed successfully on Linux!"
	echo ""
	echo "📋 Next steps:"
	echo "   • Start the service: sudo systemctl start vertex"
	echo "   • Check status: sudo systemctl status vertex"
	echo "   • View logs: sudo journalctl -u vertex -f"

elif [[ "$OS" == "Darwin" ]]; then
	if [[ ! -f "$PLIST_FILE" ]]; then
		echo "📝 Creating launchd plist: $PLIST_FILE"
		cat >"$PLIST_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
 "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.vertex.manager</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/$BINARY_NAME</string>
        <string>-port</string>
        <string>8080</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/vertex.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/vertex.stderr.log</string>
    <key>WorkingDirectory</key>
    <string>$DATA_DIR</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>VERTEX_DATA_DIR</key>
        <string>$DATA_DIR</string>
    </dict>
</dict>
</plist>
EOF
		echo "🧪 Setting permissions for plist"
		chown root:wheel "$PLIST_FILE"
		chmod 644 "$PLIST_FILE"
	else
		echo "✅ Launchd plist already exists"
	fi

	echo "🚀 Loading Vertex launch agent"
	launchctl load "$PLIST_FILE" &>/dev/null || true
	launchctl start com.vertex.manager

	echo "✅ Installation completed successfully on macOS!"
	echo ""
	echo "📋 Next steps:"
	echo "   • Check logs: tail -f /var/log/vertex.stdout.log"
	echo "   • Stop service: sudo launchctl stop com.vertex.manager"
	echo "   • Unload service: sudo launchctl unload $PLIST_FILE"
fi

echo ""
echo "📂 Data directory: $DATA_DIR"
echo "🗄️  Database location: $DATA_DIR/vertex.db"
echo "👤 Service runs as user: $USER"
