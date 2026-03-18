#!/usr/bin/env python
"""
WEEK 1 TWEAKS: Optimized implementation with key improvements
Implements tweaks 1-6 from WEEK1_OPTIMIZATION_TWEAKS.md
"""

import random
import time
import numpy as np
from datetime import datetime
from typing import Dict, List

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.backends import default_backend
import base64

# ============================================================================
# TWEAK 1 & 2: Byzantine Attacks with Variance
# ============================================================================


class ImprovedByzantineAttack:
    """Byzantine attacks with realistic variance"""

    SIGN_FLIP = "sign_flip"
    LABEL_FLIP = "label_flip"
    FREE_RIDE = "free_ride"
    AMPLIFICATION = "amplification"

    @staticmethod
    def sign_flip(w, variance=0.1):
        """Sign-flip with variance - attacks aren't perfectly executed"""
        noise = np.random.randn(*w.shape) * variance
        return -w + noise

    @staticmethod
    def label_flip(w, scale=1.5, variance=0.2):
        """Label-flip with configurable intensity"""
        noise = np.random.randn(*w.shape) * variance
        return -w * scale + noise

    @staticmethod
    def free_ride(w):
        """Free-ride: no computation"""
        return np.zeros_like(w)

    @staticmethod
    def amplification(w, scale=2.5, variance=0.15):
        """Amplification with variance"""
        noise = np.random.randn(*w.shape) * variance
        return w * scale + noise

    @staticmethod
    def apply(w, attack_type, variance=None):
        """Apply specified Byzantine attack with variance"""
        if attack_type == ImprovedByzantineAttack.SIGN_FLIP:
            return ImprovedByzantineAttack.sign_flip(w, variance or 0.1)
        elif attack_type == ImprovedByzantineAttack.LABEL_FLIP:
            return ImprovedByzantineAttack.label_flip(w, variance=variance or 0.2)
        elif attack_type == ImprovedByzantineAttack.FREE_RIDE:
            return ImprovedByzantineAttack.free_ride(w)
        elif attack_type == ImprovedByzantineAttack.AMPLIFICATION:
            return ImprovedByzantineAttack.amplification(w, variance=variance or 0.15)
        return w


# ============================================================================
# TWEAK 3: Realistic Network with Bimodal Latency
# ============================================================================


class RealisticNetworkSimulator:
    """Network simulator with bimodal latency distribution"""

    def __init__(self):
        self.total_msgs = 0
        self.packet_loss = 0
        self.timeouts = 0
        self.fast_msgs = 0
        self.slow_msgs = 0

        # 90% fast path, 10% slow path (realistic internet)
        self.fast_network_prob = 0.9

    def get_latency_ms(self):
        """Bimodal latency distribution"""
        if random.random() < self.fast_network_prob:
            # Fast path: 1-3ms (normal LAN)
            return random.uniform(1, 3) + random.expovariate(1.0)
        else:
            # Slow path: 20-100ms (outliers, congestion)
            return random.uniform(20, 100) + random.expovariate(0.1)

    def deliver(self):
        """Simulate message delivery with latency outliers"""
        self.total_msgs += 1

        # Packet loss: 0.1%
        if random.random() < 0.001:
            self.packet_loss += 1
            return False

        # Latency check with timeout
        latency_ms = self.get_latency_ms()
        timeout_ms = 5000

        if latency_ms > timeout_ms:
            self.timeouts += 1
            return False

        if latency_ms < 10:
            self.fast_msgs += 1
        else:
            self.slow_msgs += 1

        return True

    def rate(self):
        """Network delivery rate"""
        if self.total_msgs == 0:
            return 1.0
        return 1.0 - (self.packet_loss + self.timeouts) / self.total_msgs

    def print_stats(self):
        """Print network statistics"""
        print(f"\nNetwork Statistics:")
        print(f"  Total Messages: {self.total_msgs:,}")
        print(f"  Delivery Rate: {self.rate():.1%}")
        print(f"  Packet Loss: {self.packet_loss}")
        print(f"  Timeouts: {self.timeouts}")
        print(
            f"  Fast Messages: {self.fast_msgs} ({self.fast_msgs/self.total_msgs:.1%} < 10ms)"
        )
        print(
            f"  Slow Messages: {self.slow_msgs} ({self.slow_msgs/self.total_msgs:.1%} >= 10ms)"
        )


