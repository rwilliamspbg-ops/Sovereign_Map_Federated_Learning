import { EventEmitter } from 'eventemitter3';
import { nanoid } from 'nanoid';
import { z } from 'zod';
import { NetworkClient } from './network.js';
import { PrivacyEngine } from '@sovereignmap/privacy';
import { ConsensusParticipant } from '@sovereignmap/consensus';
import { IslandModeManager } from '@sovereignmap/island';
import { 
  SovereignMapError, 
  NodeInitializationError,
  PrivacyBudgetExceededError 
} from './errors.js';

/**
 * Node configuration schema with SGP-001 validation
 */
export const NodeConfigSchema = z.object({
  nodeId: z.string().min(1).max(64).default(() => `node-${nanoid(8)}`),
  region: z.string().min(2).max(64),
  coordinates: z.object({
    lat: z.number().min(-90).max(90),
    lng: z.number().min(-180).max(180),
  }),
  
  // SGP-001 Privacy Configuration
  privacyBudget: z.object({
    epsilon: z.number().min(0.1).max(10.0).default(1.0),
    delta: z.number().min(1e-10).max(1e-3).default(1e-5),
    mechanism: z.enum(['gaussian', 'laplace']).default('gaussian'),
  }).default({}),
  
  // Hardware capabilities
  hardware: z.object({
    npuTops: z.number().min(0).default(0),
    tpmAvailable: z.boolean().default(false),
    maxMemoryMB: z.number().min(512).default(4096),
  }).default({}),
  
  // Network configuration
  network: z.object({
    bootstrapPeers: z.array(z.string()).default([]),
    maxConnections: z.number().min(10).max(1000).default(50),
    enableRelay: z.boolean().default(true),
  }).default({}),
  
  // Island Mode settings
  islandMode: z.object({
    enabled: z.boolean().default(true),
    storagePath: z.string().default('./island-storage'),
    maxOfflineHours: z.number().min(1).max(168).default(24),
  }).default({}),
  
  // Consensus participation
  consensus: z.object({
    enabled: z.boolean().default(true),
    aggregatorTier: z.enum(['none', 'edge', 'regional', 'global']).default('none'),
    maxAggregationChildren: z.number().min(0).max(200).default(0),
  }).default({}),
});

export type NodeConfig = z.infer<typeof NodeConfigSchema>;

/**
 * Node operational states
 */
export enum NodeState {
  INITIALIZING = 'initializing',
  ATTESTING = 'attesting',      // TPM attestation in progress
  CONNECTING = 'connecting',     // Joining network
  SYNCING = 'syncing',           // Catching up with chain
  ONLINE = 'online',             // Fully operational
  ISLAND_MODE = 'island_mode',   // Operating autonomously
  AGGREGATING = 'aggregating',   // Acting as MOHAWK aggregator
  CONSENSUS = 'consensus',       // Participating in BFT
  OFFLINE = 'offline',           // Disconnected
  ERROR = 'error',               // Fault state
  SHUTTING_DOWN = 'shutting_down',
}

/**
 * Sovereign Map Genesis Node
 * 
 * Core client for participating in the decentralized mapping network
 * with automatic privacy preservation and Byzantine fault tolerance.
 */
