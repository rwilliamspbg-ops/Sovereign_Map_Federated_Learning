/**
 * Sovereign Map Core SDK
 * 
 * Privacy-preserving decentralized mapping protocol client
 * featuring 55.5% Byzantine fault tolerance and SGP-001 compliance.
 */

export { SovereignNode, NodeConfig, NodeState } from './node.js';
export { NetworkClient, NetworkTopology } from './network.js';
export { ProtocolMessage, MessageType } from './protocol.js';
export { 
  SovereignMapError, 
  PrivacyError, 
  ConsensusError,
  NetworkError 
} from './errors.js';

// Version info
export const SDK_VERSION = '0.1.0-alpha';
export const PROTOCOL_VERSION = '0.2.0';
export const SGP001_VERSION = '1.0.0';