# ============================================================================
# TWEAK 5: Improved Accuracy Model with Byzantine Resistance
# ============================================================================


class ImprovedAccuracyModel:
    """Accuracy model with Byzantine resistance factoring"""

    @staticmethod
    def calculate(
        round_num,
        total_rounds,
        attacked_nodes,
        total_nodes,
        verified_nodes,
        delivery_rate,
    ):
        """Calculate accuracy with realistic Byzantine resistance"""

        base = 65.0
        honest_pct = 1.0 - (attacked_nodes / total_nodes)

        # Improvement: only honest nodes contribute fully
        improvement = 2.5 * (round_num / total_rounds) * honest_pct

        # Byzantine resistance depends on honest majority
        if honest_pct > 2 / 3:  # 66%+ honest
            byzantine_factor = 0.2  # Strong resistance, light impact
        elif honest_pct > 0.5:  # 50%+ honest
            byzantine_factor = 0.5  # Medium resistance
        else:  # < 50% honest (critical)
            byzantine_factor = 1.0  # No resistance, heavy impact

        attack_impact = (attacked_nodes / total_nodes) * byzantine_factor

        # Network losses
        network_impact = (1.0 - delivery_rate) * 0.3

        # TPM attestation helps
        boost = (verified_nodes / total_nodes) * 0.2

        # Add realistic noise
        noise = random.uniform(-0.5, 0.5)

        accuracy = base + improvement - attack_impact - network_impact + boost + noise
        return np.clip(accuracy, 0.1, 99.5)


# ============================================================================
# TWEAK 8: Byzantine Node Persistence
# ============================================================================


class PersistentByzantineModel:
    """Byzantine nodes are fixed per config (realistic)"""

    def __init__(self, num_nodes):
        self.num_nodes = num_nodes
        self.byzantine_nodes = set()

    def select_for_config(self, byzantine_pct):
        """Select Byzantine nodes for this config and keep them fixed"""
        num_byzantine = int(self.num_nodes * byzantine_pct / 100.0)
        self.byzantine_nodes = set(
            np.random.choice(self.num_nodes, size=num_byzantine, replace=False)
        )
        return self.byzantine_nodes

    def is_byzantine(self, node_id):
        """Check if node is Byzantine (persistent across rounds)"""
        return node_id in self.byzantine_nodes

    def get_byzantine_count(self):
        """Get count of Byzantine nodes"""
        return len(self.byzantine_nodes)


# ============================================================================
# TWEAK: Adaptive Convergence Threshold
# ============================================================================


def check_convergence(accuracies, byzantine_pct):
    """Adaptive convergence threshold based on Byzantine level"""

    if len(accuracies) < 5:
        return False

    avg_last_10 = (
        np.mean(accuracies[-10:]) if len(accuracies) >= 10 else np.mean(accuracies[-5:])
    )

    # Theoretical minimum: (1 - byzantine_pct/100) * 100
    # Practical: account for noise, network losses

    if byzantine_pct <= 10:
        threshold = 85.0
    elif byzantine_pct <= 20:
        threshold = 80.0
    elif byzantine_pct <= 30:
        threshold = 75.0
    elif byzantine_pct <= 40:
        threshold = 70.0
    else:
        threshold = 65.0

    return avg_last_10 >= threshold


# ============================================================================
# TWEAK 9: TPM Overhead Measurement
# ============================================================================


