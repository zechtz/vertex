#!/bin/bash

# Build script for NeST Service Manager
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Version information
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo -e "${BLUE}🏗️  Building NeST Service Manager${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Commit:  ${COMMIT}${NC}"
echo -e "${YELLOW}Date:    ${DATE}${NC}"
echo

# Build frontend
echo -e "${BLUE}📦 Building frontend...${NC}"
cd web
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    npm install
fi
npm run build
cd ..
echo -e "${GREEN}✅ Frontend built successfully${NC}"
echo

# Build backend
echo -e "${BLUE}🔧 Building backend...${NC}"
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Default build
CGO_ENABLED=1 go build -ldflags="${LDFLAGS}" -o nest-up

echo -e "${GREEN}✅ Backend built successfully${NC}"
echo

# Show binary info
BINARY_SIZE=$(ls -lh nest-up | awk '{print $5}')
echo -e "${GREEN}📊 Binary created: nest-up (${BINARY_SIZE})${NC}"

# Test version
echo -e "${BLUE}🧪 Testing binary...${NC}"
./nest-up -version
echo

echo -e "${GREEN}🎉 Build complete!${NC}"
echo -e "${YELLOW}To run: ./nest-up${NC}"
echo -e "${YELLOW}Web interface: http://localhost:8080${NC}"