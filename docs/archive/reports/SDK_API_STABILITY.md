# Sovereign Map SDK - API Stability Policy v1.0

This document defines the API stability guarantees and versioning strategy for the Sovereign Map SDK packages.

## Versioning Scheme

We follow **Semantic Versioning 2.0.0** with special handling for the **0.x.y phase**:

### Current Phase: 0.x.y (Alpha - Rapid Iteration)
- **0.x.y format**: Core SDK packages are in active development
- **API contracts**: Defined but subject to refinement based on early integrations
- **Deprecation window**: 2 minor versions for breaking changes (e.g., 0.5 → 0.7)
- **Release frequency**: 2-week cadence with stable snapshots

#### Version Format
- **0.1.0-alpha.N**: Pre-release builds with unstable APIs
- **0.2.0**: Stable public release with contracts locked
- **0.2.1**: Bug fixes and patches (no API changes)
- **0.3.0**: New features with backward compatibility guarantee

### Stability Tiers

#### **Tier 1: Stable (Core API)**
**Guarantee**: Will not break until major version (`1.0.0`)

| Package | Exports | Status |
|---------|---------|--------|
| `@sovereignmap/core` | `SovereignNode`, `NetworkClient`, `MessageType` | ✅ Stable in 0.2.0+ |
| `@sovereignmap/privacy` | `PrivacyEngine`, `PrivacyBudget` | ✅ Stable in 0.2.0+ |
| `@sovereignmap/consensus` | `ConsensusParticipant`, `HierarchicalAggregator` | ✅ Stable in 0.2.0+ |
| `@sovereignmap/island` | `IslandModeManager`, `QueuedUpdate` | ✅ Stable in 0.2.0+ |

**Rules**:
- No parameter reordering
- No type narrowing of union types (e.g., `string | number` → `string`)
- Only additive changes to interfaces (new optional properties)
- Error types: New errors may be added; existing error codes never change

#### **Tier 2: Subject to Change (SPIs/Extensions)**
**Guarantee**: May change with 2-minor-version notice + deprecation warning

| Package | Exports | Status |
|---------|---------|--------|
| `@sovereignmap/core` | `NodeMetricsCollector`, `NodeState` | ⚠️ May change in 0.3–0.5 |
| `@sovereignmap/consensus` | `ConsensusConfig` (internal quorum tuning) | ⚠️ May change in 0.3–0.5 |

**Rules**:
- Must emit TypeScript deprecation warnings 2 minor versions before removal
- Must update documentation with migration path
- Mark with `@deprecated` JSDoc tags with version and replacement

#### **Tier 3: Internal (Do Not Use)**
**Guarantee**: No backward compatibility

| Exports | Examples |
|---------|----------|
| Private methods (prefixed `_`) | `_initializeHardware()` (SovereignNode) |
| Test utilities | Mock classes in test files |
| Undocumented exports | Functions in index.ts but not mentioned in README |

## Public API Surface

### @sovereignmap/core
```typescript
// Classes (Tier 1 Stable)
export class SovereignNode { ... }
export class NetworkClient { ... }

// Type exports (Tier 1 Stable)
export type NodeConfig = { ... }
export type NodeStatus = { ... }
export type MapUpdate = { ... }

// Enums (Tier 1 Stable)
export enum NodeState { Online, Offline, Syncing }
export enum MessageType { ... }

// Errors (Tier 1 Stable)
export class SovereignMapError extends Error { ... }
export class NodeInitializationError extends SovereignMapError { ... }

// Constants (Tier 1 Stable)
export const SDK_VERSION = '0.2.0'
export const PROTOCOL_VERSION = '1.0.0'

// Utilities (Tier 1 Stable)
export function createLogger(name: string): Logger { ... }
```

### @sovereignmap/privacy
```typescript
// Classes (Tier 1 Stable)
export class PrivacyEngine { ... }

// Types (Tier 1 Stable)
export interface PrivacyBudget { epsilon: number; delta: number }
export interface PrivacyStatus { ... }
```

### @sovereignmap/consensus
```typescript
// Classes (Tier 1 Stable)
export class ConsensusParticipant { ... }
export class HierarchicalAggregator { ... }

// Config (Tier 2 Subject to Change)
export interface ConsensusConfig { ... }
```

