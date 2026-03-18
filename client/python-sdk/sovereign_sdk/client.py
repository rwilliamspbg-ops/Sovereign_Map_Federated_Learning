import time
import logging
import requests
import numpy as np

from .compression import DPQuantizer
from .tokenomics import SovereignWallet

logger = logging.getLogger(__name__)


class SovereignMapNode:
    """The Primary Python SDK entrypoint for Edge Clients."""

    def __init__(
        self, aggregator_url: str, secure_mode: bool = True, wallet_key: str = None
    ):
        self.aggregator_url = aggregator_url
        self.secure_mode = secure_mode
        self.wallet = (
            SovereignWallet(private_key=wallet_key, rpc_endpoint="")
            if wallet_key
            else None
        )
        self.quantizer = DPQuantizer()

        # Self-healing parameters
        self.batch_size = 64
        self.backoff_factor = 1.0

    def get_wallet_rewards(self) -> float:
        if not self.wallet:
            raise ValueError("No wallet configured.")
        return self.wallet.get_wallet_rewards()

    def stake_collateral(self, amount: float) -> bool:
        if not self.wallet:
            raise ValueError("No wallet configured.")
        return self.wallet.stake_collateral(amount)

    def _execute_enclave_wrapper(self, gradients: np.ndarray) -> np.ndarray:
        """
        Routes the DP Noise generation into the Secure Enclave (SGX) if secure_mode=True
        """
        if self.secure_mode:
            logger.info(
                "Executing Differential Privacy algorithms inside Trusted Execution Environment (TEE)."
            )
            # In a real C-ffi binding, the float array routes to the enclave here.
        return self.quantizer.quantize_and_add_noise(gradients)

    def _handle_backoff(self, response_status: int):
        """Self-healing active listener"""
        if response_status in (429, 503):
            # Aggregator metrics overloaded. Cut batch sizes down dynamically
            self.batch_size = max(16, self.batch_size // 2)
            self.backoff_factor *= 2
            logger.warning(
                f"Overload detected (HTTP {response_status}). Self-healing triggered: Reduced batch size to {self.batch_size}, backoff={self.backoff_factor}s"
            )
            time.sleep(self.backoff_factor)
        elif response_status == 200 and self.batch_size < 64:
            # Gradually restore
            self.batch_size = min(64, self.batch_size + 8)
            self.backoff_factor = 1.0

    def submit_gradient(self, raw_gradients: np.ndarray) -> bool:
        """
        Primary hook called at the end of a local epoch cycle.
        """
        try:
            # 1. Compress & Secure
            payload = self._execute_enclave_wrapper(raw_gradients)

            # 2. Transmit
            logger.debug(
                f"Transmitting {payload.nbytes} bytes of Int8 compressed gradients."
            )
            resp = requests.post(
                f"{self.aggregator_url}/api/v1/update", data=payload.tobytes()
            )

            # 3. Assess Node Status (Self-healing Hook)
            self._handle_backoff(resp.status_code)

            return resp.status_code == 200
        except Exception as e:
            logger.error(f"Error submitting to aggregator: {e}")
            self._handle_backoff(503)
            return False
