package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/p2p"
)

type bootstrapNode struct {
	ID        string `json:"id"`
	Multiaddr string `json:"multiaddr"`
	Region    string `json:"region"`
	Role      string `json:"role"`
}

type networkConfig struct {
	Network             string `json:"network"`
	Transport           string `json:"transport"`
	PubSub              string `json:"pubsub"`
	DefaultTopic        string `json:"default_topic"`
	RoundTimeoutSeconds int    `json:"round_timeout_seconds"`
	MinPeers            int    `json:"min_peers"`
	MaxPeers            int    `json:"max_peers"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: sovereign-node <start|join>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		if err := run("start", os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "start failed: %v\n", err)
			os.Exit(1)
		}
	case "join":
		if err := run("join", os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "join failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func run(mode string, args []string) error {
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	nodeID := fs.String("node-id", defaultNodeID(), "local node identifier")
	bootstrapPath := fs.String("bootstrap", "network/bootstrap/bootstrap_nodes.json", "path to bootstrap nodes json")
	seedsPath := fs.String("seeds", "network/bootstrap/seed_peers.json", "path to seed peers json")
	configPath := fs.String("config", "network/bootstrap/network_config.json", "path to network config json")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadNetworkConfig(*configPath)
	if err != nil {
		return err
	}
	bootstrapNodes, err := loadBootstrapNodes(*bootstrapPath)
	if err != nil {
		return err
	}
	seedPeers, err := loadSeedPeers(*seedsPath)
	if err != nil {
		return err
	}

	network := p2p.NewNetwork(*nodeID, 3, time.Duration(cfg.RoundTimeoutSeconds)*time.Second)
	addBootstrapPeers(network, bootstrapNodes, seedPeers)
	dialed := network.DialAllPeers()
	topic := strings.TrimSpace(cfg.DefaultTopic)
	if topic == "" {
		topic = "fl.rounds"
	}
	network.JoinTopic(topic)

	hello := fmt.Sprintf("node=%s mode=%s ts=%s", *nodeID, mode, time.Now().UTC().Format(time.RFC3339))
	publishedTo, err := network.Publish(topic, []byte(hello))
	if err != nil {
		return fmt.Errorf("publish startup gossip: %w", err)
	}

	fmt.Printf("node=%s mode=%s network=%s transport=%s pubsub=%s topic=%s peers=%d dialed=%d gossip_fanout=%d\n",
		*nodeID, mode, cfg.Network, cfg.Transport, cfg.PubSub, topic, network.GetActivePeerCount(), dialed, publishedTo)

	if mode == "join" && dialed == 0 {
		return errors.New("join requires at least one bootstrap or seed peer")
	}

	return nil
}

func defaultNodeID() string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "sovereign-node"
	}
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}

func loadBootstrapNodes(path string) ([]bootstrapNode, error) {
	var nodes []bootstrapNode
	if err := readJSON(path, &nodes); err != nil {
		return nil, fmt.Errorf("load bootstrap nodes from %s: %w", path, err)
	}
	return nodes, nil
}

func loadSeedPeers(path string) ([]string, error) {
	var seeds []string
	if err := readJSON(path, &seeds); err != nil {
		return nil, fmt.Errorf("load seed peers from %s: %w", path, err)
	}
	return seeds, nil
}

func loadNetworkConfig(path string) (networkConfig, error) {
	var cfg networkConfig
	if err := readJSON(path, &cfg); err != nil {
		return networkConfig{}, fmt.Errorf("load network config from %s: %w", path, err)
	}
	if cfg.RoundTimeoutSeconds <= 0 {
		cfg.RoundTimeoutSeconds = 120
	}
	if cfg.Network == "" {
		cfg.Network = "sovereign-testnet"
	}
	if cfg.Transport == "" {
		cfg.Transport = "quic"
	}
	if cfg.PubSub == "" {
		cfg.PubSub = "gossip"
	}
	if cfg.DefaultTopic == "" {
		cfg.DefaultTopic = "fl.rounds"
	}
	return cfg, nil
}

func readJSON(path string, out interface{}) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

func addBootstrapPeers(network *p2p.Network, nodes []bootstrapNode, seeds []string) {
	for _, n := range nodes {
		id := strings.TrimSpace(n.ID)
		if id == "" {
			continue
		}
		addr := strings.TrimSpace(n.Multiaddr)
		if addr == "" {
			continue
		}
		network.AddPeer(id, addr, 1.0)
	}

	for i, seed := range seeds {
		addr := strings.TrimSpace(seed)
		if addr == "" {
			continue
		}
		network.AddPeer(fmt.Sprintf("seed-%d", i+1), addr, 0.9)
	}
}
