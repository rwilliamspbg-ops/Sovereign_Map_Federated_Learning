import { describe, expect, it, vi } from 'vitest';
import { NetworkClient } from './network.js';

const logger = {
  info: vi.fn(),
  debug: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
  child: vi.fn().mockReturnThis()
} as any;

function makeConnection(peer: string, ok = true) {
  return {
    remotePeer: { toString: () => peer },
    newStream: vi.fn(async () => {
      if (!ok) throw new Error('stream failed');
      return { sink: vi.fn(async () => undefined) };
    })
  };
}

describe('NetworkClient', () => {
  it('reports disconnected status by default', () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, 'n1', logger);

    expect(client.isConnected()).toBe(false);
    const status = client.getStatus();
    expect(status.connected).toBe(false);
    expect(status.peers).toBe(0);
  });

  it('broadcasts to connected peers and returns per-peer results', async () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, 'n1', logger);

    const connections = [makeConnection('peer-a', true), makeConnection('peer-b', false)];
    (client as any).status.connected = true;
    (client as any).node = {
      connectionManager: {
        getConnections: () => connections
      }
    };

    const result = await client.broadcast({ timestamp: 123, hello: 'world' });

    expect(result.id).toBe(123);
    expect(result.accepted).toBe(true);
    expect(result.peersReached).toBe(1);
    expect(result.results).toHaveLength(2);
  });

  it('throws when broadcasting while disconnected', async () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, 'n1', logger);

    await expect(client.broadcast({ timestamp: 1 })).rejects.toThrow('Network not connected');
  });

  it('updates status and disconnects underlying node', async () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, 'n1', logger);
    const stop = vi.fn(async () => undefined);

    (client as any).status.connected = true;
    (client as any).node = {
      stop,
      connectionManager: {
        getConnections: () => [makeConnection('peer-a', true)]
      }
    };

    const status = client.getStatus();
    expect(status.connected).toBe(true);
    expect(status.peers).toBe(1);
    expect(status.latency).toBe(50);

    const state = await client.getState();
    expect(state.peers).toBe(1);

    await client.disconnect();
    expect(stop).toHaveBeenCalledTimes(1);
    expect(client.isConnected()).toBe(false);
  });
});
