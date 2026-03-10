//go:build opencv
// +build opencv

package camera

import (
	"context"
	"fmt"
	"image"
	"time"

	"gocv.io/x/gocv"
)

// FrameSource identifies camera input type.
type FrameSource string

const (
	SourceWebcam   FrameSource = "webcam"
	SourceRTSP     FrameSource = "rtsp"
	SourceFile     FrameSource = "file"
	SourceMobileIP FrameSource = "mobile-ip"
)

// CaptureConfig controls camera frame ingestion.
type CaptureConfig struct {
	Source     FrameSource
	DeviceID   int
	RTSPURL    string
	FilePath   string
	FPS        int
	Width      int
	Height     int
	BufferSize int
}

// FrameCaptureResult tracks captured frame count for ingestion.
type FrameCaptureResult struct {
	Frames    int
	Timestamp time.Time
	Image     *gocv.Mat
}

// Capture manages real-time camera frame ingestion.
type Capture struct {
	config CaptureConfig
	camera *gocv.VideoCapture
	buffer chan FrameCaptureResult
}

// NewCapture creates camera capture handler for specified source.
func NewCapture(cfg CaptureConfig) (*Capture, error) {
	if cfg.FPS == 0 {
		cfg.FPS = 30
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 100
	}
	if cfg.Width == 0 {
		cfg.Width = 1280
	}
	if cfg.Height == 0 {
		cfg.Height = 720
	}

	var camera *gocv.VideoCapture
	var err error

	switch cfg.Source {
	case SourceWebcam:
		camera, err = gocv.OpenVideoCapture(cfg.DeviceID)
	case SourceRTSP, SourceMobileIP:
		camera, err = gocv.OpenVideoCapture(cfg.RTSPURL)
	case SourceFile:
		camera, err = gocv.OpenVideoCapture(cfg.FilePath)
	default:
		return nil, fmt.Errorf("unsupported source: %s", cfg.Source)
	}

	if err != nil {
		return nil, fmt.Errorf("open camera: %w", err)
	}

	// Set resolution
	camera.Set(gocv.VideoCaptureFrameWidth, float64(cfg.Width))
	camera.Set(gocv.VideoCaptureFrameHeight, float64(cfg.Height))
	camera.Set(gocv.VideoCaptureFPS, float64(cfg.FPS))

	return &Capture{
		config: cfg,
		camera: camera,
		buffer: make(chan FrameCaptureResult, cfg.BufferSize),
	}, nil
}

// Start begins continuous frame capture with background worker.
func (c *Capture) Start(ctx context.Context) error {
	if c.camera == nil {
		return fmt.Errorf("camera not initialized")
	}

	go c.captureLoop(ctx)
	return nil
}

// captureLoop continuously reads frames from camera until context cancelled.
func (c *Capture) captureLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second / time.Duration(c.config.FPS))
	defer ticker.Stop()

	frameCount := 0
	mat := gocv.NewMat()
	defer mat.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ok := c.camera.Read(&mat); !ok || mat.Empty() {
				continue
			}

			frameCount++
			frameCopy := mat.Clone()

			select {
			case c.buffer <- FrameCaptureResult{
				Frames:    frameCount,
				Timestamp: time.Now(),
				Image:     &frameCopy,
			}:
			default:
				// Buffer full, drop frame
				frameCopy.Close()
			}
		}
	}
}

// NextFrame returns next captured frame from buffer (blocking).
func (c *Capture) NextFrame(ctx context.Context) (FrameCaptureResult, error) {
	select {
	case <-ctx.Done():
		return FrameCaptureResult{}, ctx.Err()
	case frame := <-c.buffer:
		return frame, nil
	}
}

// Close releases camera resources.
func (c *Capture) Close() error {
	if c.camera != nil {
		if err := c.camera.Close(); err != nil {
			return fmt.Errorf("close camera: %w", err)
		}
	}
	close(c.buffer)
	return nil
}

// ToBytes converts frame to JPEG-encoded bytes for transmission.
func (f FrameCaptureResult) ToBytes() ([]byte, error) {
	if f.Image == nil {
		return nil, fmt.Errorf("frame image is nil")
	}
	buf, err := gocv.IMEncode(".jpg", *f.Image)
	if err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}
	return buf.GetBytes(), nil
}

// ToImage converts Mat to standard Go image.
func (f FrameCaptureResult) ToImage() (image.Image, error) {
	if f.Image == nil {
		return nil, fmt.Errorf("frame image is nil")
	}
	img, err := f.Image.ToImage()
	if err != nil {
		return nil, fmt.Errorf("convert to image: %w", err)
	}
	return img, nil
}
