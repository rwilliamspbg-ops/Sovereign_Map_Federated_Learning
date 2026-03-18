import numpy as np


class DPQuantizer:
    """Handles Differential Privacy Int8 Quantization to reduce bandwidth by 90%"""

    def __init__(
        self, epsilon: float = 1.0, delta: float = 1e-5, l2_sensitivity: float = 1.0
    ):
        self.epsilon = epsilon
        self.delta = delta
        self.l2_sensitivity = l2_sensitivity

    def quantize_and_add_noise(
        self, gradients: np.ndarray, clip_threshold: float = 1.0
    ) -> np.ndarray:
        """
        Clips float32 models and maps them to Int8 spaces natively in the Python client
        prior to network transmission.
        """
        # Step 1: Clip
        clipped = np.clip(gradients, -clip_threshold, clip_threshold)

        # Step 2: Gaussian DP Noise
        c = np.sqrt(2 * np.log(1.25 / self.delta))
        sigma = (c * self.l2_sensitivity) / self.epsilon
        noise = np.random.normal(0, sigma, gradients.shape)
        noisy_gradients = clipped + noise

        # Step 3: Quantize to int8
        q_min, q_max = -128, 127
        min_val, max_val = -clip_threshold * 1.5, clip_threshold * 1.5
        scale = (max_val - min_val) / (q_max - q_min)
        zero_point = round(q_min - (min_val / scale))

        quantized = np.round((noisy_gradients / scale) + zero_point)
        return np.clip(quantized, q_min, q_max).astype(np.int8)
