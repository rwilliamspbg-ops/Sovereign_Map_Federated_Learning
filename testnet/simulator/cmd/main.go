package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/testnet/simulator"
)

func main() {
	nodes := flag.Int("nodes", 50, "number of simulated nodes")
	rounds := flag.Int("rounds", 300, "number of rounds to run")
	roundMs := flag.Int("round-ms", 250, "base round duration in milliseconds")
	stragglerRate := flag.Float64("straggler-rate", 0.1, "fraction of rounds with straggler delay [0,1]")
	maliciousRate := flag.Float64("malicious-rate", 0.02, "fraction of rounds with malicious-event detection [0,1]")
	seed := flag.Int64("seed", 0, "random seed (0 uses current time)")
	flag.Parse()

	cfg := simulator.Config{
		NodeCount:         *nodes,
		Rounds:            *rounds,
		RoundDuration:     time.Duration(*roundMs) * time.Millisecond,
		StragglerRate:     *stragglerRate,
		MaliciousNodeRate: *maliciousRate,
		RandomSeed:        *seed,
	}

	result := simulator.Run(cfg)
	summary := simulator.FormatSummary(result)
	fmt.Println(summary)

	if result.RoundsCompleted < result.RoundsRequested {
		fmt.Fprintln(os.Stderr, "simulation did not complete all requested rounds")
		os.Exit(1)
	}
}
