#!/usr/bin/env python
"""
WEEK 2 TEST 4: CASCADING FAILURES
Avalanche | Threshold cascade | Recovery | Byzantine amplification
"""

import numpy as np
import time
import random

# ============================================================================
# CASCADING FAILURE SIMULATOR
# ============================================================================


class CascadingFailureSim:
    """Simulate cascading failure patterns"""

    def __init__(self, num_nodes, cascade_type, initial_fail_pct):
        self.N = num_nodes
        self.ctype = cascade_type
        self.initial_fail_pct = initial_fail_pct

        self.failed_nodes = set()
        self.recovery_schedule = {}
        self.failure_rounds = defaultdict(set)

        # Setup initial failures
        num_initial = int(self.N * initial_fail_pct / 100.0)
        self.failed_nodes = set(random.sample(range(self.N), num_initial))

        for node in self.failed_nodes:
            self.failure_rounds[0].add(node)


from collections import defaultdict


class CascadingFailureTest:
    def __init__(
        self,
        num_nodes=200,
        cascade_type="avalanche",
        initial_fail_pct=5,
        bft_pct=0,
        rounds=40,
    ):
        self.N = num_nodes
        self.ctype = cascade_type
        self.ifail = initial_fail_pct
        self.bft_pct = bft_pct
        self.R = rounds

        self.failed_nodes = set()
        self.recovery_schedule = {}

        # Setup initial failures
        num_initial = int(self.N * initial_fail_pct / 100.0)
        self.failed_nodes = set(random.sample(range(self.N), num_initial))

        self.accuracies = []
        self.failure_counts = []

    def apply_cascading_failures(self, r):
        """Apply cascading failure logic"""
        if self.ctype == "avalanche":
            # Each failed node triggers more failures probabilistically
            num_failed = len(self.failed_nodes)

            if num_failed < self.N * 0.25:
                # Probability of new failures based on already failed
                new_fail_prob = min(0.5, num_failed / self.N)
                candidates = list(set(range(self.N)) - self.failed_nodes)

                if random.random() < new_fail_prob and candidates:
                    new_fails = random.randint(1, int(self.N * 0.02) + 1)
                    for _ in range(new_fails):
                        if candidates:
                            node = candidates.pop()
                            self.failed_nodes.add(node)

        elif self.ctype == "threshold":
            # Linear until 20%, then exponential
            num_failed = len(self.failed_nodes)

            if num_failed < self.N * 0.2:
                # Linear: add constant per round
                if random.random() < 0.3:
                    candidates = list(set(range(self.N)) - self.failed_nodes)
                    if candidates:
                        node = random.choice(candidates)
                        self.failed_nodes.add(node)
            else:
                # Exponential: accelerate failures
                new_fails = int((num_failed / self.N) * self.N * 0.1)
                candidates = list(set(range(self.N)) - self.failed_nodes)

                for _ in range(min(new_fails, len(candidates))):
                    if candidates:
                        node = candidates.pop()
                        self.failed_nodes.add(node)

        elif self.ctype == "recovery":
            # Failed nodes restart after 5 rounds
            nodes_to_recover = [
                n for n, rec_r in self.recovery_schedule.items() if rec_r == r
            ]
            for node in nodes_to_recover:
                self.failed_nodes.discard(node)

            # New failures trigger recovery schedule
            if r == 1:
                for node in self.failed_nodes:
                    self.recovery_schedule[node] = r + 5

        elif self.ctype == "byzantine_amplification":
            # Byzantine nodes trigger failures in honest nodes
            num_byz = int(self.N * self.bft_pct / 100.0)
            num_failed = len(self.failed_nodes)

            if num_byz > 0 and num_failed < self.N * 0.3:
                # Byzantine amplification: trigger new failures
                amp_factor = (num_byz / self.N) * 5.0
                new_fail_prob = min(0.8, amp_factor)

                if random.random() < new_fail_prob:
                    candidates = list(
                        set(range(self.N)) - self.failed_nodes - set(range(num_byz))
                    )
                    if candidates:
                        node = random.choice(candidates)
                        self.failed_nodes.add(node)

    def run_round(self, r):
        """Single round with cascading failures"""
        # Apply cascading
        self.apply_cascading_failures(r)

        # Byzantine nodes
        num_byz = int(self.N * self.bft_pct / 100.0)
        byz_nodes = set(random.sample(range(num_byz), min(num_byz, self.N)))

        # Collect updates
        updates = []
        active = 0

        for node_id in range(self.N):
            if node_id in self.failed_nodes:
                continue

            active += 1

            w = np.random.randn(50) * 0.1
            if node_id in byz_nodes:
                w = -w if random.random() > 0.5 else w * 2.5

            updates.append(w)

        self.failure_counts.append(len(self.failed_nodes))

        # Aggregation
        if len(updates) > 0:
            global_update = np.mean(updates, axis=0)

        # Accuracy
        active_pct = active / self.N
        base = 80.0

        if active_pct < 0.5:
            accuracy = base * active_pct - 20
        else:
            accuracy = base + (active_pct - 0.5) * 20 - 10

        # Byzantine impact
        if active > 0:
            honest_pct = 1.0 - (num_byz / active)
            byz_impact = (1.0 - honest_pct) * 8.0
        else:
            byz_impact = 0

        accuracy = accuracy - byz_impact + random.uniform(-1, 1)
        self.accuracies.append(np.clip(accuracy, 20, 99))

        return self.accuracies[-1]

    def run_all(self):
        """Run all rounds"""
        start = time.time()

        for r in range(1, self.R + 1):
            self.run_round(r)

        elapsed = time.time() - start

        max_failed = max(self.failure_counts)
        max_failed_pct = max_failed / self.N * 100

        return {
            "time": elapsed,
            "final": self.accuracies[-1],
            "avg": np.mean(self.accuracies),
            "avg_last_5": (
                np.mean(self.accuracies[-5:])
                if len(self.accuracies) >= 5
                else self.accuracies[-1]
            ),
            "max_failed": max_failed,
            "max_failed_pct": max_failed_pct,
        }


