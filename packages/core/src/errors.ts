/**
 * Comprehensive error handling for Sovereign Map SDK
 */

export class SovereignMapError extends Error {
  public readonly code: string;
  public readonly isOperational: boolean;
  public readonly timestamp: Date;

  constructor(
    message: string,
    code: string = 'SOVEREIGNMAP_ERROR',
    isOperational: boolean = true,
    cause?: Error
  ) {
    super(message, { cause });
    this.name = 'SovereignMapError';
    this.code = code;
    this.isOperational = isOperational;
    this.timestamp = new Date();
    Error.captureStackTrace(this, this.constructor);
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      timestamp: this.timestamp.toISOString(),
      stack: this.stack,
      cause: this.cause?.message
    };
  }
}

export class NodeInitializationError extends SovereignMapError {
  constructor(message: string, cause?: Error) {
    super(message, 'NODE_INIT_ERROR', true, cause);
    this.name = 'NodeInitializationError';
  }
}

export class PrivacyBudgetExceededError extends SovereignMapError {
  public readonly currentBudget: number;
  public readonly requiredBudget: number;

  constructor(current: number, required: number) {
    super(
      `Privacy budget exhausted (current: ${current}, required: ${required})`,
      'PRIVACY_BUDGET_EXCEEDED',
      true
    );
    this.name = 'PrivacyBudgetExceededError';
    this.currentBudget = current;
    this.requiredBudget = required;
  }
}

export class ConsensusError extends SovereignMapError {
  public readonly roundId?: string;

  constructor(message: string, roundId?: string, cause?: Error) {
    super(message, 'CONSENSUS_ERROR', true, cause);
    this.name = 'ConsensusError';
    this.roundId = roundId;
  }
}

export class NetworkError extends SovereignMapError {
  public readonly endpoint?: string;
  public readonly statusCode?: number;

  constructor(message: string, endpoint?: string, statusCode?: number, cause?: Error) {
    super(message, 'NETWORK_ERROR', true, cause);
    this.name = 'NetworkError';
    this.endpoint = endpoint;
    this.statusCode = statusCode;
  }
}

export class HardwareAttestationError extends SovereignMapError {
  public readonly devicePath?: string;

  constructor(message: string, devicePath?: string, cause?: Error) {
    super(message, 'HW_ATTESTATION_ERROR', true, cause);
    this.name = 'HardwareAttestationError';
    this.devicePath = devicePath;
  }
}

export class IslandModeError extends SovereignMapError {
  constructor(message: string, cause?: Error) {
    super(message, 'ISLAND_MODE_ERROR', true, cause);
    this.name = 'IslandModeError';
  }
}

export class ZKProofError extends SovereignMapError {
  constructor(message: string, cause?: Error) {
    super(message, 'ZK_PROOF_ERROR', true, cause);
    this.name = 'ZKProofError';
  }
}