export class SovereignNode extends EventEmitter<{
  stateChange: (from: NodeState, to: NodeState) => void;
  privacyBudgetUpdate: (remaining: number, total: number) => void;
  connectivityLost: () => void;
  connectivityRestored: (stats: SyncStats) => void;
  mapUpdateSubmitted: (updateId: string, proof: string) => void;
  byzantineFaultDetected: (nodeId: string, faultType: string) => void;
  rewardEarned: (amount: number, reason: string) => void;
}> {
  
  public readonly config: NodeConfig;
  public state: NodeState = NodeState.INITIALIZING;
  public readonly id: string;
  
  private network: NetworkClient;
  private privacy: PrivacyEngine;
  private consensus: ConsensusParticipant | null = null;
  private island: IslandModeManager;
  private startTime: Date;
  private metrics: NodeMetrics;
  
  constructor(config: Partial<NodeConfig> = {}) {
    super();
    
    // Validate and apply defaults
    this.config = NodeConfigSchema.parse(config);
    this.id = this.config.nodeId;
    this.startTime = new Date();
    this.metrics = new NodeMetrics();
    
    // Initialize subsystems
    this.privacy = new PrivacyEngine(this.config.privacyBudget);
    this.network = new NetworkClient(this.config.network, this.id);
    this.island = new IslandModeManager(this.config.islandMode);
    
    // Setup event handlers
    this.setupEventHandlers();
  }
  
  /**
   * Initialize the node with full attestation and network join
   */
  async initialize(): Promise<void> {
    try {
      this.setState(NodeState.INITIALIZING);
      
      // Step 1: Hardware attestation (if TPM available)
      if (this.config.hardware.tpmAvailable) {
        this.setState(NodeState.ATTESTING);
        await this.performAttestation();
      }
      
      // Step 2: Initialize privacy engine
      await this.privacy.initialize();
      
      // Step 3: Setup Island Mode storage
      await this.island.initialize();
      
      // Step 4: Connect to network
      this.setState(NodeState.CONNECTING);
      await this.network.connect();
      
      // Step 5: Sync state
      this.setState(NodeState.SYNCING);
      await this.syncState();
      
      // Step 6: Setup consensus if enabled
      if (this.config.consensus.enabled) {
        this.consensus = new ConsensusParticipant(this.id, this.network);
        await this.consensus.initialize();
      }
      
      this.setState(NodeState.ONLINE);
      this.emit('ready', this.getStatus());
      
    } catch (error) {
      this.setState(NodeState.ERROR);
      throw new NodeInitializationError(
        `Failed to initialize node ${this.id}: ${error.message}`,
        { cause: error }
      );
    }
  }
  
  /**
   * Submit a map update with automatic SGP-001 privacy protection
   */
  async submitMapUpdate(update: MapUpdate): Promise<SubmissionResult> {
    // Check privacy budget
    if (!this.privacy.hasBudgetFor(update)) {
      throw new PrivacyBudgetExceededError(
        `Privacy budget exhausted (ε=${this.config.privacyBudget.epsilon})`
      );
    }
    
    // Apply differential privacy
    const privatizedUpdate = await this.privacy.apply(update);
    
    // Generate cryptographic proof
    const proof = await this.generateProof(privatizedUpdate);
    
    // Submit based on connectivity
    let result: SubmissionResult;
    
    if (this.network.isConnected()) {
      result = await this.submitOnline(privatizedUpdate, proof);
    } else {
      result = await this.island.queueUpdate(privatizedUpdate, proof);
      this.setState(NodeState.ISLAND_MODE);
    }
    
    this.emit('mapUpdateSubmitted', result.updateId, proof);
    this.metrics.recordUpdate(update);
    
    return result;
  }
  
  /**
   * Participate in federated learning round
   */
  async participateInRound(round: FLRound): Promise<RoundResult> {
    if (!this.consensus) {
      throw new SovereignMapError('Consensus not enabled for this node');
    }
    
    // Train local model
    const localUpdate = await this.trainLocalModel(round);
    
    // Submit with BFT verification
    return this.consensus.submitUpdate(round.id, localUpdate);
  }
  
  /**
   * Enter aggregator mode (MOHAWK protocol)
   */
  async becomeAggregator(config: AggregatorConfig): Promise<void> {
    if (this.config.consensus.aggregatorTier === 'none') {
      throw new SovereignMapError('Node not configured as aggregator');
    }
    
    this.setState(NodeState.AGGREGATING);
    
    const aggregator = new HierarchicalAggregator({
      nodeId: this.id,
      tier: this.config.consensus.aggregatorTier,
      maxChildren: this.config.consensus.maxAggregationChildren,
      ...config,
    });
    
    await aggregator.start();
    
    // Listen for child updates
    aggregator.on('aggregate', async (updates) => {
      const zkProof = await this.generateAggregationProof(updates);
      await this.submitAggregate(updates, zkProof);
    });
  }
  
  /**
   * Graceful shutdown with state preservation
   */
  async shutdown(): Promise<void> {
    this.setState(NodeState.SHUTTING_DOWN);
    
    // Persist pending updates
    await this.island.flush();
    
    // Leave consensus gracefully
    if (this.consensus) {
      await this.consensus.leave();
    }
    
    // Disconnect network
    await this.network.disconnect();
    
    // Cleanup privacy engine
    await this.privacy.destroy();
    
    this.setState(NodeState.OFFLINE);
  }
  
  /**
   * Get comprehensive node status
   */
  getStatus(): NodeStatus {
    return {
      id: this.id,
      state: this.state,
      uptime: Date.now() - this.startTime.getTime(),
      privacy: this.privacy.getStatus(),
      network: this.network.getStatus(),
      island: this.island.getStatus(),
      metrics: this.metrics.getSnapshot(),
      hardware: this.config.hardware,
    };
  }
  
  // Private methods
  
  private setState(newState: NodeState): void {
    const oldState = this.state;
    this.state = newState;
    this.emit('stateChange', oldState, newState);
  }
  
  private setupEventHandlers(): void {
    // Network connectivity
    this.network.on('disconnected', () => {
      this.emit('connectivityLost');
      this.setState(NodeState.ISLAND_MODE);
    });
    
    this.network.on('reconnected', async () => {
      const stats = await this.island.sync();
      this.emit('connectivityRestored', stats);
      this.setState(NodeState.ONLINE);
    });
    
    // Byzantine fault detection
    this.network.on('byzantineDetected', (nodeId, faultType) => {
      this.emit('byzantineFaultDetected', nodeId, faultType);
      this.metrics.recordByzantineFault(nodeId, faultType);
    });
    
    // Privacy budget tracking
    this.privacy.on('budgetUpdate', (remaining, total) => {
      this.emit('privacyBudgetUpdate', remaining, total);
    });
  }
  
  private async performAttestation(): Promise<AttestationResult> {
    const { HardwareAttestation } = await import('@sovereignmap/hardware');
    const attestor = new HardwareAttestation();
    return attestor.attest(this.id);
  }
  
  private async generateProof(update: PrivatizedUpdate): Promise<string> {
    // Use WASM-wrapped gnark for 10ms proof generation
    const { ZKProver } = await import('@sovereignmap/privacy/zk');
    const prover = new ZKProver();
    return prover.generateProof(update);
  }
  
  private async submitOnline(
    update: PrivatizedUpdate, 
    proof: string
  ): Promise<SubmissionResult> {
    return this.network.broadcast({
      type: 'MAP_UPDATE',
      payload: update,
      proof,
      timestamp: Date.now(),
      nodeId: this.id,
    });
  }
  
  private async syncState(): Promise<void> {
    // Sync with network for latest chain state
    const state = await this.network.getState();
    await this.island.syncWithState(state);
  }
}

