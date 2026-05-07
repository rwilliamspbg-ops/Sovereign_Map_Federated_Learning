import { describe, expect, it } from "vitest";
import type { GradientAck, GradientMessage } from "./index.js";
import {
  SDK_VERSION,
  PROTOCOL_VERSION,
  SGP001_VERSION,
  MOHAWK_GRADIENT_PROTOCOL_ID,
  MessageType,
  createLogger,
  SovereignMapError,
  NodeInitializationError,
  PrivacyBudgetExceededError,
} from "./index.js";

describe("core exports", () => {
  it("exposes stable SDK metadata", () => {
    expect(SDK_VERSION).toBe("0.1.0-alpha.1");
    expect(PROTOCOL_VERSION).toBe("1.0.0");
    expect(SGP001_VERSION).toBe("1.0.0");
  });

  it("exposes protocol message enum", () => {
    expect(MessageType.MapUpdate).toBe("map_update");
    expect(MessageType.Heartbeat).toBe("heartbeat");
  });

  it("exposes the upstream gradient stream protocol id", () => {
    expect(MOHAWK_GRADIENT_PROTOCOL_ID).toBe("/mohawk/gradient/1.0.0");
  });

  it("exposes gradient message and ack vocabulary", () => {
    const message: GradientMessage = {
      nodeId: "node-a",
      taskId: "task-a",
      round: 1,
      gradients: [0.1, 0.2, 0.3],
      timestampMs: 123,
    };
    const ack: GradientAck = {
      accepted: true,
      negotiatedKEX: "x25519",
      kexPublicKeyLen: 32,
      batchAccepted: 1,
      batchRejected: 0,
    };

    expect(message.taskId).toBe("task-a");
    expect(message.gradients).toHaveLength(3);
    expect(ack.accepted).toBe(true);
    expect(ack.kexPublicKeyLen).toBe(32);
  });

  it("creates logger and serializes custom errors", () => {
    const logger = createLogger({ level: "debug", nodeId: "node-test" });
    expect(logger).toBeTruthy();

    const rootError = new SovereignMapError("base error", "BASE");
    expect(rootError.toJSON().code).toBe("BASE");

    const initError = new NodeInitializationError("init failed");
    expect(initError.name).toBe("NodeInitializationError");

    const budgetError = new PrivacyBudgetExceededError(0.1, 0.2);
    expect(budgetError.currentBudget).toBe(0.1);
    expect(budgetError.requiredBudget).toBe(0.2);
  });
});
