package capture

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type Pipeline struct {
	Device   string
	Width    int
	Height   int
	Interval time.Duration

	mu          sync.Mutex
	lastFrame   string
	overlayText string
	isRunning   bool
	stopSignal  chan struct{}
}

func NewPipeline(device string, width, height int, interval time.Duration) *Pipeline {
	return &Pipeline{
		Device:     device,
		Width:      width,
		Height:     height,
		Interval:   interval,
		stopSignal: make(chan struct{}),
	}
}

func (p *Pipeline) SetOverlayText(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.overlayText = text
}

func (p *Pipeline) Start() error {
	if p.Device == "dummy" {
		log.Println("Running in DUMMY mode. Simulating camera capture.")
		p.mu.Lock()
		p.isRunning = true
		// A 1x1 black JPEG base64 string to prevent crashes when LLM expects an image
		p.lastFrame = "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////wgALCAABAAEBAREA/8QAFBABAAAAAAAAAAAAAAAAAAAAAP/aAAgBAQABPxA="
		p.mu.Unlock()
		return nil
	}

	var dev interface{} = p.Device
	if id, err := strconv.Atoi(p.Device); err == nil {
		dev = id
	}

	webcam, err := gocv.OpenVideoCapture(dev)
	if err != nil {
		return fmt.Errorf("failed to open video capture: %w", err)
	}

	p.mu.Lock()
	p.isRunning = true
	p.mu.Unlock()

	go p.captureLoop(webcam, dev)
	return nil
}

func (p *Pipeline) Stop() {
	p.mu.Lock()
	if p.isRunning {
		if p.stopSignal != nil {
			close(p.stopSignal)
		}
		p.isRunning = false
	}
	p.mu.Unlock()
}

func (p *Pipeline) GetLastFrame() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastFrame
}

func (p *Pipeline) captureLoop(webcam *gocv.VideoCapture, dev interface{}) {
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	resized := gocv.NewMat()
	defer resized.Close()

	// Check if we have a display environment
	hasDisplay := os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	var window *gocv.Window
	if hasDisplay {
		window = gocv.NewWindow("Vision Agent - Live Feed")
		defer window.Close()
		log.Printf("Starting live capture and visualization on device %v", dev)
	} else {
		log.Printf("No display detected. Starting live capture in HEADLESS mode on device %v", dev)
	}

	lastEncodeTime := time.Time{} // Force immediate first encode

	for {
		select {
		case <-p.stopSignal:
			log.Println("Capture loop stopping...")
			return
		default:
			if ok := webcam.Read(&img); !ok || img.Empty() {
				log.Printf("Device %v: failed to read frame, attempting to reconnect...", dev)
				webcam.Close()
				var err error
				webcam, err = gocv.OpenVideoCapture(dev)
				if err != nil {
					log.Printf("Device %v: reconnection failed: %v", dev, err)
					time.Sleep(2 * time.Second)
				}
				continue
			}

			// 1. Throttled Base64 Encoding for LLM
			if time.Since(lastEncodeTime) >= p.Interval {
				gocv.Resize(img, &resized, image.Point{X: p.Width, Y: p.Height}, 0, 0, gocv.InterpolationLinear)
				buf, err := gocv.IMEncode(".jpg", resized)
				if err == nil {
					encoded := base64.StdEncoding.EncodeToString(buf.GetBytes())
					buf.Close()
					p.mu.Lock()
					p.lastFrame = encoded
					p.mu.Unlock()
					lastEncodeTime = time.Now()
				}
			}

			// 2. Draw HUD and Show (only if display is available)
			if hasDisplay && window != nil {
				p.mu.Lock()
				overlay := p.overlayText
				p.mu.Unlock()

				if overlay != "" {
					p.drawHUD(&img, overlay)
				}

				window.IMShow(img)
				if window.WaitKey(1) == 27 { // ESC to exit
					p.Stop()
					return
				}
			} else {
				// In headless mode, we don't want to peg the CPU at 100% reading frames
				// so we add a tiny sleep to mimic a framerate.
				time.Sleep(30 * time.Millisecond)
			}
		}
	}
}

func (p *Pipeline) drawHUD(img *gocv.Mat, text string) {
	// Background semi-transparent rectangle for readability
	gocv.Rectangle(img, image.Rect(0, 0, img.Cols(), 80), color.RGBA{0, 0, 0, 150}, -1)

	// Draw lines of text
	lines := p.wrapText(text, 50) // Approx 50 chars per line
	for i, line := range lines {
		if i > 2 {
			break // Show only first 3 lines on HUD
		}
		gocv.PutText(img, line, image.Pt(10, 25+(i*20)), gocv.FontHersheySimplex, 0.6, color.RGBA{0, 255, 0, 0}, 2)
	}
}

func (p *Pipeline) wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	lines := []string{}
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+len(word)+1 <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}
