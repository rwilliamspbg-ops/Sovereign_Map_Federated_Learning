// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"
	"sync"
)

const (
	// Reputation scores are in basis points [0, 10000].
	defaultReputationScore  uint32 = 7000
	defaultAttestationScore uint32 = 5000
	defaultQualityScore     uint32 = 7000
	maxReputationScore      uint32 = 10000
	minSelectionReputation  uint32 = 100
)

// ReputationPolicy controls governance-tunable reputation behavior.
type ReputationPolicy struct {
	SlashPenalty                      uint32 `json:"slash_penalty"`
	RewardGain                        uint32 `json:"reward_gain"`
	EpochRecovery                     uint32 `json:"epoch_recovery"`
	ReputationWeight                  uint32 `json:"reputation_weight"`
	AttestationWeight                 uint32 `json:"attestation_weight"`
	QualityWeight                     uint32 `json:"quality_weight"`
	MaxAttestationAgeBlocks           uint64 `json:"max_attestation_age_blocks"`
	InvalidAttestationPenalty         uint32 `json:"invalid_attestation_penalty"`
	InvalidAttestationSlashBps        uint32 `json:"invalid_attestation_slash_bps"`
	StaleAttestationSlashBps          uint32 `json:"stale_attestation_slash_bps"`
	MaxConsecutiveAttestationFailures uint32 `json:"max_consecutive_attestation_failures"`
	MinQualityScore                   uint32 `json:"min_quality_score"`
	QualityPenalty                    uint32 `json:"quality_penalty"`
	MaxConsecutiveLowQuality          uint32 `json:"max_consecutive_low_quality"`
}

type ValidatorSetMetrics struct {
	ValidatorCount           int    `json:"validator_count"`
	TotalStake               uint64 `json:"total_stake"`
	AverageReputation        uint32 `json:"average_reputation"`
	AverageAttestation       uint32 `json:"average_attestation"`
	AverageQuality           uint32 `json:"average_quality"`
	LowReputationCount       int    `json:"low_reputation_count"`
	LowAttestationCount      int    `json:"low_attestation_count"`
	StaleAttestationCount    int    `json:"stale_attestation_count"`
	MissingAttestationCount  int    `json:"missing_attestation_count"`
	TotalAttestationFailures uint64 `json:"total_attestation_failures"`
	JailedCount              int    `json:"jailed_count"`
}

// Validator represents a consensus validator in the network
type Validator struct {
	NodeID                         string `json:"node_id"`
	StakedAmount                   uint64 `json:"staked_amount"`
	ReputationScore                uint32 `json:"reputation_score"`
	AttestationScore               uint32 `json:"attestation_score"`
	ParticipationScore             uint32 `json:"participation_score"`
	AccumulatedRewards             uint64 `json:"accumulated_rewards"`
	AccumulatedVotes               uint64 `json:"accumulated_votes"`
	SlashCount                     uint32 `json:"slash_count"`
	LowQualityStreak               uint32 `json:"low_quality_streak"`
	ConsecutiveAttestationFailures uint32 `json:"consecutive_attestation_failures"`
	AttestationFailuresTotal       uint32 `json:"attestation_failures_total"`
	AttestationStale               bool   `json:"attestation_stale"`
	LastAttestationBlock           uint64 `json:"last_attestation_block"`
	LastValidAttestationBlock      uint64 `json:"last_valid_attestation_block"`
	LastAttestationEvidence        string `json:"last_attestation_evidence"`
	Jailed                         bool   `json:"jailed"`
	CommissionRate                 uint32 `json:"commission_rate"` // basis points (0-10000)
	LastActiveBlock                uint64 `json:"last_active_block"`
	JailedUntilBlock               uint64 `json:"jailed_until_block"`
}

// ValidatorSet manages active consensus validators
type ValidatorSet struct {
	mu              sync.RWMutex
	Validators      map[string]*Validator
	TotalStake      uint64
	EpochNumber     uint64
	EpochStartBlock uint64
	EpochDuration   uint64
	MinValidators   int
	MaxValidators   int
	MinStakeAmount  uint64
	Policy          ReputationPolicy
}

