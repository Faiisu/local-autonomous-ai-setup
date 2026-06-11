package agent

import (
	"fmt"
	"vision-detect/internal/capture"
	"vision-detect/internal/ollama"
)

type VisionTool struct {
	pipeline *capture.Pipeline
	ollama   *ollama.Client
}

func NewVisionTool(p *capture.Pipeline, c *ollama.Client) *VisionTool {
	return &VisionTool{
		pipeline: p,
		ollama:   c,
	}
}

func (t *VisionTool) AnalyzeCameraFrame(prompt string) (string, error) {
	frame := t.pipeline.GetLastFrame()
	if frame == "" {
		return "", fmt.Errorf("no camera frame available yet")
	}

	if prompt == "" {
		prompt = "Describe what you see in this image briefly and concisely."
	}

	return t.ollama.GenerateObservation(prompt, frame)
}

// ToolSchema returns the JSON schema for this tool to be used by an orchestrator
func (t *VisionTool) ToolSchema() string {
	return `{
		"name": "analyze_camera_frame",
		"description": "Captures the current frame from the local camera and analyzes it using a vision LLM.",
		"parameters": {
			"type": "object",
			"properties": {
				"prompt": {
					"type": "string",
					"description": "Specific question or instruction for the vision model (e.g., 'Is there anyone in the room?')."
				}
			}
		}
	}`
}
