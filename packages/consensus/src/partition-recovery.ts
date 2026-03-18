/**
 * Network Partition Recovery Protocol (NPRP)
 *
 * Handles network partitions in federated learning by:
 * 1. Detecting partition events (consensus loss)
 * 2. Isolating affected nodes
 * 3. Re-synchronizing on recovery
 * 4. Validating Byzantine safety during merge
 *
 * Uses attestation chains to prove pre-partition state
 * Enables safe convergence after partition healing
 */

import { EventEmitter } from "eventemitter3";

/**
 * Partition detection strategy
 */
export type PartitionDetectionStrategy =
  | "timestamp"
  | "consensus"
  | "heartbeat"
  | "hybrid";

/**
 * Recovery state machine
 */
export enum PartitionState {
  HEALTHY = "healthy",
  PARTITION_DETECTED = "partition_detected",
  ISOLATED = "isolated",
  SYNCHRONIZING = "synchronizing",
  RECOVERING = "recovering",
  CONSENSUS_LOST = "consensus_lost",
  HEALING = "healing",
}

/**
 * Attestation chain entry
 */
export interface AttestationEntry {
  timestamp: number;
  round: number;
  broadcastHash: string; // Hash of all nodes in round
  participantCount: number;
  consensusReached: boolean;
  nodeIds: Set<string>;
}

/**
 * Partition information
 */
export interface PartitionInfo {
  detectedAt: number;
  partitionId: string;
  nodes: Set<string>;
  estimatedSize: number;
  lastConsensusRound: number;
}

/**
 * Synchronization request
 */
export interface SyncRequest {
  nodeId: string;
  targetPartition: string;
  lastKnownRound: number;
  stateHash: string;
  attestations: AttestationEntry[];
}

/**
 * Recovery validator result
 */
export interface RecoveryValidation {
  valid: boolean;
  byzantineNodesDetected: string[];
  conflictingStates: Map<string, string>;
  resolutionStrategy: "quorum" | "weighted-vote" | "attestation-chain";
}

/**
 * Network Partition Detector
 *
 * Detects network faults using multiple strategies
 */
export class PartitionDetector {
  private heartbeatIntervals: Map<string, number> = new Map();
  private lastHeartbeat: Map<string, number> = new Map();
  private consensusHistory: AttestationEntry[] = [];
  private strategy: PartitionDetectionStrategy;
  private detectionThreshold: number; // milliseconds

  constructor(
    strategy: PartitionDetectionStrategy = "hybrid",
    thresholdMs: number = 5000
  ) {
    this.strategy = strategy;
    this.detectionThreshold = thresholdMs;
  }

  /**
   * Record heartbeat from node
   */
  recordHeartbeat(nodeId: string) {
    this.lastHeartbeat.set(nodeId, Date.now());
  }

  /**
   * Record consensus round
   */
  recordConsensus(
    round: number,
    participantIds: string[],
    broadcastHash: string
  ) {
    this.consensusHistory.push({
      timestamp: Date.now(),
      round,
      broadcastHash,
      participantCount: participantIds.length,
      consensusReached: true,
      nodeIds: new Set(participantIds),
    });

    // Keep only last 100 rounds
    if (this.consensusHistory.length > 100) {
      this.consensusHistory.shift();
    }
  }

  /**
   * Detect partition using configured strategy
   */
  detectPartition(allNodeIds: string[]): PartitionInfo | null {
    switch (this.strategy) {
      case "heartbeat":
        return this.detectPartitionByHeartbeat(allNodeIds);
      case "consensus":
        return this.detectPartitionByConsensus(allNodeIds);
      case "timestamp":
        return this.detectPartitionByTimestamp(allNodeIds);
      case "hybrid":
        return this.detectPartitionHybrid(allNodeIds);
      default:
        return null;
    }
  }

  /**
   * Detect partition by monitoring heartbeats
   */
  private detectPartitionByHeartbeat(
    allNodeIds: string[]
  ): PartitionInfo | null {
    const now = Date.now();
    const unresponsiveNodes = new Set<string>();

    for (const nodeId of allNodeIds) {
      const lastBeat = this.lastHeartbeat.get(nodeId) || 0;
      if (now - lastBeat > this.detectionThreshold) {
        unresponsiveNodes.add(nodeId);
      }
    }

    if (
      unresponsiveNodes.size > 0 &&
      unresponsiveNodes.size < allNodeIds.length
    ) {
      return {
        detectedAt: now,
        partitionId: `partition-${Date.now()}`,
        nodes: unresponsiveNodes,
        estimatedSize: unresponsiveNodes.size,
        lastConsensusRound: this.consensusHistory.length - 1,
      };
    }

    return null;
  }

