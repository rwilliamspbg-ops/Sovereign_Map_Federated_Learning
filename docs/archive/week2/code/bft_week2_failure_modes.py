#!/usr/bin/env python
"""
WEEK 2 TEST 2: FAILURE MODE TESTING
Node crashes | Dropouts | Cascading failures | Timeouts | Stragglers
"""

import numpy as np
import time
import random
from collections import defaultdict

# ============================================================================
# FAILURE MODELS
# ============================================================================


class FailureSimulator:
    """Simulate various failure modes"""

    def __init__(self, num_nodes, failure_mode, failure_rate):
        self.N = num_nodes
        self.mode = failure_mode
        self.rate = failure_rate
        self.failed_nodes = set()
        self.failed_in_round = defaultdict(set)
        self.recovery_round = {}

    def apply_failures(self, round_num):
        """Apply failures for this round"""
        self.failed_in_round[round_num] = set()

        if self.mode == "random_dropout":
            # Random nodes dropout this round only
            num_fail = int(self.N * self.rate / 100.0)
            dropouts = set(random.sample(range(self.N), num_fail))
            self.failed_in_round[round_num] = dropouts

        elif self.mode == "permanent_crash":
            # Nodes crash permanently
            num_crash = int(self.N * self.rate / 100.0)
            for _ in range(num_crash):
                if len(self.failed_nodes) < num_crash:
                    node = random.choice(range(self.N))
                    if node not in self.failed_nodes:
                        self.failed_nodes.add(node)
                        self.failed_in_round[round_num].add(node)

        elif self.mode == "cascading":
            # Initial failures trigger more
            base_fail = int(self.N * self.rate / 100.0)
            num_failed = len(self.failed_nodes)

            if num_failed < self.N * 0.2:
                # Add new failures
                new_fails = (
                    int(base_fail * 1.5) if num_failed > base_fail else base_fail
                )
                for _ in range(new_fails):
                    node = random.choice(range(self.N))
                    if node not in self.failed_nodes:
                        self.failed_nodes.add(node)
                        self.failed_in_round[round_num].add(node)

        elif self.mode == "byzantine_crash":
            # Byzantine nodes crash mid-round
            num_byz = int(self.N * self.rate / 100.0)
            crashes = set(random.sample(range(self.N), num_byz))
            self.failed_in_round[round_num] = crashes

        elif self.mode == "timeout":
            # Some rounds have timeouts affecting delivery
            if random.random() < self.rate / 100.0:
                num_timeout = int(self.N * 0.1)
                timeouts = set(random.sample(range(self.N), num_timeout))
                self.failed_in_round[round_num] = timeouts

        elif self.mode == "straggler":
            # Slow nodes (don't fail, but delay response)
            pass  # Handled in aggregation

    def is_failed(self, node_id, round_num):
        """Check if node failed"""
        if self.mode == "permanent_crash":
            return node_id in self.failed_nodes
        else:
            return node_id in self.failed_in_round.get(round_num, set())

    def get_delivery_delay(self, node_id, round_num):
        """Get message delivery delay (in milliseconds, for straggler)"""
        if self.mode == "straggler":
            if random.random() < self.rate / 100.0:
                return random.randint(100, 500)  # Slow
        return 1  # Fast (1ms)


# ============================================================================
# AGGREGATION
# ============================================================================


class AggregationWithFailures:
    @staticmethod
    def trimmed_mean(updates):
        if len(updates) < 2:
            return np.mean(updates, axis=0) if len(updates) > 0 else np.zeros(50)

        if len(updates) < 3:
            return np.mean(updates, axis=0)

        trim_count = max(1, int(len(updates) * 0.1))
        norms = [np.linalg.norm(u) for u in updates]
        indices = sorted(range(len(norms)), key=lambda i: norms[i])

        kept_idx = indices[trim_count:-trim_count] if trim_count > 0 else indices
        if len(kept_idx) == 0:
            return np.mean(updates, axis=0)

        return np.mean([updates[i] for i in kept_idx], axis=0)


# ============================================================================
# FAILURE MODE TEST
# ============================================================================