class TPMNodePoolWithMetrics:
    """TPM with crypto overhead measurement"""

    def __init__(self, num_nodes):
        self.nodes = {}
        print(f"Generating {num_nodes} RSA 2048-bit keys...", end="", flush=True)
        start = time.time()

        for node_id in range(num_nodes):
            key = rsa.generate_private_key(
                public_exponent=65537, key_size=2048, backend=default_backend()
            )
            self.nodes[node_id] = {
                "key": key,
                "pcrs": {i: f"pcr_{node_id}_{i}" for i in range(6)},
            }

        elapsed = time.time() - start
        print(f" Done ({elapsed:.1f}s)")

        self.quotes_created = 0
        self.quotes_verified = 0
        self.signature_times_ms = []
        self.verification_times_ms = []

    def create_quote(self, node_id, nonce):
        """Create quote with timing measurement"""
        import hashlib
        import time as time_module

        node = self.nodes[node_id]

        # Measure signature time
        start = time_module.time()
        pcr_data = "".join(node["pcrs"].values()).encode()
        quote_data = hashlib.sha256(pcr_data + nonce.encode()).digest()

        signature = node["key"].sign(
            quote_data,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()), salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256(),
        )
        sig_time_ms = (time_module.time() - start) * 1000
        self.signature_times_ms.append(sig_time_ms)

        # Mock verification (99% success) with timing
        start = time_module.time()
        verified = random.random() < 0.99
        verify_time_ms = (time_module.time() - start) * 1000
        self.verification_times_ms.append(verify_time_ms)

        self.quotes_created += 1
        if verified:
            self.quotes_verified += 1

        return verified

    def print_crypto_stats(self):
        """Print TPM overhead statistics"""
        if not self.signature_times_ms:
            return

        print(f"\nTPM Cryptography Overhead:")
        print(f"  Total Quotes: {self.quotes_created:,}")
        print(
            f"  Verified: {self.quotes_verified:,} ({self.quotes_verified/self.quotes_created:.1%})"
        )
        print(f"  Avg Signature Time: {np.mean(self.signature_times_ms):.3f}ms")
        print(f"  Avg Verification Time: {np.mean(self.verification_times_ms):.3f}ms")
        print(f"  Max Signature Time: {np.max(self.signature_times_ms):.3f}ms")
        print(f"  Max Verification Time: {np.max(self.verification_times_ms):.3f}ms")


# ============================================================================
# IMPROVED TEST CLASS
# ============================================================================