// Types

interface MapUpdate {
  location: { lat: number; lng: number; alt?: number };
  pointCloud?: Uint8Array;
  imagery?: Uint8Array;
  timestamp: number;
  quality: number; // 0-1 IoU score
}

interface PrivatizedUpdate extends MapUpdate {
  noiseApplied: boolean;
  privacyProof: string;
  epsilonConsumed: number;
}

interface SubmissionResult {
  updateId: string;
  status: 'confirmed' | 'pending' | 'queued';
  proof: string;
  estimatedConfirmationTime?: number;
}

interface SyncStats {
  updatesSynced: number;
  updatesQueued: number;
  conflictsResolved: number;
  syncDuration: number;
}

interface FLRound {
  id: string;
  globalModelHash: string;
  trainingConfig: object;
  deadline: number;
}

interface RoundResult {
  accepted: boolean;
  reward: number;
  roundId: string;
}

interface AggregatorConfig {
  regionBoundary?: GeoPolygon;
  specialization?: 'urban' | 'highway' | 'offroad';
}

interface NodeStatus {
  id: string;
  state: NodeState;
  uptime: number;
  privacy: PrivacyStatus;
  network: NetworkStatus;
  island: IslandStatus;
  metrics: MetricsSnapshot;
  hardware: NodeConfig['hardware'];
}

class NodeMetrics {
  private updates: number = 0;
  private byzantineFaults: Map<string, number> = new Map();
  
  recordUpdate(update: MapUpdate): void {
    this.updates++;
  }
  
  recordByzantineFault(nodeId: string, type: string): void {
    const current = this.byzantineFaults.get(nodeId) || 0;
    this.byzantineFaults.set(nodeId, current + 1);
  }
  
  getSnapshot(): MetricsSnapshot {
    return {
      totalUpdates: this.updates,
      byzantineFaultsDetected: Object.fromEntries(this.byzantineFaults),
    };
  }
}

interface MetricsSnapshot {
  totalUpdates: number;
  byzantineFaultsDetected: Record<string, number>;
}
