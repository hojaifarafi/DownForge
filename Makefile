.PHONY: all build build-backend build-frontend run clean test

APP_NAME = downforge
BUILD_DIR = .

all: build

# Build both backend and frontend
build: build-frontend build-backend

# Build React UI into web/dist
build-frontend:
	@echo "🔨 Building React UI..."
	cd web && npm run build

# Build Go binary
build-backend:
	@echo "🔨 Building Go backend binary..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/downforge

# Run the complete application
run: build
	@echo "🚀 Starting DownForge..."
	./$(APP_NAME) -port 8080

# Run backend with live Vite dev server in parallel
dev:
	@echo "⚡ Starting development servers..."
	@echo "Run backend: go run ./cmd/downforge"
	@echo "Run frontend: cd web && npm run dev"

clean:
	@echo "🧹 Cleaning artifacts..."
	rm -f $(BUILD_DIR)/$(APP_NAME)
	rm -rf web/dist

test:
	@echo "🧪 Running backend tests..."
	go test -v ./...
