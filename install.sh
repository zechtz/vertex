#!/bin/bash

# Vertex Installation Script
set -e

BINARY_NAME="vertex"
INSTALL_DIR="$HOME/.local/bin"
DATA_DIR="$HOME/.vertex"
USER="vertex"
GROUP="vertex"
PLIST_FILE="$HOME/Library/LaunchAgents/com.vertex.manager.plist"

echo "🚀 Installing Vertex Service Manager..."

# Running as user (no root required for LaunchAgent approach)
echo "👤 Running as user: $(whoami)"

OS="$(uname)"

if [[ "$OS" == "Linux" ]]; then
	echo "🐧 Linux detected: setting up to run as current user"
	USER=$(whoami)
	GROUP=$(id -gn)
	
	echo "👤 Service will run as current user: $USER"
	
	# No need to create system user - we'll run as current user
	# No need for sudo permissions - current user already has access to their files
	# Java will be detected from user's environment
elif [[ "$OS" == "Darwin" ]]; then
	echo "🍎 macOS detected: setting up to run as current user"
	USER=$(whoami)
	GROUP="staff"
	
	echo "👤 Service will run as current user: $USER"
	
	# No need to create vertex user - we'll run as current user
	# No need for sudo permissions - current user already has access to their files
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
chmod 755 "$DATA_DIR"

# Database files will be owned by current user automatically

# Check for existing installation and stop service if running
if [[ "$OS" == "Linux" ]]; then
	if systemctl --user is-active --quiet vertex 2>/dev/null; then
		echo "🛑 Stopping existing vertex service"
		systemctl --user stop vertex
	fi
elif [[ "$OS" == "Darwin" ]]; then
	if launchctl list | grep -q "com.vertex.manager"; then
		echo "🛑 Stopping existing vertex service"
		launchctl stop com.vertex.manager &>/dev/null || true
		launchctl unload "$PLIST_FILE" &>/dev/null || true
	fi
fi

# Create local bin directory and copy binary
mkdir -p "$INSTALL_DIR"

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
	# Create user systemd service directory
	mkdir -p "$HOME/.config/systemd/user"
	
	# Create systemd user service file
	SERVICE_FILE="$HOME/.config/systemd/user/vertex.service"
	if [[ ! -f "$SERVICE_FILE" ]]; then
		echo "🔧 Creating systemd user service file"
		cat >"$SERVICE_FILE" <<EOF
[Unit]
Description=Vertex Service Manager
Documentation=https://github.com/zechtz/vertex
After=network.target
Wants=network.target

[Service]
Type=simple
WorkingDirectory=$DATA_DIR
Environment=VERTEX_DATA_DIR=$DATA_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME -port 8080
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
EOF
		chmod 644 "$SERVICE_FILE"
		echo "🔄 Reloading systemd user daemon and enabling vertex service"
		systemctl --user daemon-reload
		systemctl --user enable vertex
	else
		echo "✅ Systemd user service file already exists"
		echo "🔄 Reloading systemd user daemon"
		systemctl --user daemon-reload
	fi

	echo "✅ Installation completed successfully on Linux!"
	echo ""
	echo "📋 Next steps:"
	echo "   • Start the service: systemctl --user start vertex"
	echo "   • Check status: systemctl --user status vertex"
	echo "   • View logs: journalctl --user -u vertex -f"

elif [[ "$OS" == "Darwin" ]]; then
	# Check for system-wide Java installation
	echo "☕ Checking for system-wide Java installation..."
	SYSTEM_JAVA_PATHS=(
		"/opt/homebrew/opt/openjdk"
		"/usr/local/opt/openjdk"
		"/Library/Java/JavaVirtualMachines/openjdk.jdk/Contents/Home"
		"/System/Library/Java/JavaVirtualMachines/1.8.0.jdk/Contents/Home"
	)
	
	FOUND_SYSTEM_JAVA=""
	for java_path in "${SYSTEM_JAVA_PATHS[@]}"; do
		if [[ -d "$java_path" && -x "$java_path/bin/java" ]]; then
			FOUND_SYSTEM_JAVA="$java_path"
			echo "✅ Found system Java: $java_path"
			break
		fi
	done
	
	if [[ -z "$FOUND_SYSTEM_JAVA" ]]; then
		echo "⚠️  No system-wide Java found. For best compatibility, install Java system-wide:"
		echo "   macOS: brew install openjdk"
		echo "   Then run: sudo $(basename "$0") again"
		echo ""
		echo "   Or install from: https://adoptopenjdk.net/"
		echo ""
		echo "⚡ Continuing anyway - service will attempt runtime detection..."
	fi

	# Create LaunchAgents directory if it doesn't exist
	mkdir -p "$HOME/Library/LaunchAgents"
	
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
    <key>UserName</key>
    <string>$USER</string>
    <key>GroupName</key>
    <string>$GROUP</string>
    <key>StandardOutPath</key>
    <string>$HOME/.vertex/vertex.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>$HOME/.vertex/vertex.stderr.log</string>
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
		chmod 644 "$PLIST_FILE"
	else
		echo "✅ Launchd plist already exists"
	fi

	# Create log files and set permissions
	echo "📝 Setting up log files"
	touch "$HOME/.vertex/vertex.stdout.log" "$HOME/.vertex/vertex.stderr.log" 
	chmod 644 "$HOME/.vertex/vertex.stdout.log" "$HOME/.vertex/vertex.stderr.log"

	# Java will be detected automatically when the service starts
	echo "☕ Java environment will be detected automatically"

	echo "🚀 Loading Vertex launch agent"
	launchctl load "$PLIST_FILE" &>/dev/null || true
	launchctl start com.vertex.manager

	echo "✅ Installation completed successfully on macOS!"
	echo ""
	echo "📋 Next steps:"
	echo "   • Check logs: tail -f $HOME/.vertex/vertex.stdout.log"
	echo "   • Stop service: launchctl stop com.vertex.manager"
	echo "   • Unload service: launchctl unload $PLIST_FILE"
fi

echo ""
echo "📂 Data directory: $DATA_DIR"
echo "🗄️  Database location: $DATA_DIR/vertex.db"
echo "👤 Service runs as user: $USER"
