# nonraid-ui

A modern, beautiful WebUI for managing [nonraid](https://github.com/qvr/nonraid) storage arrays. Built with Go and SvelteKit for minimal dependencies and maximum performance.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Zero Runtime Dependencies**: Compiles to a single static binary
- **Real-time Monitoring**: Auto-refreshing status every 5 seconds
- **Modern UI**: Beautiful, responsive interface built with TailwindCSS
- **Resync Progress**: Live tracking of array resync operations
- **Disk Management**: Comprehensive disk status and health monitoring
- **Dark Mode Support**: Easy on the eyes, day or night

## Screenshots

The UI displays:
- Array health status and statistics
- Active resync progress with speed and ETA
- Detailed disk table with device info, filesystem usage, and health status
- Parity configuration and capacity information

## Architecture

```
┌─────────────────┐
│   Web Browser   │
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────┐
│  Go Backend     │ ← Single binary with embedded frontend
│  (Port 3000)    │   Runs as systemd service with root privileges
└────────┬────────┘
         │ executes
         ▼
   nmdctl status -o json
```

## Prerequisites

- Go 1.21 or later (for building)
- Node.js 18+ (for building frontend)
- nonraid installed and configured
- Root access (required to run \`nmdctl\`)

## Quick Start

### Building from Source

```bash
# Clone the repository
git clone https://github.com/jrarseneau/nonraid-ui.git
cd nonraid-ui

# Build (builds both frontend and backend)
make build

# Or use the build script
chmod +x scripts/build.sh
./scripts/build.sh
```

This creates a single binary: \`./nonraid-ui\`

### Installation as Systemd Service

```bash
# Install the service (requires root)
sudo make install

# Start the service
sudo systemctl start nonraid-ui

# Enable on boot
sudo systemctl enable nonraid-ui

# Check status
sudo systemctl status nonraid-ui
```

The UI will be available at \`http://YOUR_SERVER_IP:3000\`

### Running Manually

```bash
# Run with default settings (port 3000, all interfaces)
sudo ./nonraid-ui

# Custom port and host
sudo ./nonraid-ui --port 8080 --host 127.0.0.1
```

## Development

### Development Mode

Run the frontend and backend separately for hot-reloading:

```bash
# Terminal 1: Start frontend dev server (port 5173)
cd frontend
npm install
npm run dev

# Terminal 2: Start backend (port 3000)
go run ./cmd/server
```

The frontend dev server will proxy API requests to the backend.

### Project Structure

```
nonraid-ui/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── nmdctl/
│   │   ├── types.go             # Data structures for nmdctl output
│   │   └── client.go            # nmdctl command executor
│   └── api/
│       ├── handlers.go          # HTTP handlers and routing
│       └── frontend/dist/       # Embedded frontend (generated)
├── frontend/                    # SvelteKit frontend
│   ├── src/
│   │   ├── routes/              # SvelteKit routes
│   │   ├── lib/
│   │   │   ├── components/      # Svelte components
│   │   │   ├── api.ts           # API client
│   │   │   ├── types.ts         # TypeScript types
│   │   │   └── utils.ts         # Utility functions
│   │   ├── app.html
│   │   └── app.css
│   └── package.json
├── scripts/
│   ├── build.sh                 # Build script
│   ├── install.sh               # Installation script
│   └── nonraid-ui.service       # Systemd service file
├── Makefile                     # Build automation
└── README.md
```

## API Endpoints

- \`GET /api/status\` - Returns current array status (JSON)
- \`GET /\` - Serves the web interface

## Configuration

### Systemd Service

The systemd service runs as root (required for \`nmdctl\`) and includes security hardening:
- \`PrivateTmp=true\` - Isolated /tmp
- \`ProtectSystem=strict\` - Read-only system directories
- \`ProtectHome=true\` - No access to user home directories
- \`NoNewPrivileges=true\` - Cannot escalate privileges

Edit \`/etc/systemd/system/nonraid-ui.service\` to customize:
- Port (default: 3000)
- Host binding (default: 0.0.0.0)

After editing, reload and restart:
```bash
sudo systemctl daemon-reload
sudo systemctl restart nonraid-ui
```

## Building for Production

The build process:
1. Installs frontend dependencies
2. Builds the SvelteKit app (optimized, minified)
3. Embeds the frontend into the Go binary using \`embed\`
4. Compiles a static Go binary (no CGO, portable)

Result: A single ~10-20MB binary with everything included.

## Uninstalling

```bash
sudo make uninstall
```

This stops the service, removes the binary, and cleans up systemd.

## Troubleshooting

### Service won't start
```bash
# Check service logs
sudo journalctl -u nonraid-ui -f

# Verify nmdctl works
sudo nmdctl status -o json

# Check permissions
sudo systemctl status nonraid-ui
```

### Can't access the UI
- Check firewall: \`sudo ufw allow 3000\`
- Verify service is running: \`sudo systemctl status nonraid-ui\`
- Check the bind address in the service file

### Build fails
- Ensure Go 1.21+ is installed: \`go version\`
- Ensure Node.js 18+ is installed: \`node --version\`
- Clean and rebuild: \`make clean && make build\`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Acknowledgments

- [nonraid](https://github.com/qvr/nonraid) - The open-source alternative to Unraid
- Built with [Go](https://golang.org/), [SvelteKit](https://kit.svelte.dev/), and [TailwindCSS](https://tailwindcss.com/)

## Roadmap

- [x] Read-only array status display
- [ ] Array management controls (start/stop)
- [ ] Disk management (add/remove disks)
- [ ] Resync controls (pause/resume)
- [ ] Email notifications
- [ ] Historical metrics and graphs
- [ ] Docker support
- [ ] Authentication