// NewValidatorSet creates a new validator set
func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		Validators:      make(map[string]*Validator),
		TotalStake:      0,
		EpochNumber:     0,
		EpochStartBlock: 0,
		EpochDuration:   1000, // blocks per epoch
		MinValidators:   4,
		MaxValidators:   100,
		MinStakeAmount:  1000000, // minimum stake to become validator
		Policy: ReputationPolicy{
			SlashPenalty:                      500,
			RewardGain:                        50,
			EpochRecovery:                     25,
			ReputationWeight:                  60,
			AttestationWeight:                 25,
			QualityWeight:                     15,
			MaxAttestationAgeBlocks:           500,
			InvalidAttestationPenalty:         350,
			InvalidAttestationSlashBps:        100,
			StaleAttestationSlashBps:          50,
			MaxConsecutiveAttestationFailures: 3,
			MinQualityScore:                   5000,
			QualityPenalty:                    200,
			MaxConsecutiveLowQuality:          5,
		},
	}
}

// AddValidator adds a new validator to the set
func (vs *ValidatorSet) AddValidator(nodeID string, stakeAmount uint64) error {
	if stakeAmount < vs.MinStakeAmount {
		return fmt.Errorf("stake amount %d below minimum %d", stakeAmount, vs.MinStakeAmount)
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	if _, exists := vs.Validators[nodeID]; exists {
		return fmt.Errorf("validator already exists: %s", nodeID)
	}

	vs.Validators[nodeID] = &Validator{
		NodeID:             nodeID,
		StakedAmount:       stakeAmount,
		ReputationScore:    defaultReputationScore,
		AttestationScore:   defaultAttestationScore,
		ParticipationScore: defaultQualityScore,
		CommissionRate:     1000, // 10%
	}
	vs.TotalStake += stakeAmount

	return nil
}

// Stake adds stake to an existing or new validator
func (vs *ValidatorSet) Stake(nodeID string, amount uint64) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	validator, exists := vs.Validators[nodeID]
	if !exists {
		if amount < vs.MinStakeAmount {
			return fmt.Errorf("initial stake amount %d below minimum %d", amount, vs.MinStakeAmount)
		}
		validator = &Validator{
			NodeID:             nodeID,
			StakedAmount:       amount,
			ReputationScore:    defaultReputationScore,
			AttestationScore:   defaultAttestationScore,
			ParticipationScore: defaultQualityScore,
			CommissionRate:     1000,
		}
		vs.Validators[nodeID] = validator
	} else {
		validator.StakedAmount += amount
	}
	vs.TotalStake += amount

	return nil
}

// Unstake removes stake from a validator
func (vs *ValidatorSet) Unstake(nodeID string, amount uint64) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	validator, exists := vs.Validators[nodeID]
	if !exists {
		return fmt.Errorf("validator not found: %s", nodeID)
	}

	if validator.StakedAmount < amount {
		return fmt.Errorf("insufficient stake: %d < %d", validator.StakedAmount, amount)
	}

	validator.StakedAmount -= amount
	vs.TotalStake -= amount

	// Remove validator if below minimum stake
	if validator.StakedAmount < vs.MinStakeAmount {
		delete(vs.Validators, nodeID)
	}

	return nil
}

// SlashValidator reduces validator stake for misbehavior
func (vs *ValidatorSet) SlashValidator(nodeID string, penalty uint64) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	validator, exists := vs.Validators[nodeID]
	if !exists {
		return fmt.Errorf("validator not found: %s", nodeID)
	}

	slashAmount := penalty
	if slashAmount > validator.StakedAmount {
		slashAmount = validator.StakedAmount
	}

	validator.StakedAmount -= slashAmount
	vs.TotalStake -= slashAmount
	validator.SlashCount++
	if validator.ReputationScore > vs.Policy.SlashPenalty {
		validator.ReputationScore -= vs.Policy.SlashPenalty
	} else {
		validator.ReputationScore = minSelectionReputation
	}

	// Jail validator if too many violations
	if validator.SlashCount > 3 {
		validator.Jailed = true
		validator.JailedUntilBlock = vs.EpochStartBlock + vs.EpochDuration
	}

	return nil
}

