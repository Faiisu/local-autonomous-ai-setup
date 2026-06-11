package capture

import (
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"strconv"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type Pipeline struct {
	Device   string
	Width    int
	Height   int
	Interval time.Duration

	mu         sync.Mutex
	lastFrame  string
	isRunning  bool
	stopSignal chan struct{}
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

	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	log.Printf("Starting capture loop on device %v every %v", dev, p.Interval)

	for {
		select {
		case <-p.stopSignal:
			log.Println("Capture loop stopping...")
			return
		case <-ticker.C:
			if ok := webcam.Read(&img); !ok || img.Empty() {
				log.Printf("Device %v: failed to read frame, attempting to reconnect...", dev)
				webcam.Close()

				// Reconnect logic
				var err error
				webcam, err = gocv.OpenVideoCapture(dev)
				if err != nil {
					log.Printf("Device %v: reconnection failed: %v", dev, err)
					time.Sleep(2 * time.Second)
				}
				continue
			}

			// Preprocessing: Resize
			gocv.Resize(img, &resized, image.Point{X: p.Width, Y: p.Height}, 0, 0, gocv.InterpolationLinear)

			// Compress to JPEG
			buf, err := gocv.IMEncode(".jpg", resized)
			if err != nil {
				log.Printf("Failed to encode frame: %v", err)
				continue
			}

			// Convert to Base64
			encoded := base64.StdEncoding.EncodeToString(buf.GetBytes())
			buf.Close()

			p.mu.Lock()
			p.lastFrame = encoded
			p.mu.Unlock()
		}
	}
}
