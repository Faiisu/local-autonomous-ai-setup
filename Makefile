# Project: Local Vision LLM & Image Capture Pipeline

BINARY_NAME=openclaw
MODEL=llava:13b

# Keep Go dependencies localized to the project folder
export GOPATH := $(PWD)/.go
export PATH := $(GOPATH)/bin:$(PATH)

.PHONY: all setup build run run-pw check clean pull-models purge

all: build

setup:
	@echo "Installing system dependencies..."
	sudo apt-get update && sudo apt-get install -y golang libopencv-dev pkg-config curl pipewire-v4l2
	@echo "Fixing OpenCV pkg-config..."
	@if grep -q "prefix=/usr/local" /usr/lib/pkgconfig/opencv4.pc; then \
		sudo sed -i 's|prefix=/usr/local|prefix=/usr|g' /usr/lib/pkgconfig/opencv4.pc; \
		sudo sed -i 's|libdir=$${exec_prefix}/lib|libdir=$${exec_prefix}/lib/aarch64-linux-gnu|g' /usr/lib/pkgconfig/opencv4.pc; \
	fi
	@echo "Checking if Ollama is installed..."
	@if ! command -v ollama > /dev/null; then \
		echo "Installing Ollama..."; \
		curl -fsSL https://ollama.com/install.sh | sh; \
	fi
	go get gocv.io/x/gocv@v0.35.0
	go mod tidy
	$(MAKE) pull-models

pull-models:
	@echo "Pulling vision model $(MODEL)..."
	ollama pull $(MODEL)

build:
	@echo "Building Go binary..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/openclaw

run: build
	@echo "Running normally..."
	./bin/$(BINARY_NAME)

run-pw: build
	@echo "Running with PipeWire (pw-v4l2)..."
	@if command -v pw-v4l2 > /dev/null; then \
		pw-v4l2 ./bin/$(BINARY_NAME); \
	else \
		echo "Error: pw-v4l2 not found. Run 'make setup' to install pipewire-v4l2."; \
		exit 1; \
	fi

check:
	@echo "Checking camera hardware..."
	@ls /dev/video* || echo "Warning: No camera found at /dev/video*"
	@echo "Checking PipeWire status..."
	@pactl info | grep "Server Name" || echo "Warning: PipeWire/PulseAudio not detected"
	@echo "Checking Ollama status..."
	@ollama list | grep $(MODEL) || echo "Warning: Model $(MODEL) not found. Run 'make pull-models'"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean || true

purge: clean
	@echo "Removing downloaded Ollama models..."
	-ollama rm $(MODEL)
	@echo "Stopping and uninstalling Ollama..."
	-sudo systemctl stop ollama
	-sudo systemctl disable ollama
	-sudo rm -rf /usr/local/bin/ollama /usr/local/lib/ollama /etc/systemd/system/ollama.service
	-sudo rm -rf /usr/share/ollama
	-sudo userdel ollama
	-sudo groupdel ollama
	@echo "Removing localized Go dependencies..."
	rm -rf .go/
	@echo "Removing system dependencies (golang, libopencv-dev, pkg-config)..."
	sudo apt-get remove --purge -y golang libopencv-dev pkg-config
	sudo apt-get autoremove -y
	@echo "Project and environment fully purged."
