package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	apiBaseURL := flag.String("api-base-url", "http://localhost:8082", "node-agent API base URL for consensus metrics")
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
	httpClient := &http.Client{Timeout: 3 * time.Second}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(combinedMetricsPrometheus(
			schedulerMetrics.Prometheus(*nodeID),
			simulator.Prometheus(simResult, *scenario),
			validatorMetricsPrometheus(validatorSet, governanceMetrics),
			blockchainMetricsPrometheus(flChain),
			consensusMetricsPrometheus(httpClient, *apiBaseURL),
		)))
	})

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

	http.HandleFunc("/metrics/consensus", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(consensusMetricsPrometheus(httpClient, *apiBaseURL)))
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
	server := &http.Server{
		Addr:              *listen,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
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
	fmt.Fprintf(&b, "sovereign_validators_total %d\n", metrics.ValidatorCount)

	b.WriteString("# HELP sovereign_validators_avg_reputation Average validator reputation (bps)\n")
	b.WriteString("# TYPE sovereign_validators_avg_reputation gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_avg_reputation %d\n", metrics.AverageReputation)

	b.WriteString("# HELP sovereign_validators_avg_attestation Average validator attestation trust score (bps)\n")
	b.WriteString("# TYPE sovereign_validators_avg_attestation gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_avg_attestation %d\n", metrics.AverageAttestation)

	b.WriteString("# HELP sovereign_validators_avg_quality Average validator participation quality score (bps)\n")
	b.WriteString("# TYPE sovereign_validators_avg_quality gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_avg_quality %d\n", metrics.AverageQuality)

	b.WriteString("# HELP sovereign_validators_low_reputation_total Validators with low reputation\n")
	b.WriteString("# TYPE sovereign_validators_low_reputation_total gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_low_reputation_total %d\n", metrics.LowReputationCount)

	b.WriteString("# HELP sovereign_validators_low_attestation_total Validators with low attestation score\n")
	b.WriteString("# TYPE sovereign_validators_low_attestation_total gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_low_attestation_total %d\n", metrics.LowAttestationCount)

	b.WriteString("# HELP sovereign_validators_stale_attestation_total Validators with stale attestation state\n")
	b.WriteString("# TYPE sovereign_validators_stale_attestation_total gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_stale_attestation_total %d\n", metrics.StaleAttestationCount)

	b.WriteString("# HELP sovereign_validators_missing_attestation_total Validators without a valid attestation yet\n")
	b.WriteString("# TYPE sovereign_validators_missing_attestation_total gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_missing_attestation_total %d\n", metrics.MissingAttestationCount)

	b.WriteString("# HELP sovereign_validators_attestation_failures_total Total invalid attestation events\n")
	b.WriteString("# TYPE sovereign_validators_attestation_failures_total gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_attestation_failures_total %d\n", metrics.TotalAttestationFailures)

	b.WriteString("# HELP sovereign_validators_jailed_total Jailed validators\n")
	b.WriteString("# TYPE sovereign_validators_jailed_total gauge\n")
	fmt.Fprintf(&b, "sovereign_validators_jailed_total %d\n", metrics.JailedCount)

	b.WriteString("# HELP sovereign_reputation_policy_weight_reputation Governance-set reputation weight\n")
	b.WriteString("# TYPE sovereign_reputation_policy_weight_reputation gauge\n")
	fmt.Fprintf(&b, "sovereign_reputation_policy_weight_reputation %d\n", policy.ReputationWeight)

	b.WriteString("# HELP sovereign_reputation_policy_weight_attestation Governance-set attestation weight\n")
	b.WriteString("# TYPE sovereign_reputation_policy_weight_attestation gauge\n")
	fmt.Fprintf(&b, "sovereign_reputation_policy_weight_attestation %d\n", policy.AttestationWeight)

	b.WriteString("# HELP sovereign_reputation_policy_weight_quality Governance-set quality weight\n")
	b.WriteString("# TYPE sovereign_reputation_policy_weight_quality gauge\n")
	fmt.Fprintf(&b, "sovereign_reputation_policy_weight_quality %d\n", policy.QualityWeight)

	b.WriteString("# HELP sovereign_governance_executions_total Governance execution attempts observed by exporter\n")
	b.WriteString("# TYPE sovereign_governance_executions_total counter\n")
	fmt.Fprintf(&b, "sovereign_governance_executions_total %d\n", executions)

	b.WriteString("# HELP sovereign_governance_failures_total Failed governance executions observed by exporter\n")
	b.WriteString("# TYPE sovereign_governance_failures_total counter\n")
	fmt.Fprintf(&b, "sovereign_governance_failures_total %d\n", failures)

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
	fmt.Fprintf(&b, "sovereign_fl_rounds_total %d\n", m.TotalRounds)

	b.WriteString("# HELP sovereign_fl_rounds_verified FL rounds that passed proof verification\n")
	b.WriteString("# TYPE sovereign_fl_rounds_verified gauge\n")
	fmt.Fprintf(&b, "sovereign_fl_rounds_verified %d\n", m.VerifiedRounds)

	b.WriteString("# HELP sovereign_fl_rounds_failed FL rounds that failed proof verification\n")
	b.WriteString("# TYPE sovereign_fl_rounds_failed gauge\n")
	fmt.Fprintf(&b, "sovereign_fl_rounds_failed %d\n", m.FailedRounds)

	b.WriteString("# HELP sovereign_fl_verification_ratio Ratio of verified to total FL rounds (0-1)\n")
	b.WriteString("# TYPE sovereign_fl_verification_ratio gauge\n")
	fmt.Fprintf(&b, "sovereign_fl_verification_ratio %.4f\n", m.VerifiedRatio)

	b.WriteString("# HELP sovereign_fl_avg_confidence_bps Average FL proof confidence in basis points\n")
	b.WriteString("# TYPE sovereign_fl_avg_confidence_bps gauge\n")
	fmt.Fprintf(&b, "sovereign_fl_avg_confidence_bps %d\n", m.AverageConfidenceBps)

	return b.String()
}

