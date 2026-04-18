package autonomy

import (
	"math"
	"sync"
	"time"
)

// MapCell stores confidence-scored occupancy for a world coordinate.
type MapCell struct {
	X             int       `json:"x"`
	Y             int       `json:"y"`
	Occupancy     float64   `json:"occupancy"`
	Confidence    float64   `json:"confidence"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// LayeredMapSnapshot is a read model for planning and API surfaces.
type LayeredMapSnapshot struct {
	Layer          string    `json:"layer"`
	CellCount      int       `json:"cell_count"`
	MeanConfidence float64   `json:"mean_confidence"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// LayeredMapStore keeps a confidence-scored map for one logical layer.
type LayeredMapStore struct {
	mu    sync.RWMutex
	layer string
	cells map[[2]int]MapCell
}

func NewLayeredMapStore(layer string) *LayeredMapStore {
	return &LayeredMapStore{
		layer: layer,
		cells: make(map[[2]int]MapCell),
	}
}

func (m *LayeredMapStore) UpsertCell(cell MapCell) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cell.Confidence = ClampUnitInterval(cell.Confidence)
	cell.Occupancy = ClampUnitInterval(cell.Occupancy)
	if cell.LastUpdatedAt.IsZero() {
		cell.LastUpdatedAt = time.Now().UTC()
	}
	m.cells[[2]int{cell.X, cell.Y}] = cell
}

func (m *LayeredMapStore) ApplyConfidenceDecay(now time.Time, halfLife time.Duration) {
	if halfLife <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.cells {
		age := now.Sub(v.LastUpdatedAt)
		if age <= 0 {
			continue
		}
		decay := math.Pow(0.5, age.Seconds()/halfLife.Seconds())
		v.Confidence = ClampUnitInterval(v.Confidence * decay)
		m.cells[k] = v
	}
}

func (m *LayeredMapStore) Snapshot() LayeredMapSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mean := 0.0
	for _, c := range m.cells {
		mean += c.Confidence
	}
	if len(m.cells) > 0 {
		mean = mean / float64(len(m.cells))
	}
	return LayeredMapSnapshot{
		Layer:          m.layer,
		CellCount:      len(m.cells),
		MeanConfidence: mean,
		GeneratedAt:    time.Now().UTC(),
	}
}

// TwinStore keeps the live digital twin state keyed by entity id.
type TwinStore struct {
	mu     sync.RWMutex
	states map[string]TwinEntityState
}

func NewTwinStore() *TwinStore {
	return &TwinStore{states: make(map[string]TwinEntityState)}
}

func (t *TwinStore) Upsert(state TwinEntityState) {
	t.mu.Lock()
	defer t.mu.Unlock()
	state.PoseConfidence = ClampUnitInterval(state.PoseConfidence)
	state.RiskScore = ClampUnitInterval(state.RiskScore)
	if state.ObservedAt.IsZero() {
		state.ObservedAt = time.Now().UTC()
	}
	t.states[state.EntityID] = state
}

func (t *TwinStore) Get(entityID string) (TwinEntityState, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	state, ok := t.states[entityID]
	return state, ok
}

func (t *TwinStore) StateLag(entityID string, now time.Time) (time.Duration, bool) {
	state, ok := t.Get(entityID)
	if !ok {
		return 0, false
	}
	return now.Sub(state.ObservedAt), true
}
