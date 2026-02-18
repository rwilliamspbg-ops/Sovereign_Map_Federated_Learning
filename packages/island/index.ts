/**
 * Island Mode Manager
 * 
 * Enables autonomous operation when network connectivity is lost.
 * Features tamper-evident state logging and automatic reconciliation.
 */

import { Level } from 'level';
import { ulid } from 'ulid';
import { createHash, createSign } from 'crypto';

export interface IslandModeConfig {
  enabled: boolean;
  storagePath: string;
  maxOfflineHours: number;
}

export interface QueuedUpdate {
  id: string;
  timestamp: number;
  update: object;
  proof: string;
  sequenceNumber: number;
}

export interface IslandStatus {
  mode: 'online' | 'island';
  updatesQueued: number;
  lastSync: number;
  storageUsed: number;
  chainIntegrity: boolean;
}

/**
 * Tamper-evident state manager for offline operation
 */
export class IslandModeManager {
  
  private db: Level<string, string>;
  private config: IslandModeConfig;
  private updateChain: string[] = []; // Hash chain for tamper detection
  private sequenceNumber: number = 0;
  private lastHash: string = 'genesis';
  
  constructor(config: IslandModeConfig) {
    this.config = config;
    this.db = new Level(config.storagePath, { valueEncoding: 'json' });
  }
  
  async initialize(): Promise<void> {
    await this.db.open();
    
    // Load chain state if exists
    try {
      const saved = await this.db.get('__chain_state__');
      const state = JSON.parse(saved);
      this.sequenceNumber = state.sequenceNumber;
      this.lastHash = state.lastHash;
      this.updateChain = state.chain;
    } catch {
      // Fresh start
      await this.initializeChain();
    }
  }
  
  /**
   * Queue an update while in Island Mode
   */
  async queueUpdate(update: object, proof: string): Promise<{ id: string; position: number }> {
    if (!this.config.enabled) {
      throw new Error('Island Mode not enabled');
    }
    
    this.sequenceNumber++;
    
    // Create tamper-evident entry
    const entry: QueuedUpdate = {
      id: ulid(),
      timestamp: Date.now(),
      update,
      proof,
      sequenceNumber: this.sequenceNumber,
    };
    
    // Compute hash chain
    const entryHash = this.hashEntry(entry);
    const chainedHash = this.computeChainedHash(entryHash, this.lastHash);
    
    // Store with hash chain
    await this.db.put(`update:${entry.id}`, JSON.stringify({
      ...entry,
      chainedHash,
      previousHash: this.lastHash,
    }));
    
    // Update chain state
    this.updateChain.push(chainedHash);
    this.lastHash = chainedHash;
    await this.saveChainState();
    
    return { id: entry.id, position: this.sequenceNumber };
  }
  
  /**
   * Sync queued updates when connectivity restored
   */
  async sync(): Promise<SyncResult> {
    const updates: QueuedUpdate[] = [];
    const stream = this.db.iterator({ gt: 'update:', lt: 'update:~' });
    
    for await (const [key, value] of stream) {
      const parsed = JSON.parse(value);
      updates.push(parsed);
    }
    
    // Verify chain integrity before sync
    const integrity = this.verifyChainIntegrity(updates);
    if (!integrity.valid) {
      throw new Error(`Chain integrity violated at position ${integrity.violationAt}`);
    }
    
    // Sort by sequence number
    updates.sort((a, b) => a.sequenceNumber - b.sequenceNumber);
    
    return {
      updatesQueued: updates.length,
      updates: updates,
      chainIntegrity: true,
      syncDuration: 0, // To be filled by caller
    };
  }
  
  /**
   * Flush storage after successful sync
   */
  async flush(): Promise<void> {
    const batch = this.db.batch();
    
    for await (const [key] of this.db.iterator({ gt: 'update:', lt: 'update:~' })) {
      batch.del(key);
    }
    
    // Reset chain
    this.sequenceNumber = 0;
    this.updateChain = [];
    this.lastHash = 'genesis';
    batch.put('__chain_state__', JSON.stringify({
      sequenceNumber: 0,
      lastHash: 'genesis',
      chain: [],
    }));
    
    await batch.write();
  }
  
  /**
   * Verify local chain hasn't been tampered with
   */
  async verifyIntegrity(): Promise<boolean> {
    const updates = await this.getAllUpdates();
    return this.verifyChainIntegrity(updates).valid;
  }
  
  getStatus(): IslandStatus {
    return {
      mode: 'online', // Updated by caller based on connectivity
      updatesQueued: this.sequenceNumber,
      lastSync: Date.now(), // Last successful sync timestamp
      storageUsed: 0, // TODO: Implement
      chainIntegrity: true, // TODO: Verify on demand
    };
  }
  
  private async initializeChain(): Promise<void> {
    await this.db.put('__chain_state__', JSON.stringify({
      sequenceNumber: 0,
      lastHash: 'genesis',
      chain: [],
    }));
  }
  
  private hashEntry(entry: QueuedUpdate): string {
    return createHash('sha256')
      .update(JSON.stringify(entry))
      .digest('hex');
  }
  
  private computeChainedHash(entryHash: string, previousHash: string): string {
    return createHash('sha256')
      .update(entryHash + previousHash)
      .digest('hex');
  }
  
  private async saveChainState(): Promise<void> {
    await this.db.put('__chain_state__', JSON.stringify({
      sequenceNumber: this.sequenceNumber,
      lastHash: this.lastHash,
      chain: this.updateChain,
    }));
  }
  
  private async getAllUpdates(): Promise<any[]> {
    const updates = [];
    for await (const [, value] of this.db.iterator({ gt: 'update:', lt: 'update:~' })) {
      updates.push(JSON.parse(value));
    }
    return updates;
  }
  
  private verifyChainIntegrity(updates: any[]): { valid: boolean; violationAt?: number } {
    let expectedPrevious = 'genesis';
    
    for (let i = 0; i < updates.length; i++) {
      const update = updates[i];
      
      if (update.previousHash !== expectedPrevious) {
        return { valid: false, violationAt: i };
      }
      
      // Verify hash computation
      const recomputed = this.computeChainedHash(
        this.hashEntry(update),
        update.previousHash
      );
      
      if (recomputed !== update.chainedHash) {
        return { valid: false, violationAt: i };
      }
      
      expectedPrevious = update.chainedHash;
    }
    
    return { valid: true };
  }
}

interface SyncResult {
  updatesQueued: number;
  updates: QueuedUpdate[];
  chainIntegrity: boolean;
  syncDuration: number;
}
