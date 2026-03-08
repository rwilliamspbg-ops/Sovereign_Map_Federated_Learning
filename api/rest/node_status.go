package rest

// NodeStatusResponse is a minimal response contract for node status requests.
type NodeStatusResponse struct {
	NodeID string `json:"node_id"`
	State  string `json:"state"`
}
