package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/scheduler"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/testnet/simulator"
)

type governanceCounters struct {
	mu         sync.Mutex
	executions uint64
	failures   uint64
}

func (g *governanceCounters) record(success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.executions++
	if !success {
		g.failures++
	}
}

func (g *governanceCounters) snapshot() (uint64, uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.executions, g.failures
}

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

	validatorSet := blockchain.NewValidatorSet()
	governanceMetrics := &governanceCounters{}
	_ = validatorSet.AddValidator("node_1", 2000000)
	_ = validatorSet.AddValidator("node_2", 1500000)
	_ = validatorSet.AddValidator("node_3", 1000000)
	_ = validatorSet.SetAttestationScore("node_1", 9500)
	_ = validatorSet.SetAttestationScore("node_2", 8000)
	_ = validatorSet.SetAttestationScore("node_3", 6500)
	_ = validatorSet.RecordParticipationQuality("node_1", 9000)
	_ = validatorSet.RecordParticipationQuality("node_2", 7200)
	_ = validatorSet.RecordParticipationQuality("node_3", 5800)

	flChain := seedDemoFLChain()

	http.HandleFunc("/metrics/scheduler", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(schedulerMetrics.Prometheus(*nodeID)))
	})

	http.HandleFunc("/metrics/simulator", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(simulator.Prometheus(simResult, *scenario)))
	})

	http.HandleFunc("/metrics/validators", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(validatorMetricsPrometheus(validatorSet, governanceMetrics)))
	})

	http.HandleFunc("/metrics/blockchain", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(blockchainMetricsPrometheus(flChain)))
	})

	http.HandleFunc("/event/governance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Success bool `json:"success"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		governanceMetrics.record(payload.Success)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	})

	fmt.Printf("starting metrics-exporter on %s\n", *listen)
	if err := http.ListenAndServe(*listen, nil); err != nil {
		fmt.Fprintf(os.Stderr, "metrics-exporter failed: %v\n", err)
		os.Exit(1)
	}
}

func validatorMetricsPrometheus(vs *blockchain.ValidatorSet, governance *governanceCounters) string {
	metrics := vs.GetMetrics()
	policy := vs.GetReputationPolicy()
	executions, failures := governance.snapshot()

	var b strings.Builder
	b.WriteString("# HELP sovereign_validators_total Number of validators in set\n")
	b.WriteString("# TYPE sovereign_validators_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_total %d\n", metrics.ValidatorCount))

	b.WriteString("# HELP sovereign_validators_avg_reputation Average validator reputation (bps)\n")
	b.WriteString("# TYPE sovereign_validators_avg_reputation gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_avg_reputation %d\n", metrics.AverageReputation))

	b.WriteString("# HELP sovereign_validators_avg_attestation Average validator attestation trust score (bps)\n")
	b.WriteString("# TYPE sovereign_validators_avg_attestation gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_avg_attestation %d\n", metrics.AverageAttestation))

	b.WriteString("# HELP sovereign_validators_avg_quality Average validator participation quality score (bps)\n")
	b.WriteString("# TYPE sovereign_validators_avg_quality gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_avg_quality %d\n", metrics.AverageQuality))

	b.WriteString("# HELP sovereign_validators_low_reputation_total Validators with low reputation\n")
	b.WriteString("# TYPE sovereign_validators_low_reputation_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_low_reputation_total %d\n", metrics.LowReputationCount))

	b.WriteString("# HELP sovereign_validators_low_attestation_total Validators with low attestation score\n")
	b.WriteString("# TYPE sovereign_validators_low_attestation_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_low_attestation_total %d\n", metrics.LowAttestationCount))

	b.WriteString("# HELP sovereign_validators_stale_attestation_total Validators with stale attestation state\n")
	b.WriteString("# TYPE sovereign_validators_stale_attestation_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_stale_attestation_total %d\n", metrics.StaleAttestationCount))

	b.WriteString("# HELP sovereign_validators_missing_attestation_total Validators without a valid attestation yet\n")
	b.WriteString("# TYPE sovereign_validators_missing_attestation_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_missing_attestation_total %d\n", metrics.MissingAttestationCount))

	b.WriteString("# HELP sovereign_validators_attestation_failures_total Total invalid attestation events\n")
	b.WriteString("# TYPE sovereign_validators_attestation_failures_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_attestation_failures_total %d\n", metrics.TotalAttestationFailures))

	b.WriteString("# HELP sovereign_validators_jailed_total Jailed validators\n")
	b.WriteString("# TYPE sovereign_validators_jailed_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_validators_jailed_total %d\n", metrics.JailedCount))

	b.WriteString("# HELP sovereign_reputation_policy_weight_reputation Governance-set reputation weight\n")
	b.WriteString("# TYPE sovereign_reputation_policy_weight_reputation gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_reputation_policy_weight_reputation %d\n", policy.ReputationWeight))

	b.WriteString("# HELP sovereign_reputation_policy_weight_attestation Governance-set attestation weight\n")
	b.WriteString("# TYPE sovereign_reputation_policy_weight_attestation gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_reputation_policy_weight_attestation %d\n", policy.AttestationWeight))

	b.WriteString("# HELP sovereign_reputation_policy_weight_quality Governance-set quality weight\n")
	b.WriteString("# TYPE sovereign_reputation_policy_weight_quality gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_reputation_policy_weight_quality %d\n", policy.QualityWeight))

	b.WriteString("# HELP sovereign_governance_executions_total Governance execution attempts observed by exporter\n")
	b.WriteString("# TYPE sovereign_governance_executions_total counter\n")
	b.WriteString(fmt.Sprintf("sovereign_governance_executions_total %d\n", executions))

	b.WriteString("# HELP sovereign_governance_failures_total Failed governance executions observed by exporter\n")
	b.WriteString("# TYPE sovereign_governance_failures_total counter\n")
	b.WriteString(fmt.Sprintf("sovereign_governance_failures_total %d\n", failures))

	return b.String()
}

