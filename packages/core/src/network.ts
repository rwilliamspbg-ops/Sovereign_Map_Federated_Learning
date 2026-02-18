/**
 * Network client for P2P communication with libp2p
 */

import { EventEmitter } from 'eventemitter3';
import type { Logger } from 'pino';
import { createLibp2p, Libp2pOptions } from 'libp2p';
import { tcp } from '@libp2p/tcp';
import { mplex } from '@libp2p/mplex';
import { noise } from '@libp2p/noise';
import { multiaddr } from '@multiformats/multiaddr';
import type { PeerId } from '@libp2p/interface-peer-id';

export interface NetworkTopology {
  type: 'mesh' | 'star' | 'hybrid';
  peers: string[];
}

export interface NetworkStatus {
  connected: boolean;
  peers: number;
  latency: number;
  bandwidth: number;
  lastSeen: number;
}

export class NetworkClient extends EventEmitter {
  private node: any; // Libp2p node
  private config: any;
  private nodeId: string;
  private logger: Logger;
  private status: NetworkStatus;
  private bootstrapPeers: string[];

  constructor(config: any, nodeId: string, logger: Logger) {
    super();
    this.config = config;
    this.nodeId = nodeId;
    this.logger = logger;
    this.bootstrapPeers = config.bootstrapPeers || [];
    this.status = {
      connected: false,
      peers: 0,
      latency: 0,
      bandwidth: 0,
      lastSeen: Date.now()
    };
  }

  async connect(): Promise<void> {
    this.logger.info('Initializing libp2p network');

    try {
      const options: Libp2pOptions = {
        transports: [tcp()],
        connectionEncryption: [noise()],
        streamMuxers: [mplex()],
        peerDiscovery: [],
        relay: this.config.enableRelay ? {
          enabled: true,
          hop: {
            enabled: false
          }
        } : undefined
      };

      this.node = await createLibp2p(options);

      // Set up event handlers
      this.node.connectionManager.on('peer:connect', (connection: any) => {
        this.logger.debug({ peer: connection.remotePeer.toString() }, 'Peer connected');
        this.updateStatus();
        this.emit('peer:connect', connection.remotePeer.toString());
      });

      this.node.connectionManager.on('peer:disconnect', (connection: any) => {
        this.logger.debug({ peer: connection.remotePeer.toString() }, 'Peer disconnected');
        this.updateStatus();
        this.emit('peer:disconnect', connection.remotePeer.toString());
      });

      // Connect to bootstrap peers
      for (const addr of this.bootstrapPeers) {
        try {
          const ma = multiaddr(addr);
          await this.node.dial(ma);
          this.logger.info({ addr }, 'Connected to bootstrap peer');
        } catch (error) {
          this.logger.warn({ error, addr }, 'Failed to connect to bootstrap peer');
        }
      }

      this.status.connected = true;
      this.logger.info({ peerId: this.node.peerId.toString() }, 'Network initialized');

    } catch (error) {
      this.logger.error({ error }, 'Failed to initialize network');
      throw error;
    }
  }

  async disconnect(): Promise<void> {
    if (this.node) {
      await this.node.stop();
      this.status.connected = false;
      this.logger.info('Network disconnected');
    }
  }

  isConnected(): boolean {
    return this.status.connected && this.node?.connectionManager.getConnections().length > 0;
  }

  getStatus(): NetworkStatus {
    this.updateStatus();
    return { ...this.status };
  }

  async broadcast(message: any): Promise<any> {
    if (!this.isConnected()) {
      throw new Error('Network not connected');
    }

    const peers = this.node.connectionManager.getConnections();
    const results = [];

    for (const connection of peers) {
      try {
        const stream = await connection.newStream('/sovereignmap/1.0.0');
        const data = JSON.stringify(message);
        await stream.sink([Buffer.from(data)]);
        results.push({ peer: connection.remotePeer.toString(), success: true });
      } catch (error) {
        this.logger.warn({ error, peer: connection.remotePeer.toString() }, 'Failed to send to peer');
        results.push({ peer: connection.remotePeer.toString(), success: false, error });
      }
    }

    return {
      id: message.timestamp,
      accepted: results.some(r => r.success),
      peersReached: results.filter(r => r.success).length,
      results
    };
  }

  async getState(): Promise<any> {
    // Fetch network state from peers
    return {
      timestamp: Date.now(),
      peers: this.node?.connectionManager.getConnections().length || 0
    };
  }

  private updateStatus(): void {
    if (!this.node) return;

    const connections = this.node.connectionManager.getConnections();
    this.status = {
      connected: connections.length > 0,
      peers: connections.length,
      latency: this.estimateLatency(),
      bandwidth: 0, // Would need actual measurement
      lastSeen: Date.now()
    };
  }

  private estimateLatency(): number {
    // Simple estimation based on connection count
    const connections = this.node?.connectionManager.getConnections().length || 0;
    return connections > 0 ? 50 : 0; // 50ms baseline
  }
}
