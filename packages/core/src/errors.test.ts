import { describe, expect, it } from "vitest";
import {
  ConsensusError,
  HardwareAttestationError,
  IslandModeError,
  NetworkError,
  SovereignMapError,
  ZKProofError,
} from "./errors.js";

describe("errors", () => {
  it("serializes with and without Error cause", () => {
    const caused = new SovereignMapError("boom", "E1", true, new Error("root"));
    expect(caused.toJSON().cause).toBe("root");

    const nonErrorCause = new SovereignMapError(
      "boom2",
      "E2",
      true,
      "bad-cause" as any
    );
    expect(nonErrorCause.toJSON().cause).toBeUndefined();
  });

  it("constructs all domain-specific error classes", () => {
    const consensus = new ConsensusError("consensus failed", "r42");
    expect(consensus.code).toBe("CONSENSUS_ERROR");
    expect(consensus.roundId).toBe("r42");

    const network = new NetworkError("network failed", "/peer/1", 503);
    expect(network.code).toBe("NETWORK_ERROR");
    expect(network.endpoint).toBe("/peer/1");
    expect(network.statusCode).toBe(503);

    const hw = new HardwareAttestationError("tpm failed", "/dev/tpm0");
    expect(hw.code).toBe("HW_ATTESTATION_ERROR");
    expect(hw.devicePath).toBe("/dev/tpm0");

    const island = new IslandModeError("island failed");
    expect(island.code).toBe("ISLAND_MODE_ERROR");

    const zk = new ZKProofError("zk failed");
    expect(zk.code).toBe("ZK_PROOF_ERROR");
  });
});
