//go:build opencv
// +build opencv

package slam

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// FeatureType identifies keypoint detector algorithm.
type FeatureType string

const (
	FeatureORB   FeatureType = "orb"
	FeatureSIFT  FeatureType = "sift"
	FeatureAKAZE FeatureType = "akaze"
)

// ExtractorConfig controls feature detection parameters.
type ExtractorConfig struct {
	Type          FeatureType
	MaxFeatures   int
	ScaleFactor   float32
	NLevels       int
	EdgeThreshold int
}

// FeatureExtractionResult tracks extracted feature counts and keypoints.
type FeatureExtractionResult struct {
	Features     int
	Keypoints    []gocv.KeyPoint
	Descriptors  gocv.Mat
	ProcessingMS float64
}

// Extractor performs keypoint detection and descriptor extraction.
type Extractor struct {
	config   ExtractorConfig
	detector gocv.ORB
	sift     gocv.SIFT
	akaze    gocv.AKAZE
	hasORB   bool
	hasSIFT  bool
	hasAKAZE bool
}

// NewExtractor creates feature extractor with specified algorithm.
func NewExtractor(cfg ExtractorConfig) (*Extractor, error) {
	if cfg.MaxFeatures == 0 {
		cfg.MaxFeatures = 1000
	}
	if cfg.ScaleFactor == 0 {
		cfg.ScaleFactor = 1.2
	}
	if cfg.NLevels == 0 {
		cfg.NLevels = 8
	}
	if cfg.EdgeThreshold == 0 {
		cfg.EdgeThreshold = 31
	}

	ext := &Extractor{config: cfg}

	switch cfg.Type {
	case FeatureORB:
		ext.detector = gocv.NewORBWithParams(
			cfg.MaxFeatures,
			cfg.ScaleFactor,
			cfg.NLevels,
			cfg.EdgeThreshold,
			0,
			2,
			gocv.ORBScoreTypeHarris,
			31,
			20,
		)
		ext.hasORB = true
	case FeatureSIFT:
		ext.sift = gocv.NewSIFT()
		ext.hasSIFT = true
	case FeatureAKAZE:
		ext.akaze = gocv.NewAKAZE()
		ext.hasAKAZE = true
	default:
		return nil, fmt.Errorf("unsupported feature type: %s", cfg.Type)
	}

	return ext, nil
}

// Extract detects keypoints and computes descriptors from image.
func (e *Extractor) Extract(img gocv.Mat) (FeatureExtractionResult, error) {
	if img.Empty() {
		return FeatureExtractionResult{}, fmt.Errorf("empty image matrix")
	}

	var keypoints []gocv.KeyPoint
	descriptors := gocv.NewMat()
	mask := gocv.NewMat()
	defer mask.Close()

	switch e.config.Type {
	case FeatureORB:
		keypoints, descriptors = e.detector.DetectAndCompute(img, mask)
	case FeatureSIFT:
		keypoints, descriptors = e.sift.DetectAndCompute(img, mask)
	case FeatureAKAZE:
		keypoints, descriptors = e.akaze.DetectAndCompute(img, mask)
	}

	return FeatureExtractionResult{
		Features:    len(keypoints),
		Keypoints:   keypoints,
		Descriptors: descriptors,
	}, nil
}

// ExtractFromBytes extracts features from JPEG/PNG encoded bytes.
func (e *Extractor) ExtractFromBytes(data []byte) (FeatureExtractionResult, error) {
	img, err := gocv.IMDecode(data, gocv.IMReadGrayScale)
	if err != nil {
		return FeatureExtractionResult{}, fmt.Errorf("decode image: %w", err)
	}
	defer img.Close()

	return e.Extract(img)
}

// ExtractFromImage extracts features from standard Go image.
func (e *Extractor) ExtractFromImage(img image.Image) (FeatureExtractionResult, error) {
	mat, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return FeatureExtractionResult{}, fmt.Errorf("convert to mat: %w", err)
	}
	defer mat.Close()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)

	return e.Extract(gray)
}

// Close releases detector resources.
func (e *Extractor) Close() error {
	if e.hasORB {
		e.detector.Close()
	}
	if e.hasSIFT {
		e.sift.Close()
	}
	if e.hasAKAZE {
		e.akaze.Close()
	}
	return nil
}

// MatchFeatures computes feature matches between two descriptor sets.
func MatchFeatures(desc1, desc2 gocv.Mat, matchType gocv.NormType) []gocv.DMatch {
	matcher := gocv.NewBFMatcherWithParams(matchType, false)
	defer matcher.Close()

	knn := matcher.KnnMatch(desc1, desc2, 1)
	matches := make([]gocv.DMatch, 0, len(knn))
	for _, group := range knn {
		if len(group) > 0 {
			matches = append(matches, group[0])
		}
	}
	return matches
}
