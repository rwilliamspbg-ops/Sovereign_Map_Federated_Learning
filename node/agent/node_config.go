package agent

// NodeConfig contains high-level runtime settings for local node execution.
type NodeConfig struct {
	NetworkName string
	MinPeers    int
	MaxPeers    int
}
