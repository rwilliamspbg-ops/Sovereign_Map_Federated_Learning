# Secure Enclave / StrongBox Wrapper Plan

This document defines the wrapper contract that replaces TPM stubs for mobile nodes while preserving existing Sovereign attestation and mTLS semantics.

## 1) Integration Goal

- Keep private keys non-exportable and generated inside Secure Enclave (iOS) or StrongBox Keymaster/Keystore (Android).
- Sign every gradient payload prior to upload.
- Attach device attestation evidence so backend policy can verify hardware-rooted identity.

## 2) Go Bridge Contract

The Go mobile bridge file is implemented in:

- go-mobile/sovereignmapclient/pkg/client/mohawk_node_bridge.go

Core contract:

- `RegisterHardwareBackedSigner(signer HardwareBackedSigner)`
- `SignGradient(payload []byte) (SignedGradientUpdate, error)`

Platform wrappers must provide:

- `Sign(alias string, payload []byte) ([]byte, error)`
- `Attestation(alias string) ([]byte, error)`

## 3) iOS Wrapper (Secure Enclave)

Suggested Swift wrapper surface:

```swift
protocol MobileHardwareSigner {
    func sign(alias: String, payload: Data) throws -> Data
    func attestation(alias: String) throws -> Data
}
```

Implementation notes:

- Generate P-256 key with `kSecAttrTokenIDSecureEnclave`.
- Store key under alias `mohawk.mobile.identity`.
- Use `SecKeyCreateSignature` with ECDSA SHA-256.
- Obtain attestation/certificate chain where available and encode as CBOR or base64 JSON blob.
- Expose methods to gomobile bridge through a thin binding layer.

## 4) Android Wrapper (StrongBox)

Suggested Kotlin wrapper surface:

```kotlin
interface MobileHardwareSigner {
    fun sign(alias: String, payload: ByteArray): ByteArray
    fun attestation(alias: String): ByteArray
}
```

Implementation notes:

- Generate key in Android Keystore using `setIsStrongBoxBacked(true)` when available.
- Use `EC` key, SHA-256 digest, deterministic alias `mohawk.mobile.identity`.
- Sign via `Signature.getInstance("SHA256withECDSA")` bound to keystore private key.
- Pull key attestation cert chain from `KeyStore.getCertificateChain(alias)`.
- If StrongBox unavailable, report fallback mode to capabilities and still require non-exportable key.

## 5) Backend Verification Path

The backend verifier should:

1. Verify signature over exact gradient payload bytes.
2. Validate attestation chain and key usage constraints.
3. Ensure certificate/public key matches mTLS client identity.
4. Cache trust score impacts by attestation freshness.

## 6) Payload Canonicalization

Before signing, canonicalize payload using a deterministic format:

- `round` (int32)
- `node_id` (string)
- `model_hash` (sha256 bytes)
- `gradient_chunk` (bytes)
- `timestamp_unix_ms` (int64)

Canonicalized bytes should be identical across iOS and Android to avoid signature verification drift.

## 7) Security Fallback Policy

- Accept software-backed key only if policy explicitly allows degraded mode.
- Mark degraded mode in capabilities and reduce trust score.
- Never allow unsigned gradient updates.

## 8) Rollout Sequence

1. Ship wrappers for signing and attestation on each platform.
2. Register wrapper with `RegisterHardwareBackedSigner` at app startup.
3. Enable `sign_every_gradient` and enforce verification server-side.
4. Add telemetry for signer availability, sign latency, and verification failure reasons.
