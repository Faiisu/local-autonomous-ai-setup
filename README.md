# Local Vision LLM & Image Capture Pipeline

This project provides a real-time image processing pipeline that integrates with an Ollama-hosted vision model (`llama3.2-vision`). It captures frames from a camera (or a simulated dummy feed), encodes them as Base64, and sends them to the local LLM to generate descriptive observations.

## Prerequisites

- Go 1.22+
- OpenCV (`libopencv-dev`)
- Ollama (installed and running)

You can automatically set up dependencies on an Ubuntu/Debian system by running:

```bash
make setup
```

## Running the Project

### Development (Dummy Camera Mode)

During development, if you do not have a camera attached to your environment (e.g., WSL2 or a headless server), the pipeline defaults to **"dummy" mode**. This mode bypasses the physical camera and provides a static, valid 1x1 black JPEG frame to the LLM.

```bash
make run
```
*Note: If `VIDEO_DEVICE` is unset, it defaults to `"dummy"`.*

### Production (Physical Camera)

When you are ready to connect a real camera or RTSP stream, provide the device ID or path via the `VIDEO_DEVICE` environment variable:

```bash
# Using /dev/video0
VIDEO_DEVICE="0" make run

# Using a specific file or RTSP stream
VIDEO_DEVICE="rtsp://your-camera-stream" make run
```

## Project Structure

- `cmd/vision-agent/main.go`: Application entrypoint.
- `internal/capture/pipeline.go`: Video capture loop using `gocv`.
- `internal/ollama/client.go`: HTTP client for interacting with the Ollama API.
- `internal/agent/tool.go`: Orchestration integration defining the `VisionTool` schema.

## Available Makefile Commands

- `make setup` - Install system dependencies (OpenCV) and pull AI models.
- `make build` - Compile the Go binary into the `bin/` folder.
- `make run`   - Automatically install all dependencies (if missing), build, and execute the vision pipeline immediately. Go dependencies are localized in `.go/`.
- `make check` - Verify camera hardware and Ollama service status.
- `make clean` - Clean build artifacts.
- `make purge` - **Free up storage space** by completely uninstalling Ollama, removing downloaded AI models, deleting localized Go modules (`.go/`), and removing installed system dependencies (`golang`, `libopencv-dev`).
