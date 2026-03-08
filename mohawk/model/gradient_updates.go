package model

// GradientUpdateMetadata tracks ownership and ordering of updates.
type GradientUpdateMetadata struct {
	NodeID string
	Round  int
}
