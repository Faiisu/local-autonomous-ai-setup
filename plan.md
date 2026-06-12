This is the English version of the project plan, designed to be flexible. It
provides a clear roadmap while allowing the LLM to adjust implementation details
(like CGO flags, OpenCV paths, or specific Ollama API changes) based on
real-time errors or environment constraints.

Project Plan: Local Vision LLM & Image Capture Pipeline

1. Project Overview

A real-time image processing system using a Local Vision LLM running on
Edge/Industrial PC (IPC). The system captures frames via Go, processes them, and
sends them to a local AI engine to generate "Observations" for a higher-level
Agent Orchestrator.

Tech Stack

  - Language: Go (Golang) with gocv
  - AI Engine: Ollama (Vision Models like llama3.2-vision)
  - OS/Hardware: Linux-based IPC, USB/RTSP Camera
  - Orchestration: Makefile-driven deployment

2. Development Phases

Phase 1: Environment & Dependency Automation

  - Goal: Create a "one-command" setup for the IPC.
  - Key Tasks:
      - Automate installation of OpenCV C++ shared libraries (libopencv-dev).
      - Verify video device access (/dev/videoX).
      - Ensure Ollama is installed and the specific vision model is pulled.
      - Success Metric: make setup completes without errors and ollama list
        shows the model.

Phase 2: High-Performance Image Pipeline (Go)

  - Goal: Efficiently capture and prepare images for the LLM.
  - Key Tasks:
      - Implement gocv capture loop with configurable frame rates (e.g., 1 frame
        every 5 seconds).
      - Dynamic Preprocessing: Resize/Compress images (e.g., 640x480) to
        optimize inference speed and reduce Base64 payload size.
      - Memory Safety: Implement strict resource management using defer and
        Close() to prevent CGO memory leaks.
      - Success Metric: Go application logs successful frame captures and Base64
        string generation.

Phase 3: Local AI Inference Integration

  - Goal: Connect the image pipeline to the Vision LLM.
  - Key Tasks:
      - Develop an HTTP client to communicate with Ollama’s REST API
        (/api/generate).
      - Handle JSON payloads containing Base64 images and custom prompts.
      - Implement error handling for inference timeouts or "model busy" states.
      - Success Metric: System receives descriptive text (JSON response) based
        on the live camera feed.

Phase 4: Agent Orchestrator Tooling

  - Goal: Expose the pipeline as a "Tool" for an AI Agent.
  - Key Tasks:
      - Define the analyze_camera_frame Tool Schema.
      - Create a clean interface for the ReAct loop to call the vision service.
      - Success Metric: An Agent can programmatically request a "visual
        observation" and receive a text summary.

Phase 5: Production Readiness & Optimization

  - Goal: Ensure long-term stability on IPC hardware.
  - Key Tasks:
      - Implement graceful shutdown (release camera handle on SIGTERM).
      - Fine-tune hardware acceleration settings (GPU/iGPU utilization).
      - Log Latency metrics (Capture time vs. Inference time).

Phase 6: Web UI Integration

  - Goal: Provide a remote dashboard for real-time monitoring.
  - Key Tasks:
      - Implement a Go-based web server with WebSocket support.
      - Stream terminal logs to the browser in real-time.
      - Display the latest captured image and AI-generated description.
      - Success Metric: Accessing http://localhost:8080 shows logs and vision data.

3. Project Management (Makefile Interface)

The project must be managed entirely through a Makefile located in the root
directory for ease of use.

| Command      | Description                                              |
| :----------- | :------------------------------------------------------- |
| `make setup` | Install system dependencies (OpenCV) and pull AI models. |
| `make build` | Compile the Go binary into the `bin/` folder.            |
| `make run`   | Build and execute the vision pipeline immediately.       |
| `make check` | Verify camera hardware and Ollama service status.        |
| `make clean` | Clean build artifacts and clear temporary memory.        |

4. Implementation Notes for the LLM

  - Flexibility: If gocv encounters header path issues, adjust CGO_CFLAGS and
    CGO_LDFLAGS dynamically in the Makefile or .env.
  - Resilience: The capture loop should not crash the entire program if the
    camera is momentarily unplugged; it should attempt to reconnect.
  - Context: The prompt sent to the LLM should be concise to keep the
    "Observation" data useful for the Agent.

5. Development Remarks

  - **Camera Setup**: During initial development, the environment uses a "dummy" video device to bypass hardware requirements. The video device path is provided via the `VIDEO_DEVICE` environment variable (defaulting to `"dummy"`). In production, this will be set to a valid device path like `/dev/video0` or a stream URL.
