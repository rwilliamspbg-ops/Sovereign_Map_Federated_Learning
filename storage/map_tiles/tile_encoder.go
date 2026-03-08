package maptiles

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"

	"gocv.io/x/gocv"
)

// TileFormat identifies tile encoding format.
type TileFormat string

const (
	FormatPNG  TileFormat = "png"
	FormatJPEG TileFormat = "jpeg"
	FormatWebP TileFormat = "webp"
)

// TileEncoderConfig stores map tile encoding preferences.
type TileEncoderConfig struct {
	Format      TileFormat
	TileSize    int
	Quality     int
	Compression int
}

// TileCoordinate identifies tile in slippy map scheme (zoom/x/y).
type TileCoordinate struct {
	Zoom int
	X    int
	Y    int
}

// LatLng represents geographic coordinates.
type LatLng struct {
	Lat float64
	Lng float64
}

// TileEncoder generates map tiles from imagery and SLAM data.
type TileEncoder struct {
	config TileEncoderConfig
}

// NewTileEncoder creates tile encoder with specified config.
func NewTileEncoder(cfg TileEncoderConfig) (*TileEncoder, error) {
	if cfg.TileSize == 0 {
		cfg.TileSize = 256
	}
	if cfg.Quality == 0 {
		cfg.Quality = 80
	}
	if cfg.Compression == 0 {
		cfg.Compression = 6
	}
	if cfg.Format == "" {
		cfg.Format = FormatPNG
	}

	return &TileEncoder{config: cfg}, nil
}

// Encode converts image to tile bytes in configured format.
func (e *TileEncoder) Encode(img image.Image) ([]byte, error) {
	buf := &bytes.Buffer{}

	switch e.config.Format {
	case FormatPNG:
		encoder := png.Encoder{CompressionLevel: png.CompressionLevel(e.config.Compression)}
		if err := encoder.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	case FormatJPEG:
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: e.config.Quality}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	case FormatWebP:
		return nil, fmt.Errorf("webp encoding not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", e.config.Format)
	}

	return buf.Bytes(), nil
}

// EncodeFromMat converts OpenCV Mat to tile bytes.
func (e *TileEncoder) EncodeFromMat(mat gocv.Mat) ([]byte, error) {
	if mat.Empty() {
		return nil, fmt.Errorf("empty mat")
	}

	// Resize to tile dimensions
	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(mat, &resized, image.Pt(e.config.TileSize, e.config.TileSize), 0, 0, gocv.InterpolationLinear)

	img, err := resized.ToImage()
	if err != nil {
		return nil, fmt.Errorf("mat to image: %w", err)
	}

	return e.Encode(img)
}

// LatLngToTile converts geographic coordinates to tile coordinate at zoom level.
func LatLngToTile(lat, lng float64, zoom int) TileCoordinate {
	n := math.Pow(2.0, float64(zoom))
	x := int((lng + 180.0) / 360.0 * n)

	latRad := lat * math.Pi / 180.0
	y := int((1.0 - math.Log(math.Tan(latRad)+(1.0/math.Cos(latRad)))/math.Pi) / 2.0 * n)

	return TileCoordinate{Zoom: zoom, X: x, Y: y}
}

// TileToLatLng converts tile coordinate to northwest corner geographic coordinates.
func TileToLatLng(tile TileCoordinate) LatLng {
	n := math.Pow(2.0, float64(tile.Zoom))
	lng := float64(tile.X)/n*360.0 - 180.0

	latRad := math.Atan(math.Sinh(math.Pi * (1.0 - 2.0*float64(tile.Y)/n)))
	lat := latRad * 180.0 / math.Pi

	return LatLng{Lat: lat, Lng: lng}
}

// TilePath generates filesystem path for tile storage.
func TilePath(tile TileCoordinate, format TileFormat) string {
	ext := "png"
	switch format {
	case FormatJPEG:
		ext = "jpg"
	case FormatWebP:
		ext = "webp"
	}
	return fmt.Sprintf("%d/%d/%d.%s", tile.Zoom, tile.X, tile.Y, ext)
}

// TileBounds returns geographic bounds of a tile.
func TileBounds(tile TileCoordinate) (nw, se LatLng) {
	nw = TileToLatLng(tile)
	se = TileToLatLng(TileCoordinate{
		Zoom: tile.Zoom,
		X:    tile.X + 1,
		Y:    tile.Y + 1,
	})
	return nw, se
}
