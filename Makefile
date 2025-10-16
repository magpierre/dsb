# Makefile for DSB - Delta Sharing Browser

# Variables
APP_NAME = dsb
APP_ID = com.example.dsb
VERSION = 1.0.0
BUILD_DIR = build
ICON = Icon.png

# Go build flags
LDFLAGS = -w -s

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building for current platform..."
	go build -ldflags="$(LDFLAGS)" -o $(APP_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(APP_NAME)
	rm -f *.apk
	rm -f *.aab

# Install fyne command if not present
.PHONY: install-fyne
install-fyne:
	@echo "Installing fyne command..."
	go install fyne.io/fyne/v2/cmd/fyne@latest

# Build Android APK
.PHONY: android
android: install-fyne
	@echo "Building Android APK..."
	fyne package -os android -appID $(APP_ID) -icon $(ICON) -name $(APP_NAME)
	@echo "Android APK created successfully!"

# Build Android AAB (for Google Play)
.PHONY: android-aab
android-aab: install-fyne
	@echo "Building Android AAB..."
	fyne package -os android/aab -appID $(APP_ID) -icon $(ICON) -name $(APP_NAME)
	@echo "Android AAB created successfully!"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin build-ios android

# Build for Linux
.PHONY: build-linux
build-linux: install-fyne
	@echo "Building for Linux..."
	fyne package -os linux -icon $(ICON) -name $(APP_NAME)

# Build for Windows
.PHONY: build-windows
build-windows: install-fyne
	@echo "Building for Windows..."
	fyne package -os windows -icon $(ICON) -name $(APP_NAME)

# Build for macOS
.PHONY: build-darwin
build-darwin: install-fyne
	@echo "Building for macOS..."
	fyne package -os darwin -icon $(ICON) -name $(APP_NAME)

# Build for iOS
.PHONY: build-ios
build-ios: install-fyne
	@echo "Building for iOS..."
	fyne package -os ios -appID $(APP_ID) -icon $(ICON) -name $(APP_NAME)

# Run the application
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	go run .

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Display help
.PHONY: help
help:
	@echo "DSB - Delta Sharing Browser Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build for current platform"
	@echo "  make android        Build Android APK"
	@echo "  make android-aab    Build Android AAB (for Google Play)"
	@echo "  make build-all      Build for all platforms"
	@echo "  make build-linux    Build for Linux"
	@echo "  make build-windows  Build for Windows"
	@echo "  make build-darwin   Build for macOS"
	@echo "  make build-ios      Build for iOS"
	@echo "  make run            Run the application"
	@echo "  make clean          Remove build artifacts"
	@echo "  make deps           Install dependencies"
	@echo "  make fmt            Format code"
	@echo "  make test           Run tests"
	@echo "  make install-fyne   Install fyne command"
	@echo "  make help           Display this help message"
