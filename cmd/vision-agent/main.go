package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vision-detect/internal/agent"
	"vision-detect/internal/capture"
	"vision-detect/internal/ollama"
)

func main() {
	// Configuration
	device := os.Getenv("VIDEO_DEVICE")
	if device == "" {
		// Auto-detect first video device, otherwise fallback to dummy
		if _, err := os.Stat("/dev/video0"); err == nil {
			device = "/dev/video0"
		} else {
			device = "dummy"
		}
	}
	width := 640
	height := 480
	captureInterval := 5 * time.Second
	ollamaURL := "http://localhost:11434"
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "llava:13b"
	}

	log.Printf("Initializing Vision Agent Pipeline with model: %s...", modelName)

	// 1. Setup Capture Pipeline
	pipeline := capture.NewPipeline(device, width, height, captureInterval)
	if err := pipeline.Start(); err != nil {
		log.Fatalf("Failed to start capture pipeline: %v", err)
	}
	defer pipeline.Stop()

	// 2. Setup Ollama Client
	ollamaClient := ollama.NewClient(ollamaURL, modelName)

	// 3. Setup Agent Tool
	visionTool := agent.NewVisionTool(pipeline, ollamaClient)

	// 4. Main loop for periodic observation (Phase 5: Latency Metrics)
	go func() {
		for {
			time.Sleep(10 * time.Second)

			log.Println("Requesting automated observation...")
			start := time.Now()

			obs, err := visionTool.AnalyzeCameraFrame("Briefly describe what you see.")
			if err != nil {
				log.Printf("Observation error: %v", err)
				continue
			}

			latency := time.Since(start)
			log.Printf("Observation [%v]: %s", latency, obs)

			// Update HUD
			pipeline.SetOverlayText(fmt.Sprintf("[%v] %s", latency.Truncate(time.Millisecond), obs))
		}
	}()

	log.Println("Vision Agent is running. Press Ctrl+C to exit.")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Vision Agent...")
}
