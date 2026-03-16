declare module '@sovereignmap/consensus' {
  export class HierarchicalAggregator {
    constructor(config?: unknown);
    on(event: string, handler: (updates: any[]) => unknown): void;
    start(): Promise<void>;
    stop(): Promise<void>;
    aggregate(updates: unknown[]): Promise<unknown>;
  }

  export class ConsensusParticipant {
    constructor(config?: unknown);
    submitUpdate(roundId: string, update: unknown): Promise<{ accepted: boolean; reward: number; roundId: string; proof?: string }>;
  }
}

declare module '@sovereignmap/hardware' {
  export class HardwareAttestation {
    constructor(config?: unknown);
    verify(): Promise<boolean>;
  }
}

declare module '@sovereignmap/privacy' {
  export class PrivacyEngine {
    constructor(config?: unknown);
    initialize(): Promise<void>;
    apply<T>(update: T): Promise<T & { noiseApplied: boolean; privacyProof: string; epsilonConsumed: number }>;
    hasBudgetFor(update: unknown): boolean;
    getStatus(): unknown;
  }
}

declare module '@sovereignmap/island' {
  export class IslandModeManager {
    constructor(config?: unknown);
    initialize(): Promise<void>;
    queueUpdate(update: unknown, proof: string): Promise<{ updateId: string; status: 'queued'; proof: string }>;
    getQueuedUpdates(): Promise<unknown[]>;
    sync(): Promise<{ updatesSynced: number; updatesQueued: number; conflictsResolved: number; syncDuration: number; bytesTransferred: number }>;
  }
}

declare module './wasm/prover.js' {
  export function generateProof(payload: unknown): Promise<string>;
}

declare module './wasm/aggregator.js' {
  export function generateAggregationProof(updates: unknown[]): Promise<string>;
}
