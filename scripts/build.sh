#!/bin/bash
set -e

echo "=== Building nonraid-ui ==="

# Build frontend
echo ""
echo "Step 1/3: Installing frontend dependencies..."
cd frontend
npm install

echo ""
echo "Step 2/3: Building frontend..."
npm run build

# Copy frontend build to Go embed location
echo ""
echo "Copying frontend assets to Go embed location..."
cd ..
mkdir -p internal/api/frontend/dist
cp -r frontend/dist/* internal/api/frontend/dist/

# Build Go backend
echo ""
echo "Step 3/3: Building Go backend with embedded frontend..."
go mod download
CGO_ENABLED=0 go build -ldflags="-s -w" -o nonraid-ui ./cmd/server

echo ""
echo "=== Build complete! ==="
echo ""
echo "Binary created: ./nonraid-ui"
echo ""
echo "To run locally:"
echo "  sudo ./nonraid-ui"
echo ""
echo "To install as systemd service:"
echo "  sudo make install"
echo ""
