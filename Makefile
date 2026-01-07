.PHONY: all build frontend backend clean install uninstall dev

all: build

# Build everything
build: frontend backend

# Build the frontend
frontend:
	@echo "Building frontend..."
	cd frontend && npm install
	@echo "Running svelte-kit sync..."
	cd frontend && npx svelte-kit sync
	@echo "Building frontend assets..."
	cd frontend && npm run build
	@echo "Copying frontend build to internal/api/frontend/dist..."
	@mkdir -p internal/api/frontend/dist
	@cp -r frontend/dist/* internal/api/frontend/dist/

# Build the Go backend with embedded frontend
backend:
	@echo "Building Go backend..."
	go mod download
	CGO_ENABLED=0 go build -ldflags="-s -w" -o nonraid-ui ./cmd/server

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf frontend/.svelte-kit
	rm -rf internal/api/frontend/dist
	rm -f nonraid-ui
	go clean

# Install the binary and systemd service
install: build
	@echo "Installing nonraid-ui..."
	install -m 755 nonraid-ui /usr/local/bin/
	install -m 644 scripts/nonraid-ui.service /etc/systemd/system/
	systemctl daemon-reload
	@echo ""
	@echo "Installation complete!"
	@echo "To start the service:"
	@echo "  sudo systemctl start nonraid-ui"
	@echo "  sudo systemctl enable nonraid-ui"

# Uninstall
uninstall:
	@echo "Uninstalling nonraid-ui..."
	systemctl stop nonraid-ui || true
	systemctl disable nonraid-ui || true
	rm -f /usr/local/bin/nonraid-ui
	rm -f /etc/systemd/system/nonraid-ui.service
	systemctl daemon-reload

# Development mode
dev:
	@echo "Starting development mode..."
	@echo "Frontend will run on http://localhost:5173"
	@echo "Backend will run on http://localhost:3000"
	@cd frontend && npm run dev &
	@go run ./cmd/server
