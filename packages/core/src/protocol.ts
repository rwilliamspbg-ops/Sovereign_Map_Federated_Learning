/**
 * Protocol message contracts for the Sovereign Map SDK.
 */

import type { MapUpdate, NodeEvents } from "./types.js";

export enum MessageType {
  MapUpdate = "map_update",
  StateSync = "state_sync",
  RoundJoin = "round_join",
  RoundResult = "round_result",
  Heartbeat = "heartbeat",
}

export interface GradientMessage {
  nodeId: string;
  taskId: string;
  round: number;
  gradients: number[];
  timestampMs?: number;
}

export interface GradientAck {
  accepted: boolean;
  reason?: string;
  negotiatedKEX?: string;
  kexPublicKeyLen?: number;
  batchAccepted?: number;
  batchRejected?: number;
}

export interface ProtocolMessage {
  id: string;
  type: MessageType;
  nodeId: string;
  timestamp: number;
  payload: MapUpdate | GradientMessage | Record<string, unknown>;
}

export type { NodeEvents };