type apiConsensusMetrics struct {
	consensusUp              float64
	asyncMode                float64
	activeNodes              float64
	quorumSize               float64
	openRounds               float64
	aggregationRound         float64
	bufferedModels           float64
	aggregationAsyncRounds   float64
	aggregationStaleDrops    float64
	aggregationAvgLatencyMS  float64
	churnJoinsTotal          float64
	churnLeavesTotal         float64
	stalenessAvgSeconds      float64
	privacyCumulativeEpsilon float64
	privacyEpsilonTarget     float64
	stragglerRatePct         float64
	attackSuccessRatePct     float64
	detectionPrecisionPct    float64
	epcUtilizationPct        float64
	attestationLatencyMS     float64
	cxlUtilizationPct        float64
	cxlThroughputGBps        float64
	npuTempC                 float64
	tpmTempC                 float64
	governanceTotalStake     float64
	stakeConcentrationPct    float64
	slashingEventsTotal      float64
	rewardAPYPct             float64
}

func consensusMetricsPrometheus(client *http.Client, apiBaseURL string) string {
	metrics := collectConsensusMetrics(client, apiBaseURL)

	var b strings.Builder
	b.WriteString("# HELP sovereign_consensus_status_up Whether consensus status endpoints are reachable (1=yes, 0=no)\n")
	b.WriteString("# TYPE sovereign_consensus_status_up gauge\n")
	fmt.Fprintf(&b, "sovereign_consensus_status_up %.0f\n", metrics.consensusUp)

	b.WriteString("# HELP sovereign_consensus_async_mode Consensus async mode enabled flag (1/0)\n")
	b.WriteString("# TYPE sovereign_consensus_async_mode gauge\n")
	fmt.Fprintf(&b, "sovereign_consensus_async_mode %.0f\n", metrics.asyncMode)

	b.WriteString("# HELP sovereign_consensus_active_nodes Active consensus node count\n")
	b.WriteString("# TYPE sovereign_consensus_active_nodes gauge\n")
	fmt.Fprintf(&b, "sovereign_consensus_active_nodes %.0f\n", metrics.activeNodes)

	b.WriteString("# HELP sovereign_consensus_quorum_size Current consensus quorum size\n")
	b.WriteString("# TYPE sovereign_consensus_quorum_size gauge\n")
	fmt.Fprintf(&b, "sovereign_consensus_quorum_size %.0f\n", metrics.quorumSize)

	b.WriteString("# HELP sovereign_consensus_open_rounds Open consensus rounds awaiting completion\n")
	b.WriteString("# TYPE sovereign_consensus_open_rounds gauge\n")
	fmt.Fprintf(&b, "sovereign_consensus_open_rounds %.0f\n", metrics.openRounds)

	b.WriteString("# HELP sovereign_aggregation_round_number Latest aggregation round number\n")
	b.WriteString("# TYPE sovereign_aggregation_round_number gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_round_number %.0f\n", metrics.aggregationRound)

	b.WriteString("# HELP sovereign_aggregation_buffered_models Buffered model submissions in current node\n")
	b.WriteString("# TYPE sovereign_aggregation_buffered_models gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_buffered_models %.0f\n", metrics.bufferedModels)

	b.WriteString("# HELP sovereign_aggregation_async_rounds_total Aggregation rounds committed in async mode\n")
	b.WriteString("# TYPE sovereign_aggregation_async_rounds_total gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_async_rounds_total %.0f\n", metrics.aggregationAsyncRounds)

	b.WriteString("# HELP sovereign_aggregation_stale_drops_total Stale model updates dropped by aggregator\n")
	b.WriteString("# TYPE sovereign_aggregation_stale_drops_total gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_stale_drops_total %.0f\n", metrics.aggregationStaleDrops)

	b.WriteString("# HELP sovereign_aggregation_average_latency_ms Average aggregation round latency in milliseconds\n")
	b.WriteString("# TYPE sovereign_aggregation_average_latency_ms gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_average_latency_ms %.0f\n", metrics.aggregationAvgLatencyMS)

	b.WriteString("# HELP sovereign_aggregation_churn_joins_total Total node join churn events\n")
	b.WriteString("# TYPE sovereign_aggregation_churn_joins_total gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_churn_joins_total %.0f\n", metrics.churnJoinsTotal)

	b.WriteString("# HELP sovereign_aggregation_churn_leaves_total Total node leave churn events\n")
	b.WriteString("# TYPE sovereign_aggregation_churn_leaves_total gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_churn_leaves_total %.0f\n", metrics.churnLeavesTotal)

	b.WriteString("# HELP sovereign_aggregation_staleness_avg_seconds Average async staleness in seconds\n")
	b.WriteString("# TYPE sovereign_aggregation_staleness_avg_seconds gauge\n")
	fmt.Fprintf(&b, "sovereign_aggregation_staleness_avg_seconds %.6f\n", metrics.stalenessAvgSeconds)

	b.WriteString("# HELP sovereign_ops_privacy_cumulative_epsilon Cumulative differential privacy epsilon consumed\n")
	b.WriteString("# TYPE sovereign_ops_privacy_cumulative_epsilon gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_privacy_cumulative_epsilon %.6f\n", metrics.privacyCumulativeEpsilon)

	b.WriteString("# HELP sovereign_ops_privacy_epsilon_target Configured differential privacy epsilon target\n")
	b.WriteString("# TYPE sovereign_ops_privacy_epsilon_target gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_privacy_epsilon_target %.6f\n", metrics.privacyEpsilonTarget)

	b.WriteString("# HELP sovereign_ops_fl_straggler_rate_pct Percentage of nodes missing gradient timeout windows\n")
	b.WriteString("# TYPE sovereign_ops_fl_straggler_rate_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_fl_straggler_rate_pct %.4f\n", metrics.stragglerRatePct)

	b.WriteString("# HELP sovereign_ops_attack_success_rate_pct Byzantine attack success rate percentage\n")
	b.WriteString("# TYPE sovereign_ops_attack_success_rate_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_attack_success_rate_pct %.4f\n", metrics.attackSuccessRatePct)

	b.WriteString("# HELP sovereign_ops_detection_precision_pct Detection precision for Byzantine anomaly policy checks\n")
	b.WriteString("# TYPE sovereign_ops_detection_precision_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_detection_precision_pct %.4f\n", metrics.detectionPrecisionPct)

	b.WriteString("# HELP sovereign_ops_tee_epc_utilization_pct TEE enclave page cache utilization percentage\n")
	b.WriteString("# TYPE sovereign_ops_tee_epc_utilization_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_tee_epc_utilization_pct %.4f\n", metrics.epcUtilizationPct)

	b.WriteString("# HELP sovereign_ops_attestation_latency_ms Mean attestation verification latency in milliseconds\n")
	b.WriteString("# TYPE sovereign_ops_attestation_latency_ms gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_attestation_latency_ms %.4f\n", metrics.attestationLatencyMS)

	b.WriteString("# HELP sovereign_ops_cxl_utilization_pct CXL utilization percentage from ops health snapshot\n")
	b.WriteString("# TYPE sovereign_ops_cxl_utilization_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_cxl_utilization_pct %.4f\n", metrics.cxlUtilizationPct)

	b.WriteString("# HELP sovereign_ops_cxl_throughput_gbps CXL throughput estimate in GB/s\n")
	b.WriteString("# TYPE sovereign_ops_cxl_throughput_gbps gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_cxl_throughput_gbps %.4f\n", metrics.cxlThroughputGBps)

	b.WriteString("# HELP sovereign_ops_npu_temp_c NPU die temperature in Celsius\n")
	b.WriteString("# TYPE sovereign_ops_npu_temp_c gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_npu_temp_c %.4f\n", metrics.npuTempC)

	b.WriteString("# HELP sovereign_ops_tpm_temp_c TPM temperature in Celsius\n")
	b.WriteString("# TYPE sovereign_ops_tpm_temp_c gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_tpm_temp_c %.4f\n", metrics.tpmTempC)

	b.WriteString("# HELP sovereign_ops_governance_total_stake Total governance stake observed in the network\n")
	b.WriteString("# TYPE sovereign_ops_governance_total_stake gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_governance_total_stake %.4f\n", metrics.governanceTotalStake)

	b.WriteString("# HELP sovereign_ops_governance_stake_concentration_pct Stake concentration percentage of largest founder stake\n")
	b.WriteString("# TYPE sovereign_ops_governance_stake_concentration_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_governance_stake_concentration_pct %.4f\n", metrics.stakeConcentrationPct)

	b.WriteString("# HELP sovereign_ops_governance_slashing_events_total Total slashing events recorded in governance view\n")
	b.WriteString("# TYPE sovereign_ops_governance_slashing_events_total gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_governance_slashing_events_total %.0f\n", metrics.slashingEventsTotal)

	b.WriteString("# HELP sovereign_ops_governance_reward_apy_pct Governance reward annual percentage yield\n")
	b.WriteString("# TYPE sovereign_ops_governance_reward_apy_pct gauge\n")
	fmt.Fprintf(&b, "sovereign_ops_governance_reward_apy_pct %.4f\n", metrics.rewardAPYPct)

	return b.String()
}