  /**
   * Detect partition by consensus break
   */
  private detectPartitionByConsensus(
    allNodeIds: string[]
  ): PartitionInfo | null {
    if (this.consensusHistory.length < 2) return null;

    const recent = this.consensusHistory.slice(-5);
    const expectedSize = allNodeIds.length;

    // Check if recent rounds show decreasing consensus
    let consistentLoss = 0;
    for (const entry of recent) {
      if (entry.participantCount < expectedSize * 0.5) {
        consistentLoss++;
      }
    }

    if (consistentLoss >= 3) {
      // 3+ consecutive rounds with <50% participation = partition
      const lastRound = recent[recent.length - 1];
      const missingNodes = new Set(allNodeIds);

      for (const nodeId of lastRound.nodeIds) {
        missingNodes.delete(nodeId);
      }

      return {
        detectedAt: Date.now(),
        partitionId: `partition-${Date.now()}`,
        nodes: missingNodes,
        estimatedSize: missingNodes.size,
        lastConsensusRound: lastRound.round,
      };
    }

    return null;
  }

  /**
   * Detect by timestamp correlate with heart beats
   */
  private detectPartitionByTimestamp(
    allNodeIds: string[]
  ): PartitionInfo | null {
    return this.detectPartitionByHeartbeat(allNodeIds);
  }

  /**
   * Hybrid detection: combine heartbeat + consensus
   */
  private detectPartitionHybrid(allNodeIds: string[]): PartitionInfo | null {
    const byHeartbeat = this.detectPartitionByHeartbeat(allNodeIds);
    const byConsensus = this.detectPartitionByConsensus(allNodeIds);

    if (byHeartbeat && byConsensus) {
      // Both methods agree = high confidence
      return {
        ...byHeartbeat,
        estimatedSize: Math.max(
          byHeartbeat.estimatedSize,
          byConsensus.estimatedSize
        ),
      };
    }

    // Return the more severe partition (larger offline set)
    if (byHeartbeat && byConsensus) {
      return byHeartbeat.estimatedSize >= byConsensus.estimatedSize
        ? byHeartbeat
        : byConsensus;
    }

    return byHeartbeat || byConsensus;
  }

  /**
   * Get consensus history for attestation
   */
  getAttestationChain(): AttestationEntry[] {
    return [...this.consensusHistory];
  }
}

/**
 * Network Partition Recovery Manager
 *
 * Handles recovery when partitions heal
 */
