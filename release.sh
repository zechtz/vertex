#!/bin/bash

# Release script for vertex
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if version is provided
if [ -z "$1" ]; then
	echo -e "${RED}❌ Usage: $0 <version>${NC}"
	echo -e "${YELLOW}Example: $0 v1.0.0${NC}"
	exit 1
fi

VERSION=$1
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo -e "${BLUE}🚀 Building release ${VERSION}${NC}"
echo

# Create release directory
mkdir -p release
cd release
rm -rf *

# Build frontend once
echo -e "${BLUE}📦 Building frontend...${NC}"
cd ../web
npm ci
npm run build
cd ../release

# Build for multiple platforms
PLATFORMS=(
	"linux/amd64"
	"linux/arm64"
	"darwin/amd64"
	"darwin/arm64"
	"windows/amd64"
)

LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

for platform in "${PLATFORMS[@]}"; do
	GOOS=${platform%/*}
	GOARCH=${platform#*/}

	echo -e "${BLUE}🔧 Building for ${GOOS}/${GOARCH}...${NC}"

	BINARY_NAME="vertex-${GOOS}-${GOARCH}"
	if [ "$GOOS" = "windows" ]; then
		BINARY_NAME="${BINARY_NAME}.exe"
	fi

	# Set CC for Windows cross-compilation
	if [ "$GOOS" = "windows" ]; then
		export CC=x86_64-w64-mingw32-gcc
	else
		unset CC
	fi

	# Build
	cd ..
	CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build \
		-ldflags="${LDFLAGS}" \
		-o "release/${BINARY_NAME}"
	cd release

	# Create checksum
	sha256sum "${BINARY_NAME}" >"${BINARY_NAME}.sha256"

	echo -e "${GREEN}✅ Built ${BINARY_NAME}${NC}"
done

# Create checksums file
echo -e "${BLUE}📋 Creating checksums...${NC}"
cat *.sha256 >checksums.txt

# Show results
echo
echo -e "${GREEN}🎉 Release ${VERSION} built successfully!${NC}"
echo -e "${YELLOW}Release files:${NC}"
ls -lh vertex-*
echo
echo -e "${YELLOW}Checksums:${NC}"
cat checksums.txt
echo

echo -e "${BLUE}📝 Next steps:${NC}"
echo -e "${YELLOW}1. Test the binaries${NC}"
echo -e "${YELLOW}2. Create git tag: git tag ${VERSION}${NC}"
echo -e "${YELLOW}3. Push tag: git push origin ${VERSION}${NC}"
echo -e "${YELLOW}4. CI will automatically create GitHub/GitLab releases${NC}"