// SelectValidators selects the next validators weighted by stake
// Uses the Verifiable Random Function (VRF) pattern
func (vs *ValidatorSet) SelectValidators(count int) []*Validator {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	// Filter active validators
	active := make([]*Validator, 0)
	for _, v := range vs.Validators {
		if !v.Jailed && v.StakedAmount >= vs.MinStakeAmount {
			active = append(active, v)
		}
	}

	if len(active) == 0 {
		return []*Validator{}
	}

	if count > len(active) {
		count = len(active)
	}

	// Weighted random selection
	selected := make([]*Validator, 0, count)
	used := make(map[string]bool)

	for len(selected) < count {
		// Select validator weighted by stake
		validator := vs.selectWeighted(active, used)
		if validator == nil {
			break
		}
		selected = append(selected, validator)
		used[validator.NodeID] = true
	}

	return selected
}

// selectWeighted performs weighted random selection
func (vs *ValidatorSet) selectWeighted(validators []*Validator, used map[string]bool) *Validator {
	totalStake := uint64(0)
	available := make([]*Validator, 0)

	for _, v := range validators {
		if !used[v.NodeID] {
			available = append(available, v)
			totalStake += vs.effectiveWeight(v)
		}
	}

	if len(available) == 0 {
		return nil
	}

	// Select random value between 0 and totalStake
	r, err := cryptoRandomUint64(totalStake)
	if err != nil {
		// Fall back to deterministic selection if secure randomness fails.
		return available[0]
	}
	accumulated := uint64(0)

	for _, v := range available {
		accumulated += vs.effectiveWeight(v)
		if r < accumulated {
			return v
		}
	}

	return available[len(available)-1]
}

// DistributeRewards distributes block rewards to validators
func (vs *ValidatorSet) DistributeRewards(blockHeight uint64, baseReward uint64, validators []string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if len(validators) == 0 {
		return
	}

	rewardPerValidator := baseReward / uint64(len(validators))

	for _, nodeID := range validators {
		if v, exists := vs.Validators[nodeID]; exists {
			score := vs.compositeScore(v)
			adjustedReward := rewardPerValidator * uint64(score) / uint64(maxReputationScore)
			if adjustedReward == 0 && rewardPerValidator > 0 {
				adjustedReward = 1
			}
			v.AccumulatedRewards += adjustedReward
			v.AccumulatedVotes++
			v.LastActiveBlock = blockHeight
			if v.ReputationScore+vs.Policy.RewardGain > maxReputationScore {
				v.ReputationScore = maxReputationScore
			} else {
				v.ReputationScore += vs.Policy.RewardGain
			}
		}
	}
}

func (vs *ValidatorSet) effectiveWeight(v *Validator) uint64 {
	return v.StakedAmount * uint64(vs.compositeScore(v)) / uint64(maxReputationScore)
}

func (vs *ValidatorSet) compositeScore(v *Validator) uint32 {
	rep := v.ReputationScore
	if rep < minSelectionReputation {
		rep = minSelectionReputation
	}
	att := v.AttestationScore
	if att < minSelectionReputation {
		att = minSelectionReputation
	}
	quality := v.ParticipationScore
	if quality < minSelectionReputation {
		quality = minSelectionReputation
	}

	totalWeight := vs.Policy.ReputationWeight + vs.Policy.AttestationWeight + vs.Policy.QualityWeight
	if totalWeight == 0 {
		return rep
	}
	weighted := uint64(rep)*uint64(vs.Policy.ReputationWeight) +
		uint64(att)*uint64(vs.Policy.AttestationWeight) +
		uint64(quality)*uint64(vs.Policy.QualityWeight)

	return toUint32Saturating(weighted / uint64(totalWeight))
}