### @sovereignmap/island
```typescript
// Classes (Tier 1 Stable)
export class IslandModeManager { ... }

// Types (Tier 1 Stable)
export interface IslandModeConfig { ... }
export interface QueuedUpdate { ... }
export interface IslandStatus { ... }
```

## Breaking Change Policy

### When to Make Breaking Changes
1. **Security fixes** that cannot be patched: Immediate major version
2. **Protocol incompatibility**: Requires agreement with blockchain validators
3. **Hardware requirements**: Only with network halt and migration window (6+ months notice)

### Breaking Change Announcement
1. Proposal with impact analysis
2. 2-week community feedback period
3. **Not before** version 1.0.0 (except critical security)
4. Announce in `docs/CHANGELOG.md` with migration guide
5. Create GitHub discussion issue

### Breaking Change Format
```
## [BREAKING] SovereignNode.initialize() signature changed in 1.0.0

**Migration**:
```typescript
// Before (0.x.y)
await node.initialize({ region, coords, budget })

// After (1.0.0)
await node.initialize({ region, bounds: { coords }, privacyBudget: budget })
```
```

## Deprecation Process

### Step 1: Add `@deprecated` JSDoc
```typescript
/**
 * @deprecated Use `newMethod()` instead. Will be removed in 0.5.0.
 */
async oldMethod(): Promise<void> { ... }
```

### Step 2: Emit Runtime Warning (Minor version - 1)
```typescript
console.warn(
  '[sovereignmap/core] oldMethod() is deprecated and will be removed in 0.5.0. Use newMethod() instead.'
);
```

### Step 3: Remove (2 minor versions later)
- Remove from index.ts exports
- Update SDK_VERSION constant
- Publish as new minor version
- Document in CHANGELOG.md

### Example Timeline
- **0.2.0**: Introduce `newMethod()`, add deprecation warning on `oldMethod()`
- **0.3.0**: Add @deprecated JSDoc tag
- **0.4.0**: Last version with `oldMethod()` included
- **0.5.0**: Remove `oldMethod()` completely

## Custom Export Points

The SDK provides convenient re-export paths for common use cases:

```typescript
// Root export (main entry point)
import { SovereignNode } from '@sovereignmap/core'

// Sub-exports
import { SovereignNode } from '@sovereignmap/core/node'
import { NetworkClient } from '@sovereignmap/core/network'

// Privacy module (separate package)
import { PrivacyEngine } from '@sovereignmap/privacy'

// Consensus (separate package)
import { ConsensusParticipant } from '@sovereignmap/consensus'

// Island mode (separate package)
import { IslandModeManager } from '@sovereignmap/island'
```

All sub-exports must mirror root exports for consistency.

## Binary Compatibility

The SDK does **not** provide binary/WASM compatibility guarantees:
- WASM builds are rebuilt with each patch version
- Hardware-accelerated paths may change based on TPM/hardware availability
- Fallback to pure-JavaScript implementations is always available

## TypeScript Version Support

| SDK Version | Min TypeScript |
|-------------|----------------|
| 0.x.y       | 5.3           |
| 1.0.0       | 5.3           |
| 2.0.0       | 5.5 (proposed)|

Upgrading TypeScript is **not a breaking change** if it's not required.

## Security Stability

**Security fixes are exempt from breaking change restrictions**:
- Zero-day vulnerabilities trigger immediate patch (0.x.y → 0.x.z)
- Supply-chain compromises trigger minor version bump (0.x.y → 0.(x+1).0)
- Cryptographic upgrades require major version (requires 1.0.0 → 2.0.0)

## Release Cadence

| Package | Cycle | SLA |
|---------|-------|-----|
| @sovereignmap/core | 2 weeks | Stable release on first Monday of sprint |
| @sovereignmap/privacy | 4 weeks | Aligned with core for coordinated security releases |
| @sovereignmap/consensus | 4 weeks | Milestone-driven (protocol agreement) |
| @sovereignmap/island | 2 weeks | Aligned with core clients |

## Feedback & Amendments

This policy is versioned in `SDK_API_STABILITY.md` and tracked in Git. Changes require:
1. GitHub discussion with API designers
2. Version bump in document header
3. Announcement in SDK CHANGELOG

Current version: **1.0** (2026-03, locked until 0.2.0 release)
