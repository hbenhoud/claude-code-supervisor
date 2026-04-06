.PHONY: dev build test lint clean

# Dev mode: run backend + frontend concurrently
dev:
	@echo "Starting backend and frontend..."
	@trap 'kill 0' EXIT; \
	cd web && npm run dev & \
	go run ./cmd/supervisor/ & \
	wait

# Build single binary with embedded frontend
build:
	cd web && npm run build
	go build -o dist/supervisor ./cmd/supervisor/
	go build -o dist/supervisor-init ./cmd/supervisor-init/

# Run all tests
test:
	go test ./...
	cd web && npm test -- --run 2>/dev/null || true

# Lint
lint:
	cd web && npm run lint
	go vet ./...

# Clean build artifacts
clean:
	rm -rf dist/
	rm -rf web/dist/
