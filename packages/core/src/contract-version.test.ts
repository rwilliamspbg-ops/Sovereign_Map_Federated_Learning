import { describe, expect, it } from "vitest";
import { AV_SENSOR_CONTRACT_VERSION } from "./types.js";

describe("AV sensor contract version", () => {
  it("is pinned to av-v1 for ingest compatibility", () => {
    expect(AV_SENSOR_CONTRACT_VERSION).toBe("av-v1");
  });
});
