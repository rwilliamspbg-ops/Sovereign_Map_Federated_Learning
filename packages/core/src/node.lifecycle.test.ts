import { describe, expect, it, vi, beforeEach } from 'vitest';

const mockHardwareAttest = vi.fn(async () => ({ ok: true }));
const mockAggregatorStart = vi.fn(async () => undefined);
const mockAggregatorStop = vi.fn(async () => undefined);
const mockAggregatorOn = vi.fn((_event: string, _handler: (updates: unknown[]) => Promise<void>) => undefined);

vi.mock('@sovereignmap/privacy', () => {
  class PrivacyEngine {
    private budgetCb?: (remaining: number, total: number) => void;
    public hasBudgetFor = vi.fn(() => true);
    public apply = vi.fn(async (u: Record<string, unknown>) => ({ ...u, noiseApplied: true }));
    public on = vi.fn((event: string, cb: (remaining: number, total: number) => void) => {
      if (event === 'budgetUpdate') {
        this.budgetCb = cb;
      }
    });
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

    public triggerBudgetUpdate(remaining: number, total: number) {
      this.budgetCb?.(remaining, total);
    }
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
    public start = mockAggregatorStart;
    public stop = mockAggregatorStop;
    public on = mockAggregatorOn;

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

  it('continues when hardware attestation fails', async () => {
    const { SovereignNode } = await import('./node.js');
    mockHardwareAttest.mockRejectedValueOnce(new Error('tpm not available'));

    const node = new SovereignNode({
      nodeId: 'node-hw-fail',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 },
      hardware: { tpmAvailable: true }
    }) as any;

    await expect(node.initializeHardware()).resolves.toBeUndefined();
    expect(node.hardware).toBeNull();
  });

  it('covers aggregator branches and shutdown handler registration', async () => {
    const { SovereignNode, NodeState } = await import('./node.js');

    const nodeNone = new SovereignNode({
      nodeId: 'node-none',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 },
      consensus: { aggregatorTier: 'none', enabled: true }
    }) as any;
    await expect(nodeNone.becomeAggregator({ tier: 'edge' })).rejects.toThrow('Node not configured as aggregator');

    const nodeAgg = new SovereignNode({
      nodeId: 'node-agg',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 },
      consensus: { aggregatorTier: 'edge', enabled: true }
    }) as any;

    await nodeAgg.becomeAggregator({ tier: 'edge', maxChildren: 4 });
    expect(nodeAgg.state).toBe(NodeState.AGGREGATING);
    expect(nodeAgg.shutdownHandlers.length).toBeGreaterThan(0);

    const onSpy = vi.spyOn(process, 'on');
    nodeAgg.setupShutdownHandlers();
    expect(onSpy).toHaveBeenCalledWith('SIGTERM', expect.any(Function));
    expect(onSpy).toHaveBeenCalledWith('SIGINT', expect.any(Function));
    onSpy.mockRestore();
  });

  it('covers training simulation and aggregation proof fallback', async () => {
    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-train',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 }
    }) as any;

    vi.useFakeTimers();
    const trainPromise = node.trainLocalModel({
      id: 'round-1',
      globalModelHash: 'h',
      round: 2,
      modelParams: { weightCount: 16 },
      trainingConfig: { epochs: 1, batchSize: 1, learningRate: 0.01 },
      deadline: Date.now() + 10000
    });
    await vi.runAllTimersAsync();
    const trained = await trainPromise;
    vi.useRealTimers();

    expect(trained.roundId).toBe('round-1');
    expect(trained.weights).toBeInstanceOf(Float64Array);
    expect(trained.weights.length).toBe(16);

    const aggProof = await node.generateAggregationProof([{ a: 1 }]);
    expect(typeof aggProof).toBe('string');
    expect(aggProof.length).toBeGreaterThan(10);
  });

  it('rethrows submit errors from underlying privacy/network flow', async () => {
    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-submit-error',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 }
    }) as any;

    node.privacy = {
      hasBudgetFor: vi.fn(() => true),
      apply: vi.fn(async () => {
        throw new Error('apply failed');
      })
    };

    await expect(node.submitMapUpdate({ location: { lat: 1, lng: 2 }, quality: 0.4 })).rejects.toThrow('apply failed');
  });

  it('covers consensus error wrap and submitOnline id fallback', async () => {
    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-consensus-fail',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 },
      consensus: { enabled: true }
    }) as any;

    node.consensus = {
      submitUpdate: vi.fn(async () => {
        throw new Error('submit failed');
      })
    };

    await expect(
      node.participateInRound({
        id: 'r1',
        globalModelHash: 'h',
        trainingConfig: { epochs: 1, batchSize: 1, learningRate: 0.1 },
        deadline: Date.now() + 1000
      })
    ).rejects.toThrow('Failed to participate in round r1');

    node.network = {
      broadcast: vi.fn(async () => ({ accepted: true, estimatedTime: 1 }))
    };

    const submitted = await node.submitOnline(
      {
        location: { lat: 1, lng: 2 },
        quality: 1,
        noiseApplied: true,
        privacyProof: 'x',
        epsilonConsumed: 0.01
      },
      'proof'
    );

    expect(submitted.updateId).toBeTruthy();
    expect(submitted.status).toBe('confirmed');
  });

  it('covers aggregator init failure, shutdown failure path, and metrics collector branches', async () => {
    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-deep-branches',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 },
      consensus: { enabled: true, aggregatorTier: 'edge' }
    }) as any;

    mockAggregatorStart.mockRejectedValueOnce(new Error('aggregator start failed'));
    await expect(node.becomeAggregator({ tier: 'edge' })).rejects.toThrow('aggregator start failed');

    node.shutdownHandlers = [
      vi.fn(async () => {
        throw new Error('handler blew up');
      })
    ];
    node.network = {
      disconnect: vi.fn(async () => {
        throw new Error('disconnect failed');
      }),
      isConnected: vi.fn(() => false),
      getStatus: vi.fn(() => ({ connected: false, peers: 0, latency: 0, bandwidth: 0, lastSeen: Date.now() }))
    };

    await expect(node.shutdown()).rejects.toThrow('disconnect failed');

    // Exercise concrete metrics collector branches that are often replaced by mocks.
    const metrics = node.metrics;
    metrics.recordUpdate({ location: { lat: 1, lng: 2 }, quality: 0.8 }, true);
    metrics.recordUpdate({ location: { lat: 1, lng: 2 }, quality: 0.6 }, false);
    metrics.recordRound({ accepted: true, reward: 5, roundId: 'r1' });
    metrics.recordByzantineFault('n1', 'equivocation');
    for (let i = 0; i < 1005; i++) {
      metrics.recordLatency(10 + (i % 3));
    }
    const snap = metrics.getSnapshot();
    expect(snap.byzantineFaultsDetected['n1:equivocation']).toBe(1);
    expect(snap.averageLatency).toBeGreaterThan(0);
    expect(snap.totalUpdates).toBe(2);
    expect(snap.successfulUpdates).toBe(1);
    expect(snap.failedUpdates).toBe(1);
    expect(snap.totalRewards).toBe(5);
  });

  it('covers aggregate callback path and submitAggregate broadcast', async () => {
    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-aggregate-callback',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 },
      consensus: { enabled: true, aggregatorTier: 'edge' }
    }) as any;

    const broadcast = vi.fn(async () => ({ accepted: true }));
    node.network = { broadcast };
    node.generateAggregationProof = vi.fn(async () => 'zk-proof');

    await node.becomeAggregator({ tier: 'edge' });
    const aggregateHandler = mockAggregatorOn.mock.calls.find((c) => c[0] === 'aggregate')?.[1];
    if (!aggregateHandler) {
      throw new Error('Aggregate handler not registered');
    }

    await aggregateHandler([{ id: 1 }, { id: 2 }]);
    expect(node.generateAggregationProof).toHaveBeenCalled();
    expect(broadcast).toHaveBeenCalled();

    await node.submitAggregate([{ id: 3 }], 'proof-x');
    expect(broadcast).toHaveBeenCalledTimes(2);
  });

  it('covers privacy budget update emission callback', async () => {
    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-privacy-callback',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 }
    }) as any;

    const budgetSpy = vi.fn();
    node.on('privacyBudgetUpdate', budgetSpy);
    await node.initializePrivacy();

    node.privacy.triggerBudgetUpdate(0.8, 1.0);
    expect(budgetSpy).toHaveBeenCalledWith({ remaining: 0.8, total: 1.0 });
  });

  it('covers wasm proof fallback and signal handler exit branches', async () => {
    vi.doMock('./wasm/prover.js', () => ({
      generateProof: vi.fn(() => {
        throw new Error('prover failed');
      })
    }));

    const { SovereignNode } = await import('./node.js');
    const node = new SovereignNode({
      nodeId: 'node-signals',
      region: 'r1',
      coordinates: { lat: 1, lng: 2 }
    }) as any;

    const proof = await node.generateProof({
      location: { lat: 1, lng: 2 },
      quality: 1,
      noiseApplied: true,
      privacyProof: 'p',
      epsilonConsumed: 0.01
    });
    expect(typeof proof).toBe('string');

    const onSpy = vi.spyOn(process, 'on');
    const exitSpy = vi.spyOn(process, 'exit').mockImplementation((() => undefined) as never);

    node.shutdown = vi.fn(async () => undefined);
    node.setupShutdownHandlers();
    const sigterm = onSpy.mock.calls.find((c) => c[0] === 'SIGTERM')?.[1] as (() => Promise<void>) | undefined;
    expect(sigterm).toBeTruthy();
    await sigterm?.();
    expect(exitSpy).toHaveBeenCalledWith(0);

    node.shutdown = vi.fn(async () => {
      throw new Error('shutdown boom');
    });
    await sigterm?.();
    expect(exitSpy).toHaveBeenCalledWith(1);

    onSpy.mockRestore();
    exitSpy.mockRestore();
  });
});