// SetAttestationScore updates attestation trust score for a validator.
func (vs *ValidatorSet) SetAttestationScore(nodeID string, score uint32) error {
	if score > maxReputationScore {
		return fmt.Errorf("attestation score out of range: %d", score)
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	v, exists := vs.Validators[nodeID]
	if !exists {
		return fmt.Errorf("validator not found: %s", nodeID)
	}
	v.AttestationScore = score
	v.LastAttestationBlock = vs.EpochStartBlock
	v.LastValidAttestationBlock = vs.EpochStartBlock
	v.AttestationStale = false
	return nil
}

// RecordAttestationEvidence records attestation proof results and applies policy penalties.
func (vs *ValidatorSet) RecordAttestationEvidence(nodeID string, score uint32, evidenceRef string, blockHeight uint64, valid bool) error {
	if score > maxReputationScore {
		return fmt.Errorf("attestation score out of range: %d", score)
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	v, exists := vs.Validators[nodeID]
	if !exists {
		return fmt.Errorf("validator not found: %s", nodeID)
	}

	v.LastAttestationBlock = blockHeight
	v.LastAttestationEvidence = evidenceRef

	if valid {
		v.AttestationScore = score
		v.LastValidAttestationBlock = blockHeight
		v.AttestationStale = false
		v.ConsecutiveAttestationFailures = 0
		return nil
	}

	v.AttestationStale = true
	v.AttestationFailuresTotal++
	v.ConsecutiveAttestationFailures++
	if score < minSelectionReputation {
		v.AttestationScore = minSelectionReputation
	} else {
		v.AttestationScore = score
	}

	if v.ReputationScore > vs.Policy.InvalidAttestationPenalty {
		v.ReputationScore -= vs.Policy.InvalidAttestationPenalty
	} else {
		v.ReputationScore = minSelectionReputation
	}

	vs.applySlash(v, vs.Policy.InvalidAttestationSlashBps)
	if v.ConsecutiveAttestationFailures >= vs.Policy.MaxConsecutiveAttestationFailures {
		v.Jailed = true
		v.JailedUntilBlock = vs.EpochStartBlock + vs.EpochDuration
	}

	return nil
}

// EnforceAttestationFreshness applies stale-attestation policy at epoch/block boundaries.
func (vs *ValidatorSet) EnforceAttestationFreshness(currentBlock uint64) []string {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if vs.Policy.MaxAttestationAgeBlocks == 0 {
		return nil
	}

	slashed := make([]string, 0)
	for _, v := range vs.Validators {
		if v.Jailed || v.StakedAmount < vs.MinStakeAmount {
			continue
		}

		age := currentBlock
		if v.LastValidAttestationBlock > 0 && currentBlock >= v.LastValidAttestationBlock {
			age = currentBlock - v.LastValidAttestationBlock
		}

		if age > vs.Policy.MaxAttestationAgeBlocks {
			if !v.AttestationStale {
				vs.applySlash(v, vs.Policy.StaleAttestationSlashBps)
				slashed = append(slashed, v.NodeID)
			}
			v.AttestationStale = true
		}
	}

	return slashed
}

func (vs *ValidatorSet) applySlash(v *Validator, slashBps uint32) {
	if v == nil || slashBps == 0 || v.StakedAmount == 0 {
		return
	}
	slashAmount := v.StakedAmount * uint64(slashBps) / 10000
	if slashAmount == 0 {
		slashAmount = 1
	}
	if slashAmount > v.StakedAmount {
		slashAmount = v.StakedAmount
	}
	v.StakedAmount -= slashAmount
	if vs.TotalStake >= slashAmount {
		vs.TotalStake -= slashAmount
	} else {
		vs.TotalStake = 0
	}
}

// RecordParticipationQuality updates anti-gaming quality signals for a validator.
func (vs *ValidatorSet) RecordParticipationQuality(nodeID string, qualityScore uint32) error {
	if qualityScore > maxReputationScore {
		return fmt.Errorf("quality score out of range: %d", qualityScore)
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	v, exists := vs.Validators[nodeID]
	if !exists {
		return fmt.Errorf("validator not found: %s", nodeID)
	}

	if v.ParticipationScore == 0 {
		v.ParticipationScore = qualityScore
	} else {
		// Smooth updates to avoid single-round gaming spikes.
		v.ParticipationScore = toUint32Saturating((uint64(v.ParticipationScore)*4 + uint64(qualityScore)) / 5)
	}

	if qualityScore < vs.Policy.MinQualityScore {
		v.LowQualityStreak++
		if v.ReputationScore > vs.Policy.QualityPenalty {
			v.ReputationScore -= vs.Policy.QualityPenalty
		} else {
			v.ReputationScore = minSelectionReputation
		}
		if v.LowQualityStreak >= vs.Policy.MaxConsecutiveLowQuality {
			v.Jailed = true
			v.JailedUntilBlock = vs.EpochStartBlock + vs.EpochDuration
		}
	} else {
		v.LowQualityStreak = 0
	}

	return nil
}

// SetReputationPolicy applies governance-controlled reputation policy parameters.
func (vs *ValidatorSet) SetReputationPolicy(policy ReputationPolicy) error {
	if policy.SlashPenalty == 0 || policy.RewardGain == 0 || policy.EpochRecovery == 0 {
		return fmt.Errorf("policy penalties and gains must be > 0")
	}
	if policy.MinQualityScore > maxReputationScore || policy.QualityPenalty > maxReputationScore {
		return fmt.Errorf("policy quality values out of range")
	}
	if policy.InvalidAttestationPenalty > maxReputationScore {
		return fmt.Errorf("invalid attestation penalty out of range")
	}
	if policy.InvalidAttestationSlashBps > 10000 || policy.StaleAttestationSlashBps > 10000 {
		return fmt.Errorf("attestation slash bps out of range")
	}
	if policy.MaxAttestationAgeBlocks == 0 {
		return fmt.Errorf("max attestation age blocks must be > 0")
	}
	if policy.MaxConsecutiveAttestationFailures == 0 {
		return fmt.Errorf("max consecutive attestation failures must be > 0")
	}
	if policy.ReputationWeight+policy.AttestationWeight+policy.QualityWeight == 0 {
		return fmt.Errorf("policy weights must sum to > 0")
	}
	if policy.MaxConsecutiveLowQuality == 0 {
		return fmt.Errorf("max consecutive low quality must be > 0")
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.Policy = policy
	return nil
}

// GetReputationPolicy returns the active governance policy.
func (vs *ValidatorSet) GetReputationPolicy() ReputationPolicy {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.Policy
}

// GetMetrics returns aggregate validator metrics for APIs and dashboards.
func (vs *ValidatorSet) GetMetrics() ValidatorSetMetrics {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	metrics := ValidatorSetMetrics{
		ValidatorCount: len(vs.Validators),
		TotalStake:     vs.TotalStake,
	}
	if len(vs.Validators) == 0 {
		return metrics
	}

	var repSum uint64
	var attSum uint64
	var qualitySum uint64
	for _, v := range vs.Validators {
		repSum += uint64(v.ReputationScore)
		attSum += uint64(v.AttestationScore)
		qualitySum += uint64(v.ParticipationScore)
		if v.ReputationScore < 4000 {
			metrics.LowReputationCount++
		}
		if v.AttestationScore < 4000 {
			metrics.LowAttestationCount++
		}
		if v.AttestationStale {
			metrics.StaleAttestationCount++
		}
		if v.LastValidAttestationBlock == 0 {
			metrics.MissingAttestationCount++
		}
		metrics.TotalAttestationFailures += uint64(v.AttestationFailuresTotal)
		if v.Jailed {
			metrics.JailedCount++
		}
	}

	count := uint64(len(vs.Validators))
	metrics.AverageReputation = toUint32Saturating(repSum / count)
	metrics.AverageAttestation = toUint32Saturating(attSum / count)
	metrics.AverageQuality = toUint32Saturating(qualitySum / count)
	return metrics
}

// GetValidator returns a validator by node ID
func (vs *ValidatorSet) GetValidator(nodeID string) (*Validator, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	v, exists := vs.Validators[nodeID]
	if !exists {
		return nil, fmt.Errorf("validator not found: %s", nodeID)
	}
	return v, nil
}

// GetValidators returns all validators
func (vs *ValidatorSet) GetValidators() []*Validator {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	result := make([]*Validator, 0, len(vs.Validators))
	for _, v := range vs.Validators {
		result = append(result, v)
	}

	// Sort by stake descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].StakedAmount > result[j].StakedAmount
	})

	return result
}

// Count returns the number of validators
func (vs *ValidatorSet) Count() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return len(vs.Validators)
}

// RotateEpoch advances to next epoch and unjails eligible validators
func (vs *ValidatorSet) RotateEpoch(currentBlock uint64) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	vs.EpochNumber++
	vs.EpochStartBlock = currentBlock

	// Unjail validators whose jail period has expired
	for _, v := range vs.Validators {
		if v.Jailed && currentBlock > v.JailedUntilBlock {
			v.Jailed = false
			v.SlashCount = 0
		}
		if v.ReputationScore < maxReputationScore {
			next := v.ReputationScore + vs.Policy.EpochRecovery
			if next > maxReputationScore {
				v.ReputationScore = maxReputationScore
			} else {
				v.ReputationScore = next
			}
		}
	}
}

func cryptoRandomUint64(max uint64) (uint64, error) {
	if max == 0 {
		return 0, fmt.Errorf("max must be greater than zero")
	}
	n, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetUint64(max))
	if err != nil {
		return 0, err
	}
	return n.Uint64(), nil
}

func toUint32Saturating(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