func collectConsensusMetrics(client *http.Client, apiBaseURL string) apiConsensusMetrics {
	metrics := apiConsensusMetrics{}
	base := strings.TrimRight(strings.TrimSpace(apiBaseURL), "/")
	if base == "" {
		return metrics
	}

	statusBody, err := fetchJSON(client, base+"/api/v1/consensus/status")
	if err == nil {
		metrics.consensusUp = 1
		if consensusNode, ok := asMap(statusBody["consensus"]); ok {
			metrics.asyncMode = boolToFloat(asBool(consensusNode["async_mode"]))
			metrics.activeNodes = asFloat(consensusNode["active_node_count"])
			metrics.quorumSize = asFloat(consensusNode["quorum_size"])
			metrics.openRounds = asFloat(consensusNode["open_rounds"])
		}
		if aggregationNode, ok := asMap(statusBody["aggregation"]); ok {
			metrics.aggregationRound = asFloat(aggregationNode["round_number"])
			metrics.bufferedModels = asFloat(aggregationNode["buffered_models"])
			if nested, ok := asMap(aggregationNode["metrics"]); ok {
				metrics.aggregationAsyncRounds = asFloat(nested["async_rounds"])
				metrics.aggregationStaleDrops = asFloat(nested["stale_drops"])
				metrics.aggregationAvgLatencyMS = asFloat(nested["average_latency_ms"])
			}
		}
	}

	generalBody, err := fetchJSON(client, base+"/api/v1/metrics")
	if err == nil {
		metrics.churnJoinsTotal = asFloat(generalBody["churn_joins_total"])
		metrics.churnLeavesTotal = asFloat(generalBody["churn_leaves_total"])
		metrics.stalenessAvgSeconds = asFloat(generalBody["async_staleness_avg_seconds"])
	}

	opsBody, err := fetchJSON(client, base+"/ops/health")
	if err == nil {
		if privacyNode, ok := asMap(opsBody["privacy_security"]); ok {
			metrics.privacyCumulativeEpsilon = asFloat(privacyNode["cumulative_epsilon"])
			metrics.privacyEpsilonTarget = asFloat(privacyNode["epsilon_target"])
			metrics.stragglerRatePct = asFloat(privacyNode["straggler_rate_pct"])
			metrics.attackSuccessRatePct = asFloat(privacyNode["attack_success_rate_pct"])
			metrics.detectionPrecisionPct = asFloat(privacyNode["detection_precision_pct"])
		}

		if teeNode, ok := asMap(opsBody["tee_hardware"]); ok {
			metrics.epcUtilizationPct = asFloat(teeNode["epc_utilization_pct"])
			metrics.attestationLatencyMS = asFloat(teeNode["attestation_latency_ms"])
			metrics.cxlUtilizationPct = asFloat(teeNode["cxl_utilization_pct"])
			metrics.cxlThroughputGBps = asFloat(teeNode["cxl_throughput_gbps"])
			metrics.npuTempC = asFloat(teeNode["npu_temp_c"])
			metrics.tpmTempC = asFloat(teeNode["tpm_temp_c"])
		}

		if govNode, ok := asMap(opsBody["governance_economics"]); ok {
			metrics.governanceTotalStake = asFloat(govNode["total_stake"])
			metrics.stakeConcentrationPct = asFloat(govNode["stake_concentration_pct"])
			metrics.slashingEventsTotal = asFloat(govNode["slashing_events_total"])
			metrics.rewardAPYPct = asFloat(govNode["reward_apy_pct"])
		}
	}

	return metrics
}

func fetchJSON(client *http.Client, url string) (map[string]interface{}, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func asMap(value interface{}) (map[string]interface{}, bool) {
	result, ok := value.(map[string]interface{})
	return result, ok
}

func asFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f
		}
		return 0
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

func asBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		normalized := strings.ToLower(strings.TrimSpace(v))
		return normalized == "true" || normalized == "1" || normalized == "yes" || normalized == "on"
	case float64:
		return v != 0
	case int:
		return v != 0
	default:
		return false
	}
}

func boolToFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}

func combinedMetricsPrometheus(parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(trimmed)
		b.WriteString("\n")
	}
	return b.String()
}
