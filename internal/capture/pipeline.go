package capture

import (
	"encoding/base64"
	"fmt"
	"image"
	"log"
	_"os"
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

	go p.captureLoop(webcam)
	return nil
}

func (p *Pipeline) Stop() {
	p.mu.Lock()
	if p.isRunning {
		close(p.stopSignal)
		p.isRunning = false
	}
	p.mu.Unlock()
}

func (p *Pipeline) GetLastFrame() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastFrame
}

func (p *Pipeline) captureLoop(webcam *gocv.VideoCapture) {
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	resized := gocv.NewMat()
	defer resized.Close()

	for {
		select {
		case <-p.stopSignal:
			return
		default:
			if ok := webcam.Read(&img); !ok || img.Empty() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			gocv.Resize(img, &resized, image.Point{X: p.Width, Y: p.Height}, 0, 0, gocv.InterpolationLinear)
			buf, err := gocv.IMEncode(".jpg", resized)
			if err == nil {
				encoded := base64.StdEncoding.EncodeToString(buf.GetBytes())
				buf.Close()
				p.mu.Lock()
				p.lastFrame = encoded
				p.mu.Unlock()
			}

			time.Sleep(p.Interval)
		}
	}
}