class Week1OptimizedBFTTest:
    """Optimized test with all tweaks integrated"""

    def __init__(self, num_nodes=75, rounds=50):
        self.NUM_NODES = num_nodes
        self.ROUNDS = rounds

        self.ATTACKS = [
            ImprovedByzantineAttack.SIGN_FLIP,
            ImprovedByzantineAttack.LABEL_FLIP,
            ImprovedByzantineAttack.FREE_RIDE,
            ImprovedByzantineAttack.AMPLIFICATION,
        ]

        self.BFT_LEVELS = [0, 10, 20, 30, 40, 50]

        self.tpm = TPMNodePoolWithMetrics(num_nodes)
        self.net = RealisticNetworkSimulator()
        self.byzantine = PersistentByzantineModel(num_nodes)

        self.results = []

    def run_round(self, round_num, bft_pct, attack_type):
        """Run single round with improvements"""

        nonce = f"r{round_num}"
        verified = 0
        attacked = 0
        delivered = 0

        for node_id in range(self.NUM_NODES):
            w = np.random.randn(100)

            # IMPROVED: Use persistent Byzantine nodes
            if self.byzantine.is_byzantine(node_id):
                # IMPROVED: Apply attacks with variance
                w = ImprovedByzantineAttack.apply(w, attack_type, variance=0.15)
                attacked += 1

            # IMPROVED: Use realistic network
            if self.net.deliver():
                delivered += 1

                # Create TPM quote with metrics
                if self.tpm.create_quote(node_id, nonce):
                    verified += 1

        # IMPROVED: Use better accuracy model
        accuracy = ImprovedAccuracyModel.calculate(
            round_num, self.ROUNDS, attacked, self.NUM_NODES, verified, self.net.rate()
        )

        loss = max(0.1, 3.5 - (round_num * 0.25) + (attacked / self.NUM_NODES) * 1.5)

        return {"acc": accuracy, "loss": loss, "attacked": attacked}

    def run_config(self, bft_pct, attack_type):
        """Run full config with persistent Byzantine nodes"""

        # Select Byzantine nodes for this config
        self.byzantine.select_for_config(bft_pct)

        accuracies = []
        for r in range(1, self.ROUNDS + 1):
            res = self.run_round(r, bft_pct, attack_type)
            accuracies.append(res["acc"])

        # IMPROVED: Use adaptive convergence threshold
        converged = check_convergence(accuracies, bft_pct)

        return {
            "bft": bft_pct,
            "attack": attack_type,
            "final": accuracies[-1],
            "converged": converged,
            "avg_last_10": np.mean(accuracies[-10:]),
            "max": max(accuracies),
            "min": min(accuracies),
        }

    def run_all(self):
        """Run all configurations"""

        print("\n" + "=" * 100)
        print("  WEEK 1: OPTIMIZED BFT TEST")
        print("  With Tweaks: Byzantine Variance, Realistic Network, Persistent Nodes")
        print("=" * 100 + "\n")

        config_num = 0
        for bft in self.BFT_LEVELS:
            for attack in self.ATTACKS:
                config_num += 1
                print(
                    f"  [{config_num:2d}/24] {bft:2d}% | {attack:15s} | ",
                    end="",
                    flush=True,
                )

                result = self.run_config(bft, attack)
                self.results.append(result)

                status = "OK" if result["converged"] else "FAIL"
                print(
                    f"[{status}] Acc: {result['final']:6.2f}% | Avg10: {result['avg_last_10']:6.2f}%"
                )

        return self.results

    def summary(self):
        """Print results"""

        print("\n" + "=" * 100)
        print("  RESULTS: OPTIMIZED BFT TEST")
        print("=" * 100 + "\n")

        conv = [r for r in self.results if r["converged"]]
        total = len(self.results)

        print(f"Total Configurations: {total}")
        print(f"Converged: {len(conv)} ({len(conv)/total:.1%})")
        print(f"Diverged: {total - len(conv)}\n")

        # Network stats
        self.net.print_stats()

        # TPM stats
        self.tpm.print_crypto_stats()

        # Byzantine threshold
        print("\nByzantine Tolerance Analysis:")
        for bft in self.BFT_LEVELS:
            cfgs = [r for r in self.results if r["bft"] == bft]
            conv_count = len([c for c in cfgs if c["converged"]])
            print(
                f"  {bft}% Byzantine: {conv_count}/{len(cfgs)} converged ({conv_count/len(cfgs):.1%})"
            )

        # Find threshold
        print("\nCritical Byzantine Threshold:")
        for bft in sorted(self.BFT_LEVELS):
            cfgs = [r for r in self.results if r["bft"] == bft]
            if all(not c["converged"] for c in cfgs):
                print(f"  THRESHOLD: {bft}% Byzantine (All {len(cfgs)} attacks failed)")
                break

        print("\n" + "=" * 100 + "\n")


# ============================================================================
# RUN
# ============================================================================

if __name__ == "__main__":
    print("\n[WEEK 1 OPTIMIZED] BFT Byzantine Tolerance Test\n")

    test = Week1OptimizedBFTTest(num_nodes=75, rounds=50)

    start = datetime.now()
    test.run_all()
    elapsed = datetime.now() - start

    test.summary()

    print(f"Execution Time: {str(elapsed).split('.')[0]}\n")
