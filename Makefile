# Project: Local Vision LLM & Image Capture Pipeline

BINARY_NAME=vision-agent
MODEL=llava

# Keep Go dependencies localized to the project folder
export GOPATH := $(PWD)/.go
export PATH := $(GOPATH)/bin:$(PATH)

.PHONY: all setup build run check clean pull-models purge

all: build

setup:
	@echo "Installing system dependencies..."
	sudo apt-get update && sudo apt-get install -y golang libopencv-dev pkg-config curl
	@echo "Checking if Ollama is installed..."
	@if ! command -v ollama > /dev/null; then \
		echo "Installing Ollama..."; \
		curl -fsSL https://ollama.com/install.sh | sh; \
	fi
	$(MAKE) pull-models

pull-models:
	@echo "Pulling vision model $(MODEL)..."
	ollama pull $(MODEL)

build:
	@echo "Building Go binary..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/vision-agent

run: setup build
	./bin/$(BINARY_NAME)

check:
	@echo "Checking camera hardware..."
	@ls /dev/video* || echo "Warning: No camera found at /dev/video*"
	@echo "Checking Ollama status..."
	@ollama list | grep $(MODEL) || echo "Warning: Model $(MODEL) not found. Run 'make pull-models'"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean || true

purge: clean
	@echo "Removing downloaded Ollama models..."
	-ollama rm $(MODEL)
	-ollama rm llama3.2-vision
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
