import { describe, expect, it, vi } from "vitest";

describe("node aggregation proof fallback", () => {
  it("falls back to software proof when wasm aggregation throws", async () => {
    vi.doMock("./wasm/aggregator.js", () => ({
      generateAggregationProof: vi.fn(() => {
        throw new Error("wasm aggregation failed");
      }),
    }));

    const { SovereignNode } = await import("./node.js");

    const node = new SovereignNode({
      nodeId: "node-agg-fallback-only",
      region: "r1",
      coordinates: { lat: 1, lng: 2 },
    }) as any;

    const proof = await node.generateAggregationProof([{ id: 1 }]);
    expect(typeof proof).toBe("string");
    expect(proof.length).toBeGreaterThan(10);
  });
});
