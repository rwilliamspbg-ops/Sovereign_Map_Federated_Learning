export async function generateProof(payload: unknown): Promise<string> {
  return JSON.stringify({ proof: 'mock-proof', payloadHash: String(payload ? 1 : 0) });
}
