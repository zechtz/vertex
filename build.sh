#!/bin/bash

# Cross-platform build script for Vertex
set -e

VERSION=${VERSION:-"dev"}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}

BUILD_FLAGS="-ldflags=-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

echo "🏗️ Building Vertex Service Manager"
echo "   Version: $VERSION"
echo "   Commit: $COMMIT"
echo "   Date: $DATE"
echo ""

# Build for current platform
echo "📦 Building for current platform..."
go build $BUILD_FLAGS -o vertex .

# Build for all platforms
echo "🌍 Building for all platforms..."

# Windows
echo "  🪟 Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build $BUILD_FLAGS -o vertex-windows-amd64.exe . 2>/dev/null || {
    echo "    ⚠️ Cross-compilation for Windows failed (CGO/SQLite dependency)"
    echo "    ℹ️ To build for Windows, run this on a Windows machine:"
    echo "       go build $BUILD_FLAGS -o vertex.exe ."
}

# Linux
echo "  🐧 Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build $BUILD_FLAGS -o vertex-linux-amd64 . 2>/dev/null || {
    echo "    ⚠️ Cross-compilation for Linux failed"
}

# macOS
echo "  🍎 Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build $BUILD_FLAGS -o vertex-darwin-amd64 . 2>/dev/null || {
    echo "    ⚠️ Cross-compilation for macOS amd64 failed"
}

echo "  🍎 Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build $BUILD_FLAGS -o vertex-darwin-arm64 . 2>/dev/null || {
    echo "    ⚠️ Cross-compilation for macOS arm64 failed"
}

echo ""
echo "✅ Build completed!"
echo ""
echo "📁 Generated files:"
ls -la vertex* 2>/dev/null | grep -E "(vertex-|vertex$)" || echo "   • vertex (current platform)"

echo ""
echo "🚀 Installation:"
echo "   • Current platform: Use the 'vertex' binary"
echo "   • Linux: sudo ./install.sh"
echo "   • Windows: powershell -ExecutionPolicy Bypass -File install.ps1"
echo ""
echo "📖 See INSTALLATION.md for detailed instructions"