import { describe, expect, it } from 'vitest';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { IslandModeManager } from './index.js';

describe('IslandModeManager', () => {
  it('queues, verifies, syncs, and flushes updates', async () => {
    const dir = mkdtempSync(join(tmpdir(), 'island-mode-'));
    const manager = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24
    });

    await manager.initialize();

    const queued = await manager.queueUpdate({ location: 'x' }, 'proof-1');
    expect(queued.position).toBe(1);

    const valid = await manager.verifyIntegrity();
    expect(valid).toBe(true);

    const sync = await manager.sync();
    expect(sync.chainIntegrity).toBe(true);
    expect(sync.updatesQueued).toBe(1);

    const statusBeforeFlush = manager.getStatus();
    expect(statusBeforeFlush.updatesQueued).toBe(1);
    expect(statusBeforeFlush.storageUsed).toBeGreaterThan(0);

    await manager.flush();

    const statusAfterFlush = manager.getStatus();
    expect(statusAfterFlush.updatesQueued).toBe(0);

    rmSync(dir, { recursive: true, force: true });
  });
});
