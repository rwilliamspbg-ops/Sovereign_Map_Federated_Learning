import base64
import hashlib
import importlib
import sys
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


def _canonical_payload(
    round_id: int, node_id: str, model_hash: str, gradient_chunk: bytes, ts_ms: int
) -> bytes:
    chunk_b64 = base64.b64encode(gradient_chunk).decode("ascii")
    canonical = (
        f"round={round_id}\n"
        f"node_id={node_id}\n"
        f"model_hash={model_hash}\n"
        f"gradient_chunk_b64={chunk_b64}\n"
        f"timestamp_unix_ms={ts_ms}"
    )
    return canonical.encode("utf-8")


class MobileVerifyGradientContractTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        try:
            cls.ecdsa = importlib.import_module("ecdsa")
            importlib.import_module("flwr")
            cls.backend = importlib.import_module("sovereignmap_production_backend_v2")
        except Exception as exc:
            raise unittest.SkipTest(f"contract test dependencies unavailable: {exc}")

    def _signed_request_payload(self) -> dict:
        sk = self.ecdsa.SigningKey.generate(curve=self.ecdsa.NIST256p)
        vk = sk.get_verifying_key()

        canonical_bytes = _canonical_payload(
            round_id=7,
            node_id="node-7777",
            model_hash="cifar10-v1",
            gradient_chunk=b"mock-gradient-chunk",
            ts_ms=1711180800000,
        )
        signature_der = sk.sign_deterministic(
            canonical_bytes,
            hashfunc=hashlib.sha256,
            sigencode=self.ecdsa.util.sigencode_der,
        )

        return {
            "node_id": "node-7777",
            "round": 7,
            "signer_alias": "mohawk.mobile.identity",
            "public_key_pem": vk.to_pem().decode("utf-8"),
            "gradient_payload_b64": base64.b64encode(canonical_bytes).decode("ascii"),
            "gradient_signature_b64": base64.b64encode(signature_der).decode("ascii"),
            "attestation_payload_b64": base64.b64encode(
                b'{"secure_hardware":"test"}'
            ).decode("ascii"),
        }

    def test_accepts_valid_signed_payload(self):
        self.backend.os.environ["MOBILE_REQUIRE_ATTESTATION"] = "true"

        client = self.backend.app.test_client()
        response = client.post(
            "/mobile/verify_gradient", json=self._signed_request_payload()
        )

        self.assertEqual(response.status_code, 200)
        body = response.get_json()
        self.assertTrue(body["accepted"])
        self.assertEqual(body["reason"], "ok")
        self.assertEqual(body["details"]["node_id"], "node-7777")
        self.assertTrue(body["details"]["attestation_present"])

    def test_rejects_bad_signature(self):
        self.backend.os.environ["MOBILE_REQUIRE_ATTESTATION"] = "true"

        payload = self._signed_request_payload()
        payload["gradient_signature_b64"] = base64.b64encode(
            b"broken-signature"
        ).decode("ascii")

        client = self.backend.app.test_client()
        response = client.post("/mobile/verify_gradient", json=payload)

        self.assertEqual(response.status_code, 400)
        body = response.get_json()
        self.assertFalse(body["accepted"])
        self.assertEqual(body["reason"], "signature_verification_failed")

    def test_rejects_missing_attestation_when_required(self):
        self.backend.os.environ["MOBILE_REQUIRE_ATTESTATION"] = "true"

        payload = self._signed_request_payload()
        payload.pop("attestation_payload_b64")

        client = self.backend.app.test_client()
        response = client.post("/mobile/verify_gradient", json=payload)

        self.assertEqual(response.status_code, 400)
        body = response.get_json()
        self.assertFalse(body["accepted"])
        self.assertEqual(body["reason"], "missing_attestation")


if __name__ == "__main__":
    unittest.main()
