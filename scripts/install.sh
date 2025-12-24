#!/bin/bash
set -e

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

echo "=== Installing nonraid-ui ==="

# Check if binary exists
if [ ! -f "./nonraid-ui" ]; then
  echo "Error: nonraid-ui binary not found. Run 'make build' first."
  exit 1
fi

# Install binary
echo "Installing binary to /usr/local/bin/..."
install -m 755 nonraid-ui /usr/local/bin/

# Install systemd service
echo "Installing systemd service..."
install -m 644 scripts/nonraid-ui.service /etc/systemd/system/
systemctl daemon-reload

echo ""
echo "=== Installation complete! ==="
echo ""
echo "To start the service:"
echo "  sudo systemctl start nonraid-ui"
echo ""
echo "To enable on boot:"
echo "  sudo systemctl enable nonraid-ui"
echo ""
echo "To check status:"
echo "  sudo systemctl status nonraid-ui"
echo ""
echo "The UI will be available at http://YOUR_SERVER_IP:3000"
echo ""
