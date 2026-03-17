import { describe, expect, it, vi, beforeEach } from 'vitest';

const mockHardwareAttest = vi.fn(async () => ({ ok: true }));

vi.mock('@sovereignmap/privacy', () => {
  class PrivacyEngine {
    public hasBudgetFor = vi.fn(() => true);
    public apply = vi.fn(async (u: Record<string, unknown>) => ({ ...u, noiseApplied: true }));
    public on = vi.fn();
    public initialize = vi.fn(async () => undefined);
    public destroy = vi.fn(async () => undefined);
    public getStatus = vi.fn(() => ({
      budgetConsumed: 0,
      budgetRemaining: 1,
      totalBudget: 1,
      updatesProcessed: 0,
      averageNoiseMagnitude: 0,
      mechanism: 'gaussian'
    }));

    constructor(_budget: Record<string, unknown>) {}
  }

  return { PrivacyEngine };
});

vi.mock('@sovereignmap/island', () => {
  class IslandModeManager {
    public initialize = vi.fn(async () => undefined);
    public queueUpdate = vi.fn(async () => ({ updateId: 'q1', status: 'queued', proof: 'p' }));
    public sync = vi.fn(async () => ({
      updatesSynced: 1,
      updatesQueued: 0,
      conflictsResolved: 0,
      syncDuration: 1,
      bytesTransferred: 1
    }));
    public syncWithState = vi.fn(async (_state: Record<string, unknown>) => undefined);
    public getStatus = vi.fn(() => ({
      mode: 'online',
      updatesQueued: 0,
      lastSync: Date.now(),
      storageUsed: 0,
      chainIntegrity: true,
      maxOfflineHours: 24
    }));
    public flush = vi.fn(async () => undefined);

    constructor(_cfg: Record<string, unknown>) {}
  }

  return { IslandModeManager };
});

vi.mock('@sovereignmap/consensus', () => {
  class ConsensusParticipant {
    public initialize = vi.fn(async () => undefined);
    public submitUpdate = vi.fn(async (_round: string, _update: Record<string, unknown>) => ({
      accepted: true,
      reward: 1,
      roundId: 'r1'
    }));
    public leave = vi.fn(async () => undefined);

    constructor(_cfg: Record<string, unknown>) {}
  }

  class HierarchicalAggregator {
    public start = vi.fn(async () => undefined);
    public stop = vi.fn(async () => undefined);
    public on = vi.fn((_event: string, _handler: (updates: unknown[]) => Promise<void>) => undefined);

    constructor(_cfg: Record<string, unknown>) {}
  }

  return { ConsensusParticipant, HierarchicalAggregator };
});

vi.mock('@sovereignmap/hardware', () => {
  class HardwareAttestation {
    public attest = mockHardwareAttest;

    constructor(_cfg: Record<string, unknown>) {}
  }

  return { HardwareAttestation };
});

describe('SovereignNode lifecycle branches', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  async function createNodeWithNetwork() {
    const { SovereignNode, NodeState } = await import('./node.js');

    const node = new SovereignNode({
      nodeId: 'node-lifecycle',
      region: 'region-lifecycle',
      coordinates: { lat: 10, lng: 20 },
      hardware: { tpmAvailable: true },
      consensus: { enabled: true }
    }) as any;

    const handlers: Record<string, (...args: unknown[]) => unknown> = {};
    node.network = {
      connect: vi.fn(async () => undefined),
      disconnect: vi.fn(async () => undefined),
      isConnected: vi.fn(() => true),
      getStatus: vi.fn(() => ({ connected: true, peers: 1, latency: 10, bandwidth: 0, lastSeen: Date.now() })),
      getState: vi.fn(async () => ({ peers: 1 })),
      broadcast: vi.fn(async () => ({ id: 'u1', accepted: false, estimatedTime: 3 })),
      on: vi.fn((event: string, cb: (...args: unknown[]) => unknown) => {
        handlers[event] = cb;
      })
    };

    node.setupShutdownHandlers = vi.fn();

    return { node, handlers, NodeState };
  }

  it('initializes end-to-end and emits ready state', async () => {
    const { node, NodeState } = await createNodeWithNetwork();
    const onReady = vi.fn();
    node.on('ready', onReady);

    await node.initialize();

    expect(mockHardwareAttest).toHaveBeenCalled();
    expect(node.state).toBe(NodeState.ONLINE);
    expect(onReady).toHaveBeenCalledTimes(1);
    expect(node.setupShutdownHandlers).toHaveBeenCalledTimes(1);
  });

  it('wraps initialization failures as NodeInitializationError', async () => {
    const { node, NodeState } = await createNodeWithNetwork();
    node.initializePrivacy = vi.fn(async () => {
      throw new Error('privacy init failed');
    });

    await expect(node.initialize()).rejects.toThrow('Failed to initialize node');
    expect(node.state).toBe(NodeState.ERROR);
  });

  it('handles network callbacks for disconnect/reconnect/byzantine branches', async () => {
    const { node, handlers, NodeState } = await createNodeWithNetwork();
    node.metrics = {
      recordUpdate: vi.fn(),
      recordRound: vi.fn(),
      getSnapshot: vi.fn(() => ({ totalUpdates: 0, successfulUpdates: 0, failedUpdates: 0, byzantineFaultsDetected: {}, averageLatency: 0, totalRewards: 0 })),
      recordByzantineFault: vi.fn()
    };

    await node.initializeIsland();
    await node.connectNetwork();

    const lostSpy = vi.fn();
    const restoredSpy = vi.fn();
    node.on('connectivityLost', lostSpy);
    node.on('connectivityRestored', restoredSpy);

    handlers.disconnected();
    expect(lostSpy).toHaveBeenCalledTimes(1);
    expect(node.state).toBe(NodeState.ISLAND_MODE);

    await handlers.reconnected();
    expect(restoredSpy).toHaveBeenCalledTimes(1);
    expect(node.state).toBe(NodeState.ONLINE);

    handlers.byzantineDetected('bad-node', 'equivocation');
    expect(node.metrics.recordByzantineFault).toHaveBeenCalledWith('bad-node', 'equivocation');
  });

  it('covers proof fallback, submitOnline pending branch, and sync fallback', async () => {
    const { node } = await createNodeWithNetwork();

    const proof = await node.generateProof({
      location: { lat: 1, lng: 2 },
      quality: 1,
      noiseApplied: true,
      privacyProof: 'x',
      epsilonConsumed: 0.01
    });

    expect(typeof proof).toBe('string');
    expect(proof.length).toBeGreaterThan(10);

    const pending = await node.submitOnline(
      {
        location: { lat: 1, lng: 2 },
        quality: 1,
        noiseApplied: true,
        privacyProof: 'x',
        epsilonConsumed: 0.01
      },
      'proof-pending'
    );

    expect(pending.status).toBe('pending');

    node.island = null;
    const syncStats = await node.syncIsland();
    expect(syncStats.updatesSynced).toBe(0);
  });

  it('throws PrivacyBudgetExceededError when budget is exhausted', async () => {
    const { node } = await createNodeWithNetwork();
    node.privacy = {
      hasBudgetFor: vi.fn(() => false),
      getStatus: vi.fn(() => ({ budgetRemaining: 0 }))
    };

    await expect(
      node.submitMapUpdate({ location: { lat: 1, lng: 2 }, quality: 0.7 })
    ).rejects.toThrow('Privacy budget exhausted');
  });
});