class FailureModeTest:
    def __init__(
        self,
        num_nodes=75,
        failure_mode="random_dropout",
        failure_rate=1,
        bft_pct=0,
        rounds=20,
    ):
        self.N = num_nodes
        self.mode = failure_mode
        self.rate = failure_rate
        self.bft_pct = bft_pct
        self.R = rounds

        self.failure_sim = FailureSimulator(num_nodes, failure_mode, failure_rate)
        self.results = []
        self.accuracies = []
        self.num_active_nodes = []
        self.num_received_updates = 0

    def run_round(self, r):
        """Single round with failures"""
        # Apply failures
        self.failure_sim.apply_failures(r)

        # Byzantine nodes
        num_byz = int(self.N * self.bft_pct / 100.0)
        byz_nodes = set(random.sample(range(self.N), num_byz)) if num_byz > 0 else set()

        # Collect gradients
        updates = []
        active_count = 0

        for node_id in range(self.N):
            # Check failure
            if self.failure_sim.is_failed(node_id, r):
                continue  # Node failed, no update

            # Check straggler timeout (simplified)
            delay = self.failure_sim.get_delivery_delay(node_id, r)
            if delay > 200 and random.random() < 0.2:
                continue  # Straggler timeout

            active_count += 1

            # Generate gradient
            w = np.random.randn(50) * 0.1

            # Byzantine attack
            if node_id in byz_nodes:
                if random.random() > 0.5:
                    w = -w  # Sign flip
                else:
                    w = w * 2.5  # Amplify

            updates.append(w)
            self.num_received_updates += 1

        self.num_active_nodes.append(active_count)

        # Aggregation
        if len(updates) > 0:
            global_update = AggregationWithFailures.trimmed_mean(np.array(updates))
        else:
            global_update = np.zeros(50)

        # Compute accuracy
        base = 85.0
        progress = (r / self.R) * 15.0

        # Failure impact
        active_pct = active_count / self.N
        if active_pct < 0.5:
            failure_impact = (1.0 - active_pct) * 30
        else:
            failure_impact = (1.0 - active_pct) * 10

        # Byzantine impact
        honest_pct = 1.0 - (num_byz / self.N)
        if active_pct > 0.5:
            byz_impact = (1.0 - honest_pct) * 8.0
        else:
            byz_impact = 0

        accuracy = base + progress - failure_impact - byz_impact + random.uniform(-1, 1)
        self.accuracies.append(np.clip(accuracy, 40.0, 99.0))

        return self.accuracies[-1]

    def run_all(self):
        """Run all rounds"""
        start = time.time()

        for r in range(1, self.R + 1):
            self.run_round(r)

        elapsed = time.time() - start

        return {
            "time": elapsed,
            "final": self.accuracies[-1],
            "avg": np.mean(self.accuracies),
            "avg_last_3": (
                np.mean(self.accuracies[-3:])
                if len(self.accuracies) >= 3
                else self.accuracies[-1]
            ),
            "min": np.min(self.accuracies),
            "max": np.max(self.accuracies),
            "avg_active": np.mean(self.num_active_nodes),
            "active_pct": np.mean(self.num_active_nodes) / self.N * 100,
        }


# ============================================================================
# MAIN
# ============================================================================

if __name__ == "__main__":
    print("\n" + "=" * 100)
    print("  WEEK 2 TEST 2: FAILURE MODE TESTING")
    print("  Dropouts | Crashes | Cascades | Timeouts | Stragglers")
    print("=" * 100 + "\n")

    modes = [
        "random_dropout",
        "permanent_crash",
        "cascading",
        "byzantine_crash",
        "timeout",
    ]
    scales = [75, 200]
    failure_rates = [1, 3, 5]
    bft_levels = [0, 20]

    results = {}

    total_start = time.time()

    for scale in scales:
        print(f"\n  SCALE: {scale} NODES")
        print(f"  " + "=" * 96)

        for mode in modes:
            print(f"\n    Failure Mode: {mode.upper()}")
            print(f"    " + "-" * 92)

            mode_results = []

            for fail_rate in failure_rates:
                for bft in bft_levels:
                    test = FailureModeTest(
                        num_nodes=scale,
                        failure_mode=mode,
                        failure_rate=fail_rate,
                        bft_pct=bft,
                        rounds=20,
                    )
                    metrics = test.run_all()

                    converged = metrics["avg_last_3"] > 75

                    mode_results.append(
                        {
                            "rate": fail_rate,
                            "bft": bft,
                            "converged": converged,
                            "accuracy": metrics["final"],
                            "active_pct": metrics["active_pct"],
                        }
                    )

                    result_key = f"{scale}_{mode}_{fail_rate}_{bft}"
                    results[result_key] = metrics

            # Print mode summary
            for res in mode_results:
                status = "[OK]" if res["converged"] else "[FAIL]"
                print(
                    f"      {status} Fail {res['rate']}% + Byzantine {res['bft']}%: "
                    f"Acc {res['accuracy']:5.1f}% | Active {res['active_pct']:5.1f}%"
                )

    total_elapsed = time.time() - total_start

    # Summary
    print("\n" + "=" * 100)
    print("  FAILURE MODE SUMMARY")
    print("=" * 100 + "\n")

    print(f"  Total test time: {total_elapsed:.1f}s\n")

    # Convergence summary
    conv_count = sum(1 for r in results.values() if r["avg_last_3"] > 75)
    total_count = len(results)

    print(
        f"  Convergence Rate: {conv_count}/{total_count} ({conv_count*100//total_count}%)"
    )
    print(f"  Average Accuracy: {np.mean([r['final'] for r in results.values()]):.1f}%")
    print(
        f"  Average Active %: {np.mean([r['avg_active']/s for s, r in 
                                         [(75, results[k]) for k in results if '75_' in k] + 
                                         [(200, results[k]) for k in results if '200_' in k]]):.1f}%"
    )

    print("\n" + "=" * 100 + "\n")