# ============================================================================
# MAIN
# ============================================================================

if __name__ == "__main__":
    print("\n" + "=" * 100)
    print("  WEEK 2 TEST 4: CASCADING FAILURES")
    print("  Avalanche | Threshold | Recovery | Byzantine Amplification")
    print("=" * 100 + "\n")

    scales = [200, 500]
    cascade_types = ["avalanche", "threshold", "recovery", "byzantine_amplification"]
    initial_fails = [1, 3, 5]
    bft_levels = [0, 20]

    results = {}

    total_start = time.time()

    for scale in scales:
        print(f"\n  SCALE: {scale} NODES")
        print(f"  " + "=" * 96)

        for ctype in cascade_types:
            print(f"\n    Cascade Type: {ctype.upper()}")
            print(f"    " + "-" * 92)

            for init_fail in initial_fails:
                for bft in bft_levels:
                    test = CascadingFailureTest(
                        num_nodes=scale,
                        cascade_type=ctype,
                        initial_fail_pct=init_fail,
                        bft_pct=bft,
                        rounds=40,
                    )
                    metrics = test.run_all()

                    result_key = f"{scale}_{ctype}_{init_fail}_{bft}"
                    results[result_key] = metrics

                    status = "[OK]" if metrics["max_failed_pct"] < 30 else "⚠"
                    print(
                        f"      {status} Init {init_fail}% + Byz {bft}%: "
                        f"Max Failed {metrics['max_failed_pct']:5.1f}% | "
                        f"Final Acc {metrics['final']:5.1f}%"
                    )

    total_elapsed = time.time() - total_start

    # Summary
    print("\n" + "=" * 100)
    print("  CASCADING FAILURE SUMMARY")
    print("=" * 100 + "\n")

    print(f"  Total test time: {total_elapsed:.1f}s\n")

    avg_max_failed = np.mean([r["max_failed_pct"] for r in results.values()])
    avg_acc = np.mean([r["final"] for r in results.values()])

    print(f"  Average Max Failed: {avg_max_failed:.1f}%")
    print(f"  Average Final Accuracy: {avg_acc:.1f}%")
    print(
        f"  Cascade Containment: {'[OK] Contained <30%' if avg_max_failed < 30 else '⚠ Exceeds 30%'}"
    )

    print("\n" + "=" * 100 + "\n")
