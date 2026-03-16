/**
 * Sovereign Map Core SDK
 * 
 * Privacy-preserving decentralized mapping protocol client
 * featuring 55.5% Byzantine fault tolerance and SGP-001 compliance.
 * 
 * @example
 * ```typescript
 * import { SovereignNode } from '@sovereignmap/core';
 * 
 * const node = new SovereignNode({
 *   region: 'bavaria-001',
 *   coordinates: { lat: 48.1351, lng: 11.5820 },
 *   privacyBudget: { epsilon: 1.0, delta: 1e-5 },
 *   hardware: { npuTops: 85, tpmAvailable: true }
 * });
 * 
 * await node.initialize();
 * 
 * const result = await node.submitMapUpdate({
 *   location: { lat: 48.1351, lng: 11.5820 },
 *   pointCloud: data,
 *   quality: 0.95
 * });
 * ```
 */

// Core exports
export { SovereignNode, NodeState } from './node.js';
export type { NodeConfig, NodeStatus } from './node.js';
export { NetworkClient } from './network.js';
export type { NetworkTopology, NetworkStatus } from './network.js';
export { MessageType } from './protocol.js';
export type { ProtocolMessage } from './protocol.js';

// Type exports
export type {
  MapUpdate,
  PrivatizedUpdate,
  SubmissionResult,
  SyncStats,
  FLRound,
  RoundResult,
  AggregatorConfig,
  MetricsSnapshot,
  PrivacyStatus,
  IslandStatus
} from './types.js';

// Error exports
export {
  SovereignMapError,
  NodeInitializationError,
  PrivacyBudgetExceededError,
  ConsensusError,
  NetworkError,
  HardwareAttestationError,
  IslandModeError
} from './errors.js';

// Constants
export const SDK_VERSION = '0.1.0-alpha.1';
export const PROTOCOL_VERSION = '1.0.0';
export const SGP001_VERSION = '1.0.0';
export const MIN_NODE_VERSION = '18.0.0';

// Logger
export { createLogger } from './logger.js';
export type { Logger } from './logger.js';
