import Foundation
import CryptoKit
import Security

enum MobileSignerError: Error {
    case keyGenerationFailed
    case signatureFailed
    case publicKeyUnavailable
    case encodingFailed
}

protocol MobileHardwareSigner {
    func sign(alias: String, payload: Data) throws -> Data
    func attestation(alias: String) throws -> Data
    func publicKeyPEM(alias: String) throws -> String
}

final class SecureEnclaveSigner: MobileHardwareSigner {
    private let keyType = kSecAttrKeyTypeECSECPrimeRandom
    private let keySize = 256

    private func keyTag(for alias: String) -> Data {
        Data(alias.utf8)
    }

    private func loadPrivateKey(alias: String) -> SecKey? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrApplicationTag as String: keyTag(for: alias),
            kSecAttrKeyType as String: keyType,
            kSecReturnRef as String: true,
        ]

        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        if status == errSecSuccess {
            return (item as! SecKey)
        }
        return nil
    }

    private func createPrivateKey(alias: String) throws -> SecKey {
        var accessControlError: Unmanaged<CFError>?
        guard let access = SecAccessControlCreateWithFlags(
            nil,
            kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            .privateKeyUsage,
            &accessControlError
        ) else {
            throw accessControlError?.takeRetainedValue() ?? MobileSignerError.keyGenerationFailed
        }

        var privateAttrs: [String: Any] = [
            kSecAttrIsPermanent as String: true,
            kSecAttrApplicationTag as String: keyTag(for: alias),
            kSecAttrAccessControl as String: access,
        ]

#if !targetEnvironment(simulator)
        privateAttrs[kSecAttrTokenID as String] = kSecAttrTokenIDSecureEnclave
#endif

        let attributes: [String: Any] = [
            kSecAttrKeyType as String: keyType,
            kSecAttrKeySizeInBits as String: keySize,
            kSecPrivateKeyAttrs as String: privateAttrs,
        ]

        var keyError: Unmanaged<CFError>?
        guard let key = SecKeyCreateRandomKey(attributes as CFDictionary, &keyError) else {
            throw keyError?.takeRetainedValue() ?? MobileSignerError.keyGenerationFailed
        }
        return key
    }

    private func privateKey(alias: String) throws -> SecKey {
        if let existing = loadPrivateKey(alias: alias) {
            return existing
        }
        return try createPrivateKey(alias: alias)
    }

    func sign(alias: String, payload: Data) throws -> Data {
        let key = try privateKey(alias: alias)
        let algorithm = SecKeyAlgorithm.ecdsaSignatureMessageX962SHA256

        guard SecKeyIsAlgorithmSupported(key, .sign, algorithm) else {
            throw MobileSignerError.signatureFailed
        }

        var signError: Unmanaged<CFError>?
        guard let signature = SecKeyCreateSignature(
            key,
            algorithm,
            payload as CFData,
            &signError
        ) else {
            throw signError?.takeRetainedValue() ?? MobileSignerError.signatureFailed
        }

        return signature as Data
    }

    func attestation(alias: String) throws -> Data {
        let key = try privateKey(alias: alias)
        guard let pub = SecKeyCopyPublicKey(key) else {
            throw MobileSignerError.publicKeyUnavailable
        }

        var pubErr: Unmanaged<CFError>?
        guard let pubData = SecKeyCopyExternalRepresentation(pub, &pubErr) as Data? else {
            throw pubErr?.takeRetainedValue() ?? MobileSignerError.publicKeyUnavailable
        }

        let digest = SHA256.hash(data: pubData)
        let payload: [String: Any] = [
            "alias": alias,
#if targetEnvironment(simulator)
            "secure_hardware": "simulator_software_fallback",
#else
            "secure_hardware": "secure_enclave",
#endif
            "public_key_b64": pubData.base64EncodedString(),
            "public_key_sha256": Data(digest).base64EncodedString(),
            "generated_at_unix": Int(Date().timeIntervalSince1970),
        ]

        guard JSONSerialization.isValidJSONObject(payload) else {
            throw MobileSignerError.encodingFailed
        }
        return try JSONSerialization.data(withJSONObject: payload, options: [])
    }

    func publicKeyPEM(alias: String) throws -> String {
        let key = try privateKey(alias: alias)
        guard let pub = SecKeyCopyPublicKey(key) else {
            throw MobileSignerError.publicKeyUnavailable
        }

        var pubErr: Unmanaged<CFError>?
        guard let rawPublicKey = SecKeyCopyExternalRepresentation(pub, &pubErr) as Data? else {
            throw pubErr?.takeRetainedValue() ?? MobileSignerError.publicKeyUnavailable
        }

        // P-256 SubjectPublicKeyInfo prefix for uncompressed EC point.
        let spkiPrefix = Data([
            0x30, 0x59,
            0x30, 0x13,
            0x06, 0x07, 0x2A, 0x86, 0x48, 0xCE, 0x3D, 0x02, 0x01,
            0x06, 0x08, 0x2A, 0x86, 0x48, 0xCE, 0x3D, 0x03, 0x01, 0x07,
            0x03, 0x42, 0x00,
        ])
        let spki = spkiPrefix + rawPublicKey
        let base64Body = spki.base64EncodedString(options: [.lineLength64Characters])
        return "-----BEGIN PUBLIC KEY-----\n\(base64Body)\n-----END PUBLIC KEY-----"
    }
}
