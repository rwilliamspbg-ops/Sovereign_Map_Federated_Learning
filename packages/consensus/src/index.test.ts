import { describe, expect, it } from 'vitest';
import { ConsensusParticipant, HierarchicalAggregator } from './index.js';

class MockNetwork {
  private handlers = new Map<string, Array<(msg: any) => void>>();

  on(event: string, cb: (msg: any) => void): void {
    const existing = this.handlers.get(event) || [];
    existing.push(cb);
    this.handlers.set(event, existing);
  }

  off(event: string, cb: (msg: any) => void): void {
    const existing = this.handlers.get(event) || [];
    this.handlers.set(
      event,
      existing.filter((handler) => handler !== cb)
    );
  }

  async broadcast(message: any): Promise<void> {
    const callbacks = this.handlers.get('message') || [];
    for (const cb of callbacks) {
      cb(message);
    }
  }
}

const logger = {
  info: () => undefined,
  debug: () => undefined,
  warn: () => undefined,
  error: () => undefined
} as any;

describe('ConsensusParticipant', () => {
  it('initializes and can accept a proposal when quorum is met', async () => {
    const network = new MockNetwork();
    const participant = new ConsensusParticipant({ nodeId: 'n1', network, logger });

    await participant.initialize();

    (participant as any).config.quorumSize = 1;

    const result = await participant.submitUpdate('round-1', {
      weights: new Float64Array([1, 2]),
      samples: 10
    });

    expect(result.accepted).toBe(true);
    expect(result.votes).toBeGreaterThanOrEqual(1);
    expect(result.reward).toBe(1);
  });
});

describe('HierarchicalAggregator', () => {
  it('aggregates updates using weighted average behavior', async () => {
    const aggregator = new HierarchicalAggregator({ tier: 'edge', logger });

    await aggregator.start();
    const result = await aggregator.aggregate([
      { nodeId: 'a', weights: new Float64Array([1, 3]), samples: 10 },
      { nodeId: 'b', weights: new Float64Array([3, 5]), samples: 30 }
    ]);
    await aggregator.stop();

    expect(result.samples).toBe(40);
    expect(result.contributors).toEqual(['a', 'b']);
    expect(Array.from(result.weights)).toEqual([2, 4]);
  });
});
