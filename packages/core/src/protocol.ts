/**
 * Protocol message contracts for the Sovereign Map SDK.
 */

import type { MapUpdate, NodeEvents } from './types.js';

export enum MessageType {
  MapUpdate = 'map_update',
  StateSync = 'state_sync',
  RoundJoin = 'round_join',
  RoundResult = 'round_result',
  Heartbeat = 'heartbeat'
}

export interface ProtocolMessage {
  id: string;
  type: MessageType;
  nodeId: string;
  timestamp: number;
  payload: MapUpdate | Record<string, unknown>;
}

export type { NodeEvents };
