# Project: Local Vision LLM & Image Capture Pipeline

BINARY_NAME=vision-agent
MODEL=llava

.PHONY: all setup build run check clean pull-models

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

run: build
	./bin/$(BINARY_NAME)

check:
	@echo "Checking camera hardware..."
	@ls /dev/video* || echo "Warning: No camera found at /dev/video*"
	@echo "Checking Ollama status..."
	@ollama list | grep $(MODEL) || echo "Warning: Model $(MODEL) not found. Run 'make pull-models'"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
