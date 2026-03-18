import { describe, expect, it } from "vitest";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { IslandModeManager } from "./index.js";

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

    const raw = await (manager as any).db.get(`update:${first.id}`);
    const parsed = JSON.parse(raw);
    parsed.previousHash = "tampered";
    await (manager as any).db.put(`update:${first.id}`, JSON.stringify(parsed));

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

    const raw = await (manager as any).db.get(`update:${queued.id}`);
    const parsed = JSON.parse(raw);
    parsed.chainedHash = "tampered-hash";
    await (manager as any).db.put(
      `update:${queued.id}`,
      JSON.stringify(parsed)
    );

    await expect(manager.sync()).rejects.toThrow(
      "Chain integrity violated at position 0"
    );

    rmSync(dir, { recursive: true, force: true });
  });
});