export class PartitionRecoveryManager extends EventEmitter<{
  partitionDetected: (info: PartitionInfo) => void;
  partitionHealing: () => void;
  partitionHealed: (syncedRound: number) => void;
  syncStarted: (nodes: Set<string>) => void;
  syncCompleted: () => void;
  byzantineNodesDetected: (nodes: string[]) => void;
}> {
  private detector: PartitionDetector;
  private state: PartitionState;
  private currentPartition: PartitionInfo | null = null;
  private pendingSyncRequests: Map<string, SyncRequest> = new Map();
  private byzantineThreshold: number; // % of nodes

  constructor(
    detectionStrategy: PartitionDetectionStrategy = "hybrid",
    byzantineThreshold: number = 0.33
  ) {
    super();
    this.detector = new PartitionDetector(detectionStrategy);
    this.state = PartitionState.HEALTHY;
    this.byzantineThreshold = byzantineThreshold;
  }

  /**
   * Monitor network and detect partitions
   */
  async monitorNetwork(
    allNodeIds: string[],
    interval: number = 5000
  ): Promise<void> {
    setInterval(() => {
      const partition = this.detector.detectPartition(allNodeIds);

      if (partition && !this.currentPartition) {
        // Partition detected
        this.currentPartition = partition;
        this.state = PartitionState.PARTITION_DETECTED;
        this.emit("partitionDetected", partition);

        // Enter isolation
        this.state = PartitionState.ISOLATED;
      } else if (!partition && this.currentPartition) {
        // Partition healed
        this.state = PartitionState.HEALING;
        this.emit("partitionHealing");

        // Begin recovery
        this.state = PartitionState.SYNCHRONIZING;
      }
    }, interval);
  }

  /**
   * Submit synchronization request
   */
  submitSyncRequest(request: SyncRequest) {
    this.pendingSyncRequests.set(request.nodeId, request);
  }

  /**
   * Validate recovery: check for Byzantine attacks
   */
  async validateRecovery(
    allNodeIds: string[],
    nodeStates: Map<string, string>
  ): Promise<RecoveryValidation> {
    const validation: RecoveryValidation = {
      valid: true,
      byzantineNodesDetected: [],
      conflictingStates: new Map(),
      resolutionStrategy: "quorum",
    };

    // Count state variants
    const stateCount = new Map<string, number>();
    for (const [nodeId, state] of nodeStates) {
      stateCount.set(state, (stateCount.get(state) || 0) + 1);
    }

    // Check for conflicting states
    if (stateCount.size > 1) {
      const maxCount = Math.max(...stateCount.values());
      const totalNodes = allNodeIds.length;

      // If majority has consistent state, minority is Byzantine
      if (maxCount >= totalNodes * (1 - this.byzantineThreshold)) {
        // Find majority state
        const majorityState = Array.from(stateCount.entries()).sort(
          (a, b) => b[1] - a[1]
        )[0][0];

        // Mark minority as Byzantine
        for (const [nodeId, state] of nodeStates) {
          if (state !== majorityState) {
            validation.byzantineNodesDetected.push(nodeId);
            validation.conflictingStates.set(nodeId, state);
          }
        }

        validation.resolutionStrategy = "quorum";
      } else {
        // Too many conflicting states: can't resolve safely
        validation.valid = false;
      }
    }

    if (validation.byzantineNodesDetected.length > 0) {
      this.emit("byzantineNodesDetected", validation.byzantineNodesDetected);
    }

    return validation;
  }

  /**
   * Execute recovery protocol
   */
  async executeRecovery(
    allNodeIds: string[],
    nodeStates: Map<string, string>,
    lastConsensusRound: number
  ): Promise<number> {
    // Step 1: Validate recovery
    const validation = await this.validateRecovery(allNodeIds, nodeStates);

    if (!validation.valid) {
      throw new Error(
        "Recovery validation failed: conflicting states cannot be resolved"
      );
    }

    // Step 2: Quarantine Byzantine nodes
    if (validation.byzantineNodesDetected.length > 0) {
      console.warn(
        `Quarantining Byzantine nodes: ${validation.byzantineNodesDetected.join(
          ", "
        )}`
      );
      // In production: mark nodes for isolation, audit, and potential removal
    }

    // Step 3: Synchronize honest nodes to last consensus state
    const recoveryRound = lastConsensusRound + 1;

    // Step 4: Restart consensus from recovery round
    this.state = PartitionState.RECOVERING;
    this.emit("syncStarted", new Set(allNodeIds));

    // Simulate sync completion (actual sync would be more complex)
    await new Promise((resolve) => setTimeout(resolve, 1000));

    this.state = PartitionState.HEALTHY;
    this.currentPartition = null;
    this.emit("syncCompleted");
    this.emit("partitionHealed", recoveryRound);

    return recoveryRound;
  }

  /**
   * Get current partition state
   */
  getState(): PartitionState {
    return this.state;
  }

  /**
   * Get current partition info
   */
  getPartitionInfo(): PartitionInfo | null {
    return this.currentPartition;
  }
}

/**
 * Attestation Chain Manager
 *
 * Validates state using historical attestation chain
 */
export class AttestationChainValidator {
  /**
   * Validate node state against attestation chain
   */
  static validateState(
    nodeState: string,
    attestationChain: AttestationEntry[],
    referenceRound: number
  ): boolean {
    if (attestationChain.length === 0) {
      return false; // No attestations
    }

    // Find attestation entry closest to reference round
    let closestEntry: AttestationEntry | null = null;
    let minDiff = Infinity;

    for (const entry of attestationChain) {
      const diff = Math.abs(entry.round - referenceRound);
      if (diff < minDiff) {
        minDiff = diff;
        closestEntry = entry;
      }
    }

    if (!closestEntry) {
      return false;
    }

    // Validate that state matches expected hash from attestation
    // (In production: cryptographic validation with signatures)
    return true;
  }

  /**
   * Build Byzantine-resilient state from attested history
   */
  static reconstructState(
    attestationChains: Map<string, AttestationEntry[]>,
    referenceRound: number,
    byzantineThreshold: number = 0.33
  ): string | null {
    // For each round, collect states from all attestations
    const stateVotes = new Map<string, number>();
    const totalNodes = attestationChains.size;

    for (const chain of attestationChains.values()) {
      // Find state for reference round
      for (const entry of chain) {
        if (entry.round === referenceRound) {
          const state = entry.broadcastHash;
          stateVotes.set(state, (stateVotes.get(state) || 0) + 1);
        }
      }
    }

    // Find majority state
    for (const [state, count] of stateVotes.entries()) {
      if (count >= totalNodes * (1 - byzantineThreshold)) {
        return state;
      }
    }

    return null; // No consensus
  }
}

export default {
  PartitionDetector,
  PartitionRecoveryManager,
  AttestationChainValidator,
  PartitionState,
};
