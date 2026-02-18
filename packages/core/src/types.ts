/**
 * Core type definitions for Sovereign Map SDK
 */

import { NodeState } from './node.js';

// Geographic types
export interface Coordinates {
  lat: number;
  lng: number;
  alt?: number;
}

export interface GeoPolygon {
  coordinates: Coordinates[];
}

// Privacy types
export interface PrivacyBudget {
  epsilon: number;
  delta: number;
  mechanism: 'gaussian' | 'laplace';
}

export interface PrivacyStatus {
  budgetConsumed: number;
  budgetRemaining: number;
  totalBudget: number;
  updatesProcessed: number;
  averageNoiseMagnitude: number;
  mechanism: string;
}

// Map update types
export interface MapUpdate {
  location: Coordinates;
  pointCloud?: Uint8Array;
  imagery?: Uint8Array;
  timestamp?: number;
  quality: number;
  metadata?: Record<string, unknown>;
}

export interface PrivatizedUpdate extends MapUpdate {
  noiseApplied: boolean;
  privacyProof: string;
  epsilonConsumed: number;
  originalHash?: string;
}

export interface SubmissionResult {
  updateId: string;
  status: 'confirmed' | 'pending' | 'queued' | 'failed';
  proof: string;
  estimatedConfirmationTime?: number;
  error?: string;
}

// Network types
export interface NetworkStatus {
  connected: boolean;
  peers: number;
  latency: number;
  bandwidth: number;
  lastSeen: number;
}

export interface SyncStats {
  updatesSynced: number;
  updatesQueued: number;
  conflictsResolved: number;
  syncDuration: number;
  bytesTransferred: number;
}

// Federated Learning types
export interface FLRound {
  id: string;
  globalModelHash: string;
  trainingConfig: {
    epochs: number;
    batchSize: number;
    learningRate: number;
  };
  deadline: number;
}

export interface RoundResult {
  accepted: boolean;
  reward: number;
  roundId: string;
  proof?: string;
}

export interface AggregatorConfig {
  regionBoundary?: GeoPolygon;
  specialization?: 'urban' | 'highway' | 'offroad' | 'indoor';
  maxChildren?: number;
  tier: 'edge' | 'regional' | 'global';
}

// Node status types
export interface NodeStatus {
  id: string;
  state: NodeState;
  uptime: number;
  version: string;
  privacy: PrivacyStatus;
  network: NetworkStatus;
  island: IslandStatus;
  metrics: MetricsSnapshot;
  hardware: HardwareCapabilities;
}

export interface HardwareCapabilities {
  npuTops: number;
  tpmAvailable: boolean;
  maxMemoryMB: number;
  cpuCores: number;
}

export interface IslandStatus {
  mode: 'online' | 'island';
  updatesQueued: number;
  lastSync: number;
  storageUsed: number;
  chainIntegrity: boolean;
  maxOfflineHours: number;
}

export interface MetricsSnapshot {
  totalUpdates: number;
  successfulUpdates: number;
  failedUpdates: number;
  byzantineFaultsDetected: Record<string, number>;
  averageLatency: number;
  totalRewards: number;
}

// Event types
export interface NodeEvents {
  stateChange: { from: NodeState; to: NodeState };
  privacyBudgetUpdate: { remaining: number; total: number };
  connectivityLost: void;
  connectivityRestored: SyncStats;
  mapUpdateSubmitted: { updateId: string; proof: string };
  byzantineFaultDetected: { nodeId: string; faultType: string };
  rewardEarned: { amount: number; reason: string };
  error: Error;
}

// Configuration types
export interface NodeConfig {
  nodeId?: string;
  region: string;
  coordinates: Coordinates;
  privacyBudget?: PrivacyBudget;
  hardware?: Partial<HardwareCapabilities>;
  network?: {
    bootstrapPeers?: string[];
    maxConnections?: number;
    enableRelay?: boolean;
  };
  islandMode?: {
    enabled?: boolean;
    storagePath?: string;
    maxOfflineHours?: number;
  };
  consensus?: {
    enabled?: boolean;
    aggregatorTier?: 'none' | 'edge' | 'regional' | 'global';
    maxAggregationChildren?: number;
  };
  logging?: {
    level?: 'trace' | 'debug' | 'info' | 'warn' | 'error';
    pretty?: boolean;
  };
}
