package io.sovereignmap.node

import android.os.Build
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import org.json.JSONArray
import org.json.JSONObject
import java.nio.charset.StandardCharsets
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.Signature
import java.security.cert.X509Certificate
import java.security.spec.ECGenParameterSpec

interface MobileHardwareSigner {
    fun sign(alias: String, payload: ByteArray): ByteArray
    fun attestation(alias: String): ByteArray
}

class StrongBoxSigner : MobileHardwareSigner {
    private val keyStoreType = "AndroidKeyStore"

    private fun ensureKey(alias: String): Boolean {
        val keyStore = KeyStore.getInstance(keyStoreType).apply { load(null) }
        if (keyStore.containsAlias(alias)) {
            return false
        }

        val generator = KeyPairGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_EC,
            keyStoreType,
        )

        val builder = KeyGenParameterSpec.Builder(
            alias,
            KeyProperties.PURPOSE_SIGN or KeyProperties.PURPOSE_VERIFY,
        )
            .setAlgorithmParameterSpec(ECGenParameterSpec("secp256r1"))
            .setDigests(KeyProperties.DIGEST_SHA256)
            .setUserAuthenticationRequired(false)

        var strongBoxBacked = false
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            try {
                builder.setIsStrongBoxBacked(true)
                strongBoxBacked = true
            } catch (_: Exception) {
                strongBoxBacked = false
            }
        }

        try {
            generator.initialize(builder.build())
            generator.generateKeyPair()
            return strongBoxBacked
        } catch (strongBoxErr: Exception) {
            // Fallback to non-StrongBox hardware-backed keystore key.
            val fallbackBuilder = KeyGenParameterSpec.Builder(
                alias,
                KeyProperties.PURPOSE_SIGN or KeyProperties.PURPOSE_VERIFY,
            )
                .setAlgorithmParameterSpec(ECGenParameterSpec("secp256r1"))
                .setDigests(KeyProperties.DIGEST_SHA256)
                .setUserAuthenticationRequired(false)

            generator.initialize(fallbackBuilder.build())
            generator.generateKeyPair()
            return false
        }
    }

    override fun sign(alias: String, payload: ByteArray): ByteArray {
        ensureKey(alias)
        val keyStore = KeyStore.getInstance(keyStoreType).apply { load(null) }
        val privateKey = keyStore.getKey(alias, null)
            ?: throw IllegalStateException("Keystore private key missing for alias=$alias")

        val signature = Signature.getInstance("SHA256withECDSA")
        signature.initSign(privateKey as java.security.PrivateKey)
        signature.update(payload)
        return signature.sign()
    }

    override fun attestation(alias: String): ByteArray {
        val strongBoxBacked = ensureKey(alias)
        val keyStore = KeyStore.getInstance(keyStoreType).apply { load(null) }
        val chain = keyStore.getCertificateChain(alias)
            ?: throw IllegalStateException("Keystore certificate chain missing for alias=$alias")

        val certArray = JSONArray()
        chain.forEach { cert ->
            certArray.put(Base64.encodeToString(cert.encoded, Base64.NO_WRAP))
        }

        val leaf = chain.firstOrNull() as? X509Certificate
        val payload = JSONObject()
            .put("alias", alias)
            .put("secure_hardware", if (strongBoxBacked) "strongbox" else "android_keystore")
            .put("attestation_cert_chain_b64", certArray)
            .put("leaf_subject", leaf?.subjectX500Principal?.name ?: "unknown")
            .put("generated_at_unix", System.currentTimeMillis() / 1000)

        return payload.toString().toByteArray(StandardCharsets.UTF_8)
    }
}
