package main

import (
	"encoding/json"
	_"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"openclaw/internal/agent"
	"openclaw/internal/capture"
	"openclaw/internal/ollama"
	"openclaw/internal/web"
)

func main() {
	// Configuration
	device := os.Getenv("VIDEO_DEVICE")
	if device == "" {
		if _, err := os.Stat("/dev/video0"); err == nil {
			device = "/dev/video0"
		} else {
			device = "dummy"
		}
	}
	width := 640
	height := 480
	captureInterval := 200 * time.Millisecond // 5 FPS for live stream
	ollamaURL := "http://localhost:11434"
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "llava:13b"
	}
	webAddr := ":8080"

	// 0. Setup Web Server
	webServer := web.NewServer()

	log.Printf("Initializing OpenClaw Pipeline with model: %s...", modelName)

	// 1. Setup Capture Pipeline (Live feed enabled)
	pipeline := capture.NewPipeline(device, width, height, captureInterval)
	if err := pipeline.Start(); err != nil {
		log.Fatalf("Failed to start capture pipeline: %v", err)
	}
	defer pipeline.Stop()

	// 2. Setup Ollama Client
	ollamaClient := ollama.NewClient(ollamaURL, modelName)

	// 3. Setup OpenClaw Tool
	openClaw := agent.NewOpenClawTool(pipeline, ollamaClient)

	// 4. Setup Analyze Endpoint (Manual on-demand)
	webServer.AnalyzeHandler = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("OpenClaw manual observation requested...")
		obs, frame, err := openClaw.AnalyzeCameraFrame("Describe what you see. Keep it short and concise.")
		if err != nil {
			log.Printf("Observation error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		webServer.UpdateObservation(frame, obs)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":      "success",
			"image":       frame,
			"description": obs,
		})
	}

	// Provide the live feed to the web server
	http.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		frame := pipeline.GetLastFrame()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"image": frame})
	})

	go func() {
		if err := webServer.Start(webAddr); err != nil {
			log.Fatalf("Web server failed: %v", err)
		}
	}()

	log.Printf("OpenClaw is running. Web UI available at http://localhost%s", webAddr)
	log.Println("Press Ctrl+C to exit.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down OpenClaw...")
}
