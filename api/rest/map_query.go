package rest

import "time"

// MapQueryRequest defines lightweight filtering for map snapshots.
type MapQueryRequest struct {
	Layer            string  `json:"layer"`
	MinConfidence    float64 `json:"min_confidence,omitempty"`
	MaxStalenessSecs int     `json:"max_staleness_secs,omitempty"`
}

// MapQueryResponse returns map layer summary and twin freshness hints.
type MapQueryResponse struct {
	Layer            string    `json:"layer"`
	TileCount        int       `json:"tile_count"`
	MeanConfidence   float64   `json:"mean_confidence"`
	TwinLagMillis    int64     `json:"twin_lag_millis"`
	LastUpdatedAtUTC time.Time `json:"last_updated_at_utc"`
}
