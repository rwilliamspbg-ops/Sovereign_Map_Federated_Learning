import { describe, expect, it, vi, afterEach } from "vitest";
import { ConsensusParticipant, HierarchicalAggregator } from "./index.js";

class MockNetwork {
  private handlers = new Map<string, Array<(msg: any) => void>>();
  public messages: any[] = [];

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
    this.messages.push(message);
    const callbacks = this.handlers.get("message") || [];
    for (const cb of callbacks) {
      cb(message);
    }
  }
}

const logger = {
  info: () => undefined,
  debug: () => undefined,
  warn: () => undefined,
  error: () => undefined,
} as any;

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("ConsensusParticipant", () => {
  it("initializes and can accept a proposal when quorum is met", async () => {
    vi.useFakeTimers();
    vi.spyOn(Math, "random").mockReturnValue(0.123456789);
    vi.spyOn(Date, "now").mockReturnValue(1700000000000);

    const network = new MockNetwork();
    const participant = new ConsensusParticipant({
      nodeId: "n1",
      network,
      logger,
    });

    const ready = vi.fn();
    participant.on("ready", ready);
    await participant.initialize();

    (participant as any).config.quorumSize = 1;

    const pending = participant.submitUpdate("round-1", {
      weights: new Float64Array([1, 2]),
      samples: 10,
    });
    await vi.advanceTimersByTimeAsync(150);
    const result = await pending;

    expect(ready).toHaveBeenCalledTimes(1);
    expect(result.accepted).toBe(true);
    expect(result.votes).toBeGreaterThanOrEqual(1);
    expect(result.reward).toBe(1);
    expect((participant as any).state).toBe("committed");
    expect(network.messages[0].type).toBe("CONSENSUS_PROPOSAL");
    expect(network.messages[0].proposal.proof).toContain(
      "zkp-1700000000000-4fzzzxjyl-"
    );
  });

  it("broadcasts leave messages and handles proposal validation branches", async () => {
    const network = new MockNetwork();
    const participant = new ConsensusParticipant({
      nodeId: "n2",
      network,
      logger,
    });

    await participant.initialize();

    await network.broadcast({
      type: "CONSENSUS_PROPOSAL",
      proposal: {
        id: "proposal-valid",
        proposer: "peer-1",
        update: { weights: [1] },
      },
    });

    await network.broadcast({
      type: "CONSENSUS_PROPOSAL",
      proposal: {
        id: "proposal-missing-update",
        proposer: "peer-2",
      },
    });

    await network.broadcast({
      type: "CONSENSUS_VOTE",
      proposalId: "proposal-valid",
      voter: "peer-1",
    });
    await participant.leave();

    expect(network.messages).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          type: "CONSENSUS_VOTE",
          proposalId: "proposal-valid",
          voter: "n2",
        }),
        expect.objectContaining({ type: "CONSENSUS_LEAVE", nodeId: "n2" }),
      ])
    );

    const ownVotes = network.messages.filter(
      (message) => message.type === "CONSENSUS_VOTE" && message.voter === "n2"
    );
    expect(ownVotes).toHaveLength(1);
  });

  it("collects external votes and times out when quorum is not reached", async () => {
    vi.useFakeTimers();

    const network = new MockNetwork();
    const participant = new ConsensusParticipant({
      nodeId: "n3",
      network,
      logger,
    });

    await participant.initialize();
    (participant as any).config.quorumSize = 2;

    const quorumPromise = (participant as any).collectVotes(
      "proposal-quorum",
      1000
    );
    await network.broadcast({
      type: "CONSENSUS_VOTE",
      proposalId: "proposal-quorum",
      voter: "peer-1",
    });
    await vi.runAllTicks();

    await expect(quorumPromise).resolves.toEqual({ reached: true, votes: 2 });

    (participant as any).config.quorumSize = 3;
    const timeoutPromise = (participant as any).collectVotes(
      "proposal-timeout",
      1000
    );
    await network.broadcast({
      type: "CONSENSUS_VOTE",
      proposalId: "proposal-other",
      voter: "peer-2",
    });
    await vi.advanceTimersByTimeAsync(1000);

    await expect(timeoutPromise).resolves.toEqual({ reached: false, votes: 1 });
  });

  it("returns a rejected consensus result when quorum is not met before timeout", async () => {
    vi.useFakeTimers();
    vi.spyOn(Math, "random").mockReturnValue(0.123456789);
    vi.spyOn(Date, "now").mockReturnValue(1700000000100);

    const network = new MockNetwork();
    const participant = new ConsensusParticipant({
      nodeId: "n4",
      network,
      logger,
    });

    await participant.initialize();
    (participant as any).config.quorumSize = 2;

    const pending = participant.submitUpdate("round-timeout", {
      weights: new Float64Array([5, 7]),
      samples: 2,
    });

    await vi.advanceTimersByTimeAsync(150);
    await vi.advanceTimersByTimeAsync(30000);

    const result = await pending;
    expect(result).toEqual({
      accepted: false,
      roundId: "round-timeout",
      reward: 0,
      votes: 1,
    });
    expect((participant as any).state).toBe("idle");
  });
});

describe("HierarchicalAggregator", () => {
  it("aggregates updates using weighted average behavior", async () => {
    const aggregator = new HierarchicalAggregator({ tier: "edge", logger });

    await aggregator.start();
    const result = await aggregator.aggregate([
      { nodeId: "a", weights: new Float64Array([1, 3]), samples: 10 },
      { nodeId: "b", weights: new Float64Array([3, 5]), samples: 30 },
    ]);
    await aggregator.stop();

    expect(result.samples).toBe(40);
    expect(result.contributors).toEqual(["a", "b"]);
    expect(Array.from(result.weights)).toEqual([2, 4]);
  });

  it("handles empty aggregation input and lifecycle events", async () => {
    const aggregator = new HierarchicalAggregator({ tier: "regional", logger });
    const started = vi.fn();
    const stopped = vi.fn();
    const aggregate = vi.fn();

    aggregator.on("started", started);
    aggregator.on("stopped", stopped);
    aggregator.on("aggregate", aggregate);

    await aggregator.start();
    const result = await aggregator.aggregate([]);
    await aggregator.stop();

    expect(started).toHaveBeenCalledTimes(1);
    expect(stopped).toHaveBeenCalledTimes(1);
    expect(aggregate).toHaveBeenCalledWith([]);
    expect(result.samples).toBe(0);
    expect(result.contributors).toEqual([]);
    expect(Array.from(result.weights)).toEqual([]);
    expect((aggregator as any).running).toBe(false);
    expect((aggregator as any).children).toEqual([]);
  });

  it("defaults missing sample counts to zero during aggregation", async () => {
    const aggregator = new HierarchicalAggregator({ tier: "edge", logger });

    const result = await aggregator.aggregate([
      { nodeId: "a", weights: new Float64Array([2, 6]) },
      { nodeId: "b", weights: new Float64Array([4, 8]), samples: 0 },
    ]);

    expect(result.samples).toBe(0);
    expect(Array.from(result.weights)).toEqual([3, 7]);
    expect(result.contributors).toEqual(["a", "b"]);
  });
});
