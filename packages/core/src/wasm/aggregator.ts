export async function generateAggregationProof(updates: unknown[]): Promise<string> {
  return JSON.stringify({
    proof: 'mock-aggregation-proof',
    count: updates.length,
  });
}
