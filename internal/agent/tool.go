package agent

import (
	"fmt"
	"openclaw/internal/capture"
	"openclaw/internal/ollama"
)

type OpenClawTool struct {
	pipeline *capture.Pipeline
	ollama   *ollama.Client
}

func NewOpenClawTool(p *capture.Pipeline, c *ollama.Client) *OpenClawTool {
	return &OpenClawTool{
		pipeline: p,
		ollama:   c,
	}
}

func (t *OpenClawTool) AnalyzeCameraFrame(prompt string) (string, string, error) {
	frame := t.pipeline.GetLastFrame()
	if frame == "" {
		return "", "", fmt.Errorf("no camera frame available yet")
	}

	if prompt == "" {
		prompt = "Describe what you see in this image briefly and concisely."
	}

	obs, err := t.ollama.GenerateObservation(prompt, frame)
	return obs, frame, err
}

// ToolSchema returns the JSON schema for this tool to be used by an orchestrator
func (t *OpenClawTool) ToolSchema() string {
	return `{
		"name": "openclaw_analyze",
		"description": "Captures the current frame from the local camera and analyzes it using OpenClaw vision intelligence.",
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
