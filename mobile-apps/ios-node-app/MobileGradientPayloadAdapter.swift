import Foundation

struct CanonicalGradientPayload {
    let round: Int
    let nodeID: String
    let modelHashHex: String
    let gradientChunkB64: String
    let timestampUnixMS: Int64

    func canonicalBytes() -> Data {
        let canonical = "round=\(round)\nnode_id=\(nodeID)\nmodel_hash=\(modelHashHex)\ngradient_chunk_b64=\(gradientChunkB64)\ntimestamp_unix_ms=\(timestampUnixMS)"
        return Data(canonical.utf8)
    }
}

struct SignedGradientEnvelope {
    let nodeID: String
    let round: Int
    let signerAlias: String
    let publicKeyPEM: String
    let gradientPayloadB64: String
    let gradientSignatureB64: String
    let attestationPayloadB64: String

    func asJSONObject() -> [String: Any] {
        [
            "node_id": nodeID,
            "round": round,
            "signer_alias": signerAlias,
            "public_key_pem": publicKeyPEM,
            "gradient_payload_b64": gradientPayloadB64,
            "gradient_signature_b64": gradientSignatureB64,
            "attestation_payload_b64": attestationPayloadB64,
        ]
    }

    func asJSONString() -> String {
        guard let data = try? JSONSerialization.data(withJSONObject: asJSONObject(), options: [.sortedKeys]) else {
            return "{}"
        }
        return String(data: data, encoding: .utf8) ?? "{}"
    }
}

enum MobilePayloadAdapter {
    static func buildSignedEnvelope(
        nodeID: String,
        round: Int,
        modelHashHex: String,
        gradientChunk: Data,
        signerAlias: String,
        signer: MobileHardwareSigner
    ) throws -> SignedGradientEnvelope {
        let payload = CanonicalGradientPayload(
            round: round,
            nodeID: nodeID,
            modelHashHex: modelHashHex,
            gradientChunkB64: gradientChunk.base64EncodedString(),
            timestampUnixMS: Int64(Date().timeIntervalSince1970 * 1000)
        )
        let canonicalBytes = payload.canonicalBytes()
        let signature = try signer.sign(alias: signerAlias, payload: canonicalBytes)
        let attestation = try signer.attestation(alias: signerAlias)
        let publicKeyPEM = try signer.publicKeyPEM(alias: signerAlias)

        return SignedGradientEnvelope(
            nodeID: nodeID,
            round: round,
            signerAlias: signerAlias,
            publicKeyPEM: publicKeyPEM,
            gradientPayloadB64: canonicalBytes.base64EncodedString(),
            gradientSignatureB64: signature.base64EncodedString(),
            attestationPayloadB64: attestation.base64EncodedString()
        )
    }
}
