package io.sovereignmap.node

import android.util.Base64
import org.json.JSONObject
import java.nio.charset.StandardCharsets

object MobileGradientPayloadAdapter {
    data class CanonicalGradientPayload(
        val round: Int,
        val nodeId: String,
        val modelHashHex: String,
        val gradientChunkB64: String,
        val timestampUnixMs: Long,
    ) {
        fun canonicalBytes(): ByteArray {
            val canonical = "round=$round\n" +
                "node_id=$nodeId\n" +
                "model_hash=$modelHashHex\n" +
                "gradient_chunk_b64=$gradientChunkB64\n" +
                "timestamp_unix_ms=$timestampUnixMs"
            return canonical.toByteArray(StandardCharsets.UTF_8)
        }
    }

    data class SignedGradientEnvelope(
        val nodeId: String,
        val round: Int,
        val signerAlias: String,
        val publicKeyPem: String,
        val gradientPayloadB64: String,
        val gradientSignatureB64: String,
        val attestationPayloadB64: String,
    ) {
        fun asJson(): JSONObject {
            return JSONObject()
                .put("node_id", nodeId)
                .put("round", round)
                .put("signer_alias", signerAlias)
                .put("public_key_pem", publicKeyPem)
                .put("gradient_payload_b64", gradientPayloadB64)
                .put("gradient_signature_b64", gradientSignatureB64)
                .put("attestation_payload_b64", attestationPayloadB64)
        }
    }

    fun buildSignedEnvelope(
        nodeId: String,
        round: Int,
        modelHashHex: String,
        gradientChunk: ByteArray,
        signerAlias: String,
        signer: MobileHardwareSigner,
    ): SignedGradientEnvelope {
        val payload = CanonicalGradientPayload(
            round = round,
            nodeId = nodeId,
            modelHashHex = modelHashHex,
            gradientChunkB64 = Base64.encodeToString(gradientChunk, Base64.NO_WRAP),
            timestampUnixMs = System.currentTimeMillis(),
        )

        val canonicalBytes = payload.canonicalBytes()
        val signature = signer.sign(signerAlias, canonicalBytes)
        val attestation = signer.attestation(signerAlias)
        val publicKeyPem = signer.publicKeyPEM(signerAlias)

        return SignedGradientEnvelope(
            nodeId = nodeId,
            round = round,
            signerAlias = signerAlias,
            publicKeyPem = publicKeyPem,
            gradientPayloadB64 = Base64.encodeToString(canonicalBytes, Base64.NO_WRAP),
            gradientSignatureB64 = Base64.encodeToString(signature, Base64.NO_WRAP),
            attestationPayloadB64 = Base64.encodeToString(attestation, Base64.NO_WRAP),
        )
    }
}
