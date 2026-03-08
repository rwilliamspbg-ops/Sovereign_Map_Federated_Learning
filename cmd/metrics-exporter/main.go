package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/scheduler"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/testnet/simulator"
)

func main() {
	listen := flag.String("listen", ":9108", "metrics exporter listen address")
	nodeID := flag.String("node-id", "metrics-exporter", "node id label for scheduler metrics")
	scenario := flag.String("scenario", "baseline-50-node", "scenario label for simulator metrics")
	flag.Parse()

	schedulerMetrics := scheduler.NewMetrics()
	schedulerMetrics.RecordRoundStart(1)
	schedulerMetrics.RecordStraggler()
	schedulerMetrics.RecordRoundComplete(1)

	simResult := simulator.Run(simulator.Config{
		NodeCount:         50,
		Rounds:            100,
		RoundDuration:     250 * time.Millisecond,
		StragglerRate:     0.10,
		MaliciousNodeRate: 0.02,
	})

	http.HandleFunc("/metrics/scheduler", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(schedulerMetrics.Prometheus(*nodeID)))
	})

	http.HandleFunc("/metrics/simulator", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(simulator.Prometheus(simResult, *scenario)))
	})

	fmt.Printf("starting metrics-exporter on %s\n", *listen)
	if err := http.ListenAndServe(*listen, nil); err != nil {
		fmt.Fprintf(os.Stderr, "metrics-exporter failed: %v\n", err)
		os.Exit(1)
	}
}
