/**
 * Island Mode Manager
 *
 * Enables autonomous operation when network connectivity is lost.
 * Features tamper-evident state logging and automatic reconciliation.
 */

import { Level } from "level";
import { ulid } from "ulid";
import { createHash } from "crypto";

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
  mode: "online" | "island";
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
  private lastHash: string = "genesis";
  private approximateStorageBytes: number = 0;

  constructor(config: IslandModeConfig) {
    this.config = config;
    this.db = new Level(config.storagePath, { valueEncoding: "json" });
  }

  async initialize(): Promise<void> {
    await this.db.open();

    // Load chain state if exists
    try {
      const saved = await this.db.get("__chain_state__");
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
  async queueUpdate(
    update: object,
    proof: string
  ): Promise<{ id: string; position: number }> {
    if (!this.config.enabled) {
      throw new Error("Island Mode not enabled");
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

    // Store with sequence-padded key for natural LevelDB index order
    const key = `update:${this.sequenceNumber.toString().padStart(12, "0")}:${entry.id}`;
    await this.db.put(
      key,
      JSON.stringify({
        ...entry,
        chainedHash,
        previousHash: this.lastHash,
      })
    );

    // Update chain state
    this.updateChain.push(chainedHash);
    this.lastHash = chainedHash;
    this.approximateStorageBytes += JSON.stringify(entry).length;
    await this.saveChainState();

    return { id: entry.id, position: this.sequenceNumber };
  }

  /**
   * Stream queued updates as an async generator while verifying chain integrity on-the-fly.
   * Eliminates full DB accumulation into memory arrays.
   */
  async *streamUpdates(): AsyncGenerator<QueuedUpdate, void, unknown> {
    const stream = this.db.iterator({ gt: "update:", lt: "update:~" });
    let expectedPrevious = "genesis";
    let position = 0;

    for await (const [, value] of stream) {
      const update = JSON.parse(value);

      if (update.previousHash !== expectedPrevious) {
        throw new Error(
          `Chain integrity violated at position ${position}`
        );
      }

      const canonical: QueuedUpdate = {
        id: update.id,
        timestamp: update.timestamp,
        update: update.update,
        proof: update.proof,
        sequenceNumber: update.sequenceNumber,
      };

      const recomputed = this.computeChainedHash(
        this.hashEntry(canonical),
        update.previousHash
      );

      if (recomputed !== update.chainedHash) {
        throw new Error(
          `Chain integrity violated at position ${position}`
        );
      }

      expectedPrevious = update.chainedHash;
      position++;

      yield canonical;
    }
  }

  /**
   * Sync queued updates when connectivity restored.
   * If an onUpdate callback is provided, updates are processed sequentially via stream
   * without memory accumulation. Otherwise, updates are returned in an array.
   */
  async sync(
    onUpdate?: (update: QueuedUpdate) => Promise<void> | void
  ): Promise<SyncResult> {
    const updates: QueuedUpdate[] = [];
    let count = 0;

    for await (const update of this.streamUpdates()) {
      count++;
      if (onUpdate) {
        await onUpdate(update);
      } else {
        updates.push(update);
      }
    }

    return {
      updatesQueued: count,
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

    for await (const [key] of this.db.iterator({
      gt: "update:",
      lt: "update:~",
    })) {
      batch.del(key);
    }

    // Reset chain
    this.sequenceNumber = 0;
    this.updateChain = [];
    this.lastHash = "genesis";
    this.approximateStorageBytes = 0;
    batch.put(
      "__chain_state__",
      JSON.stringify({
        sequenceNumber: 0,
        lastHash: "genesis",
        chain: [],
      })
    );

    await batch.write();
  }

  /**
   * Verify local chain hasn't been tampered with
   */
  async verifyIntegrity(): Promise<boolean> {
    try {
      for await (const _ of this.streamUpdates()) {
        // Stream through to verify integrity on-the-fly
      }
      return true;
    } catch {
      return false;
    }
  }

  getStatus(): IslandStatus {
    return {
      mode: "online", // Updated by caller based on connectivity
      updatesQueued: this.sequenceNumber,
      lastSync: Date.now(), // Last successful sync timestamp
      storageUsed: this.calculateStorageUsed(),
      chainIntegrity: this.updateChain.length === this.sequenceNumber,
    };
  }

  private calculateStorageUsed(): number {
    // Approximate bytes tracked as updates are queued.
    return this.approximateStorageBytes;
  }

  private async initializeChain(): Promise<void> {
    await this.db.put(
      "__chain_state__",
      JSON.stringify({
        sequenceNumber: 0,
        lastHash: "genesis",
        chain: [],
      })
    );
  }

  private hashEntry(entry: QueuedUpdate): string {
    return createHash("sha256").update(JSON.stringify(entry)).digest("hex");
  }

  private computeChainedHash(entryHash: string, previousHash: string): string {
    return createHash("sha256")
      .update(entryHash + previousHash)
      .digest("hex");
  }

  private async saveChainState(): Promise<void> {
    await this.db.put(
      "__chain_state__",
      JSON.stringify({
        sequenceNumber: this.sequenceNumber,
        lastHash: this.lastHash,
        chain: this.updateChain,
      })
    );
  }
}

interface SyncResult {
  updatesQueued: number;
  updates: QueuedUpdate[];
  chainIntegrity: boolean;
  syncDuration: number;
}