// seedDemoFLChain creates a small blockchain pre-loaded with FL verification rounds
// for the metrics exporter demo.  In production this would be replaced by a live
// chain reference shared with the node agent or sync component.
func seedDemoFLChain() *blockchain.BlockChain {
	bc := blockchain.NewBlockChain()
	proposer := blockchain.NewBlockProposer("metrics-node", bc)

	// Seed three FL rounds with varying proof confidence levels.
	rounds := []map[string]interface{}{
		{"round_id": "demo-round-1", "model_hash": "a1b2c3", "proof_type": "consensus"},
		{"round_id": "demo-round-2", "model_hash": "d4e5f6", "proof_type": "consensus"},
		{"round_id": "demo-round-3", "model_hash": "g7h8i9", "proof_type": "consensus"},
	}
	for _, rd := range rounds {
		blk, err := proposer.ProposeBlock("validator-0", rd)
		if err != nil {
			continue
		}
		_ = proposer.CommitBlock(blk)
	}
	return bc
}

// blockchainMetricsPrometheus serialises FL verification chain metrics in
// Prometheus text format.  This endpoint complements /metrics/validators by
// exposing chain-level proof-of-correctness observability.
func blockchainMetricsPrometheus(bc *blockchain.BlockChain) string {
	m := bc.GetFLVerificationMetrics()

	var b strings.Builder

	b.WriteString("# HELP sovereign_fl_rounds_total Total FL rounds committed on chain\n")
	b.WriteString("# TYPE sovereign_fl_rounds_total gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_fl_rounds_total %d\n", m.TotalRounds))

	b.WriteString("# HELP sovereign_fl_rounds_verified FL rounds that passed proof verification\n")
	b.WriteString("# TYPE sovereign_fl_rounds_verified gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_fl_rounds_verified %d\n", m.VerifiedRounds))

	b.WriteString("# HELP sovereign_fl_rounds_failed FL rounds that failed proof verification\n")
	b.WriteString("# TYPE sovereign_fl_rounds_failed gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_fl_rounds_failed %d\n", m.FailedRounds))

	b.WriteString("# HELP sovereign_fl_verification_ratio Ratio of verified to total FL rounds (0-1)\n")
	b.WriteString("# TYPE sovereign_fl_verification_ratio gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_fl_verification_ratio %.4f\n", m.VerifiedRatio))

	b.WriteString("# HELP sovereign_fl_avg_confidence_bps Average FL proof confidence in basis points\n")
	b.WriteString("# TYPE sovereign_fl_avg_confidence_bps gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_fl_avg_confidence_bps %d\n", m.AverageConfidenceBps))

	return b.String()
}
