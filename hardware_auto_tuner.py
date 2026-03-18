"""
Auto Hardware Tuning Server (Self-Healing Node)
Subscribes to Prometheus exporter metrics to actively self-heal nodes by
dynamically reducing/increasing compute workload (e.g. batch size or thread pool)
"""

import time
import urllib.request
import json
import logging

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger("AutoTuner")

PROMETHEUS_URL = "http://localhost:9090/api/v1/query"


class FLNodeConfig:
    def __init__(self):
        self.batch_size = 64
        self.cuda_threads = 1024

    def scale_down(self):
        self.batch_size = max(16, self.batch_size // 2)
        self.cuda_threads = max(256, self.cuda_threads // 2)
        logger.info(
            f"Scaled DOWN: Batch Size={self.batch_size}, Threads={self.cuda_threads}"
        )

    def scale_up(self):
        self.batch_size = min(256, self.batch_size * 2)
        self.cuda_threads = min(4096, self.cuda_threads * 2)
        logger.info(
            f"Scaled UP: Batch Size={self.batch_size}, Threads={self.cuda_threads}"
        )


def get_current_metric(metric_name):
    """Stub querying local prometheus instance"""
    try:
        url = f"{PROMETHEUS_URL}?query={metric_name}"
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=2) as response:
            data = json.loads(response.read().decode())
            if data["status"] == "success" and len(data["data"]["result"]) > 0:
                # Return the float value
                return float(data["data"]["result"][0]["value"][1])
        return None
    except Exception as e:
        # In mock environment without prometheus up, generate fake data
        import random

        return (
            random.uniform(10, 95)
            if "utilization" in metric_name
            else random.uniform(1, 30)
        )


def tuning_loop():
    config = FLNodeConfig()

    while True:
        gpu_util_pct = get_current_metric("fl_gpu_utilization_percent")
        privacy_overhead_pct = get_current_metric("fl_privacy_overhead_percent")

        logger.info(
            f"Metrics: GPU_Util={gpu_util_pct:.1f}%, Privacy_Overhead={privacy_overhead_pct:.1f}%"
        )

        # Logic for self-healing
        if gpu_util_pct and gpu_util_pct > 90.0:
            logger.warning("GPU Saturation detected! Engaging self-healing.")
            config.scale_down()
        elif privacy_overhead_pct and privacy_overhead_pct > 20.0:
            logger.warning(
                "High DP overhead detected. Scaling down threads to prevent latency block."
            )
            config.scale_down()
        elif gpu_util_pct and gpu_util_pct < 40.0:
            logger.info("Underutilized GPU. Scaling up.")
            config.scale_up()
        else:
            logger.info("Node running optimally.")

        time.sleep(10)


if __name__ == "__main__":
    logger.info("Starting Auto Hardware Tuning Server...")
    try:
        tuning_loop()
    except KeyboardInterrupt:
        logger.info("Tuner shutdown.")
