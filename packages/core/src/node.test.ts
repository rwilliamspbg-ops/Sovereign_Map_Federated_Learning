import { describe, expect, it, vi } from 'vitest';
import { NodeConfigSchema, NodeState, SovereignNode } from './node.js';

describe('NodeConfigSchema', () => {
  it('applies defaults and validates config', () => {
    const cfg = NodeConfigSchema.parse({
      region: 'r1',
      coordinates: { lat: 10, lng: 20 }
    });

    expect(cfg.nodeId).toBeTruthy();
    expect(cfg.privacyBudget.epsilon).toBe(1);
    expect(cfg.consensus.enabled).toBe(true);
    expect(cfg.islandMode.enabled).toBe(true);
  });
});

describe('SovereignNode', () => {
  function createNode() {
    return new SovereignNode({
      nodeId: 'node-test',
      region: 'region-1',
      coordinates: { lat: 10, lng: 20 },
      hardware: { tpmAvailable: false }
    });
  }

  it('submits update online and records successful metric', async () => {
    const node = createNode() as any;

    node.privacy = {
      hasBudgetFor: () => true,
      apply: vi.fn(async (u: any) => ({ ...u, noiseApplied: true }))
    };
    node.generateProof = vi.fn(async () => 'proof-123');
    node.network = {
      isConnected: () => true,
      broadcast: vi.fn(async () => ({ id: 'u1', accepted: true, estimatedTime: 1 }))
    };
    node.island = {
      queueUpdate: vi.fn(async () => ({ updateId: 'q1', status: 'queued', proof: 'proof-123' }))
    };
    node.metrics = { recordUpdate: vi.fn(), recordRound: vi.fn(), getSnapshot: () => ({ totalUpdates: 0, successfulUpdates: 0, failedUpdates: 0, byzantineFaultsDetected: {}, averageLatency: 0, totalRewards: 0 }), recordByzantineFault: vi.fn() };

    const result = await node.submitMapUpdate({
      location: { lat: 1, lng: 2 },
      quality: 0.9
    });

    expect(result.status).toBe('confirmed');
    expect(node.metrics.recordUpdate).toHaveBeenCalledWith(expect.anything(), true);
  });

  it('queues update offline and transitions to island mode', async () => {
    const node = createNode() as any;

    node.privacy = {
      hasBudgetFor: () => true,
      apply: vi.fn(async (u: any) => ({ ...u, noiseApplied: true }))
    };
    node.generateProof = vi.fn(async () => 'proof-456');
    node.network = {
      isConnected: () => false
    };
    node.island = {
      queueUpdate: vi.fn(async () => ({ updateId: 'q1', status: 'queued', proof: 'proof-456' }))
    };
    node.metrics = { recordUpdate: vi.fn(), recordRound: vi.fn(), getSnapshot: () => ({ totalUpdates: 0, successfulUpdates: 0, failedUpdates: 0, byzantineFaultsDetected: {}, averageLatency: 0, totalRewards: 0 }), recordByzantineFault: vi.fn() };

    const result = await node.submitMapUpdate({
      location: { lat: 1, lng: 2 },
      quality: 0.8
    });

    expect(result.status).toBe('queued');
    expect(node.state).toBe(NodeState.ISLAND_MODE);
    expect(node.metrics.recordUpdate).toHaveBeenCalledWith(expect.anything(), false);
  });

  it('throws if consensus is disabled and participates when available', async () => {
    const node = createNode() as any;

    await expect(
      node.participateInRound({
        id: 'r1',
        globalModelHash: 'h1',
        trainingConfig: { epochs: 1, batchSize: 1, learningRate: 0.1 },
        deadline: Date.now() + 1000
      })
    ).rejects.toThrow('Consensus not enabled for this node');

    node.consensus = {
      submitUpdate: vi.fn(async () => ({ accepted: true, reward: 2, roundId: 'r1' }))
    };
    node.trainLocalModel = vi.fn(async () => ({ weights: new Float64Array([1]), samples: 5 }));
    node.metrics = { recordUpdate: vi.fn(), recordRound: vi.fn(), getSnapshot: () => ({ totalUpdates: 0, successfulUpdates: 0, failedUpdates: 0, byzantineFaultsDetected: {}, averageLatency: 0, totalRewards: 0 }), recordByzantineFault: vi.fn() };

    const res = await node.participateInRound({
      id: 'r1',
      globalModelHash: 'h1',
      trainingConfig: { epochs: 1, batchSize: 1, learningRate: 0.1 },
      deadline: Date.now() + 1000
    });

    expect(res.accepted).toBe(true);
    expect(res.reward).toBe(2);
    expect(node.metrics.recordRound).toHaveBeenCalled();
  });

  it('returns status snapshot and performs graceful shutdown', async () => {
    const node = createNode() as any;
    const handler = vi.fn(async () => undefined);
    const flush = vi.fn(async () => undefined);
    const leave = vi.fn(async () => undefined);
    const disconnect = vi.fn(async () => undefined);
    const destroy = vi.fn(async () => undefined);

    node.shutdownHandlers = [handler];
    node.island = {
      getStatus: () => ({ mode: 'online', updatesQueued: 0, lastSync: Date.now(), storageUsed: 1, chainIntegrity: true }),
      flush
    };
    node.consensus = { leave };
    node.privacy = {
      getStatus: () => ({ budgetConsumed: 0, budgetRemaining: 1, totalBudget: 1, updatesProcessed: 0, averageNoiseMagnitude: 0, mechanism: 'gaussian' }),
      destroy
    };
    node.network = {
      isConnected: () => true,
      getStatus: () => ({ connected: true, peers: 1, latency: 10, bandwidth: 0, lastSeen: Date.now() }),
      disconnect
    };
    node.metrics = { recordUpdate: vi.fn(), recordRound: vi.fn(), getSnapshot: () => ({ totalUpdates: 1, successfulUpdates: 1, failedUpdates: 0, byzantineFaultsDetected: {}, averageLatency: 1, totalRewards: 2 }), recordByzantineFault: vi.fn() };

    const status = node.getStatus();
    expect(status.id).toBe('node-test');
    expect(status.network.connected).toBe(true);

    await node.shutdown();

    expect(handler).toHaveBeenCalledTimes(1);
    expect(flush).toHaveBeenCalledTimes(1);
    expect(leave).toHaveBeenCalledTimes(1);
    expect(disconnect).toHaveBeenCalledTimes(1);
    expect(destroy).toHaveBeenCalledTimes(1);
    expect(node.state).toBe(NodeState.OFFLINE);
  });
});
