import { beforeEach, describe, expect, it, vi } from "vitest";
import { NetworkClient } from "./network.js";

const mocked = vi.hoisted(() => ({
  createLibp2p: vi.fn(),
  multiaddr: vi.fn((addr: string) => ({ toString: () => addr })),
  tcp: vi.fn(() => ({})),
  noise: vi.fn(() => ({})),
  yamux: vi.fn(() => ({})),
}));

vi.mock("libp2p", () => ({ createLibp2p: mocked.createLibp2p }));
vi.mock("@multiformats/multiaddr", () => ({ multiaddr: mocked.multiaddr }));
vi.mock("@libp2p/tcp", () => ({ tcp: mocked.tcp }));
vi.mock("@chainsafe/libp2p-noise", () => ({ noise: mocked.noise }));
vi.mock("@chainsafe/libp2p-yamux", () => ({ yamux: mocked.yamux }));

const logger = {
  info: vi.fn(),
  debug: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
  child: vi.fn().mockReturnThis(),
} as any;

function makeConnection(peer: string, ok = true) {
  return {
    remotePeer: { toString: () => peer },
    newStream: vi.fn(async () => {
      if (!ok) throw new Error("stream failed");
      return { sink: vi.fn(async () => undefined) };
    }),
  };
}

describe("NetworkClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("reports disconnected status by default", () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, "n1", logger);

    expect(client.isConnected()).toBe(false);
    const status = client.getStatus();
    expect(status.connected).toBe(false);
    expect(status.peers).toBe(0);
  });

  it("broadcasts to connected peers and returns per-peer results", async () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, "n1", logger);

    const connections = [
      makeConnection("peer-a", true),
      makeConnection("peer-b", false),
    ];
    (client as any).status.connected = true;
    (client as any).node = {
      connectionManager: {
        getConnections: () => connections,
      },
    };

    const result = await client.broadcast({ timestamp: 123, hello: "world" });

    expect(result.id).toBe(123);
    expect(result.accepted).toBe(true);
    expect(result.peersReached).toBe(1);
    expect(result.results).toHaveLength(2);
  });

  it("throws when broadcasting while disconnected", async () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, "n1", logger);

    await expect(client.broadcast({ timestamp: 1 })).rejects.toThrow(
      "Network not connected"
    );
  });

  it("updates status and disconnects underlying node", async () => {
    const client = new NetworkClient({ bootstrapPeers: [] }, "n1", logger);
    const stop = vi.fn(async () => undefined);

    (client as any).status.connected = true;
    (client as any).node = {
      stop,
      connectionManager: {
        getConnections: () => [makeConnection("peer-a", true)],
      },
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

  it("initializes libp2p, wires events, and dials bootstrap peers", async () => {
    const handlers: Record<string, (connection: any) => void> = {};
    const dial = vi.fn(async (addr: any) => {
      if (String(addr.toString()).includes("bad-peer")) {
        throw new Error("dial failed");
      }
    });

    const connections = [makeConnection("peer-x", true)];
    mocked.createLibp2p.mockResolvedValue({
      dial,
      peerId: { toString: () => "self-peer" },
      connectionManager: {
        on: vi.fn((event: string, cb: (connection: any) => void) => {
          handlers[event] = cb;
        }),
        getConnections: () => connections,
      },
    });

    const client = new NetworkClient(
      {
        bootstrapPeers: [
          "/ip4/127.0.0.1/tcp/1000",
          "/ip4/127.0.0.1/tcp/bad-peer",
        ],
      },
      "n1",
      logger
    );

    const onConnect = vi.fn();
    const onDisconnect = vi.fn();
    client.on("peer:connect", onConnect);
    client.on("peer:disconnect", onDisconnect);

    await client.connect();

    expect(mocked.createLibp2p).toHaveBeenCalledTimes(1);
    expect(dial).toHaveBeenCalledTimes(2);
    expect(logger.warn).toHaveBeenCalled();

    handlers["peer:connect"]({ remotePeer: { toString: () => "peer-x" } });
    handlers["peer:disconnect"]({ remotePeer: { toString: () => "peer-x" } });
    expect(onConnect).toHaveBeenCalledWith("peer-x");
    expect(onDisconnect).toHaveBeenCalledWith("peer-x");
  });

  it("surfaces connect errors when libp2p initialization fails", async () => {
    mocked.createLibp2p.mockRejectedValue(new Error("libp2p init failed"));
    const client = new NetworkClient({ bootstrapPeers: [] }, "n1", logger);

    await expect(client.connect()).rejects.toThrow("libp2p init failed");
    expect(logger.error).toHaveBeenCalled();
  });
});
