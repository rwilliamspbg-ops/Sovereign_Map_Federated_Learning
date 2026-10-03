import { describe, expect, it } from "vitest";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { IslandModeManager, QueuedUpdate } from "./index.js";

describe("IslandModeManager", () => {
  it("queues, verifies, syncs, and flushes updates", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-"));
    const manager = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });

    await manager.initialize();

    const queued = await manager.queueUpdate({ location: "x" }, "proof-1");
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

  it("loads existing chain state on initialize", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-load-"));

    const first = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });
    await first.initialize();
    await first.queueUpdate({ location: "a" }, "proof-a");
    await (first as any).db.close();

    const second = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });
    await second.initialize();

    const status = second.getStatus();
    expect(status.updatesQueued).toBe(1);
    expect(status.chainIntegrity).toBe(true);

    await (second as any).db.close();

    rmSync(dir, { recursive: true, force: true });
  });

  it("throws when queueing update while island mode is disabled", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-disabled-"));
    const manager = new IslandModeManager({
      enabled: false,
      storagePath: dir,
      maxOfflineHours: 24,
    });

    await manager.initialize();
    await expect(
      manager.queueUpdate({ location: "x" }, "proof-x")
    ).rejects.toThrow("Island Mode not enabled");

    rmSync(dir, { recursive: true, force: true });
  });

  it("detects tampering via previous hash mismatch", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-tamper-prev-"));
    const manager = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });

    await manager.initialize();
    const first = await manager.queueUpdate({ location: "x" }, "proof-1");
    await manager.queueUpdate({ location: "y" }, "proof-2");

    const key = `update:${first.position.toString().padStart(12, "0")}:${first.id}`;
    const raw = await (manager as any).db.get(key);
    const parsed = JSON.parse(raw);
    parsed.previousHash = "tampered";
    await (manager as any).db.put(key, JSON.stringify(parsed));

    await expect(manager.sync()).rejects.toThrow(
      "Chain integrity violated at position 0"
    );
    expect(await manager.verifyIntegrity()).toBe(false);

    rmSync(dir, { recursive: true, force: true });
  });

  it("detects tampering via chained hash mismatch", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-tamper-hash-"));
    const manager = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });

    await manager.initialize();
    const queued = await manager.queueUpdate({ location: "z" }, "proof-z");

    const key = `update:${queued.position.toString().padStart(12, "0")}:${queued.id}`;
    const raw = await (manager as any).db.get(key);
    const parsed = JSON.parse(raw);
    parsed.chainedHash = "tampered-hash";
    await (manager as any).db.put(key, JSON.stringify(parsed));

    await expect(manager.sync()).rejects.toThrow(
      "Chain integrity violated at position 0"
    );

    rmSync(dir, { recursive: true, force: true });
  });

  it("streams updates sequentially with custom handler without array accumulation", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-stream-"));
    const manager = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });

    await manager.initialize();
    await manager.queueUpdate({ location: "p1" }, "proof-p1");
    await manager.queueUpdate({ location: "p2" }, "proof-p2");

    const handled: QueuedUpdate[] = [];
    const result = await manager.sync((update) => {
      handled.push(update);
    });

    expect(result.updatesQueued).toBe(2);
    expect(result.updates).toEqual([]); // No array accumulation in SyncResult when handler is passed
    expect(handled.length).toBe(2);
    expect(handled[0].proof).toBe("proof-p1");
    expect(handled[1].proof).toBe("proof-p2");

    rmSync(dir, { recursive: true, force: true });
  });

  it("supports async generator streamUpdates()", async () => {
    const dir = mkdtempSync(join(tmpdir(), "island-mode-generator-"));
    const manager = new IslandModeManager({
      enabled: true,
      storagePath: dir,
      maxOfflineHours: 24,
    });

    await manager.initialize();
    await manager.queueUpdate({ data: 1 }, "proof-1");
    await manager.queueUpdate({ data: 2 }, "proof-2");

    const items: QueuedUpdate[] = [];
    for await (const update of manager.streamUpdates()) {
      items.push(update);
    }

    expect(items.length).toBe(2);
    expect(items[0].sequenceNumber).toBe(1);
    expect(items[1].sequenceNumber).toBe(2);

    rmSync(dir, { recursive: true, force: true });
  });
});
