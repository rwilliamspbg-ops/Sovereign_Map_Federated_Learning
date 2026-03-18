#!/usr/bin/env python3
"""
BFT Stress Test Execution Script
Manages the 12-configuration test across 75 nodes with 200+ rounds each
"""

import json
import requests
import time
import sys
from datetime import datetime, timedelta
from typing import List, Dict
import subprocess


class BFTTestExecutor:
    def __init__(self, backend_url: str = "http://localhost:8081"):
        self.backend_url = backend_url
        self.results = {}
        self.start_time = None

    def check_backend_health(self) -> bool:
        """Verify backend is ready"""
        try:
            resp = requests.get(f"{self.backend_url}/health", timeout=5)
            return resp.status_code == 200
        except:
            return False

    def wait_for_backend(self, max_wait: int = 60) -> bool:
        """Wait for backend to become ready"""
        start = time.time()
        while time.time() - start < max_wait:
            if self.check_backend_health():
                print(f"✅ Backend healthy at {self.backend_url}")
                return True
            print(f"⏳ Waiting for backend... ({int(time.time()-start)}s)")
            time.sleep(5)
        return False

    def start_test(self) -> bool:
        """Start BFT comprehensive test"""
        try:
            print(f"\n{'='*80}")
            print(f"  🚀 STARTING BFT STRESS TEST - 75 NODES | 200 ROUNDS EACH")
            print(f"{'='*80}\n")

            self.start_time = datetime.now()

            # POST to start test
            print("📤 Sending test start request...")
            resp = requests.post(f"{self.backend_url}/start_bft_test", timeout=300)

            if resp.status_code != 200:
                print(f"❌ Failed to start test: {resp.status_code}")
                return False

            data = resp.json()
            print(f"✅ Test started successfully")
            print(f"   Tests Queued: {data.get('tests_run', 'N/A')}")

            return True
        except Exception as e:
            print(f"❌ Error: {e}")
            return False

    def monitor_progress(self, check_interval: int = 10, max_wait: int = 3600):
        """Monitor test progress"""
        print(f"\n📊 MONITORING TEST PROGRESS (updating every {check_interval}s)\n")

        last_results = None
        stalled_count = 0

        start = time.time()
        while time.time() - start < max_wait:
            try:
                resp = requests.get(f"{self.backend_url}/bft_results", timeout=10)
                if resp.status_code != 200:
                    print(f"⚠️  Failed to fetch results: {resp.status_code}")
                    time.sleep(check_interval)
                    continue

                current_results = resp.json()
                num_tests = len(current_results.get("tests", []))

                elapsed = datetime.now() - self.start_time
                elapsed_min = int(elapsed.total_seconds() / 60)

                # Check for progress
                if current_results == last_results:
                    stalled_count += 1
                else:
                    stalled_count = 0

                # Display status
                print(f"[{elapsed_min:03d}m] Tests Completed: {num_tests}/12", end="")

                if num_tests > 0:
                    last_test = current_results["tests"][-1]
                    print(
                        f" | Latest: {last_test['byzantine_percentage']}% Byzantine "
                        f"({last_test['aggregation_method']}) | "
                        f"Final Acc: {last_test['final_accuracy']:.1f}% | "
                        f"Converged: {'✅' if last_test['converged'] else '❌'}",
                        end="",
                    )

                print()

                last_results = current_results

                # Check if all done
                if num_tests >= 12:
                    print(f"\n✅ ALL TESTS COMPLETED!")
                    return current_results

                # Check for stall
                if stalled_count > 10:
                    print(
                        f"⚠️  No progress for {stalled_count * check_interval}s, checking backend..."
                    )
                    if not self.check_backend_health():
                        print(f"❌ Backend unhealthy!")
                        return None

                time.sleep(check_interval)

            except Exception as e:
                print(f"⚠️  Error fetching progress: {e}")
                time.sleep(check_interval)

        print(f"⏱️  Timeout reached ({max_wait}s)")
        return current_results

    def get_summary(self) -> Dict:
        """Get test summary"""
        try:
            resp = requests.get(f"{self.backend_url}/bft_summary", timeout=10)
            if resp.status_code == 200:
                return resp.json()
        except:
            pass
        return {}

    def generate_report(self, results: Dict, summary: Dict) -> str:
        """Generate markdown report"""

        report = f"""# 🔐 Byzantine Fault Tolerance Stress Test Report

**Generated:** {datetime.now().isoformat()}  
**Node Count:** 75 agents + 1 aggregator  
**Rounds per Configuration:** 200  
**Total Configurations Tested:** 12 (6 Byzantine % × 2 aggregation methods)  

---

## Executive Summary

"""

        if summary:
            critical = summary.get("critical_threshold")
            if critical:
                report += (
                    f"**Critical Threshold Found:** {critical}% Byzantine nodes\n\n"
                )
                report += (
                    f"System convergence **FAILS** above this Byzantine percentage.\n\n"
                )

        report += "## Test Results by Configuration\n\n"

        if results.get("tests"):
            for test in results["tests"]:
                report += f"### {test['byzantine_percentage']}% Byzantine + {test['aggregation_method'].title()}\n\n"
                report += f"- **Nodes:** {test['num_nodes']} total, {test['num_byzantine']} Byzantine\n"
                report += f"- **Rounds Completed:** {test['rounds_completed']}\n"
                report += f"- **Final Accuracy:** {test['final_accuracy']:.2f}%\n"
                report += f"- **Avg Accuracy (Last 10):** {test['avg_accuracy_last_10']:.2f}%\n"
                report += (
                    f"- **Converged:** {'✅ Yes' if test['converged'] else '❌ No'}\n"
                )
                report += f"- **Stable:** {'✅ Yes' if test['stable'] else '⚠️  No'}\n"
                report += f"- **Max Accuracy Reached:** {test['max_accuracy']:.2f}%\n"
                report += (
                    f"- **Accuracy Variance:** {test['accuracy_variance']:.4f}\n\n"
                )

        return report

    def run_full_test(self):
        """Execute complete test sequence"""

        # Check backend
        if not self.wait_for_backend():
            print(f"❌ Backend unreachable at {self.backend_url}")
            return False

        # Start test
        if not self.start_test():
            print(f"❌ Failed to start test")
            return False

        # Monitor
        results = self.monitor_progress(check_interval=15, max_wait=3600)

        if not results:
            print(f"❌ Test monitoring failed")
            return False

        # Get summary
        summary = self.get_summary()

        # Generate report
        report = self.generate_report(results, summary)

        elapsed = datetime.now() - self.start_time
        print(f"\n{'='*80}")
        print(f"  ✅ BFT STRESS TEST COMPLETED")
        print(f"{'='*80}")
        print(f"\nTotal Duration: {str(elapsed).split('.')[0]}")
        print(f"Tests Run: {len(results.get('tests', []))}")
        print(
            f"\nCritical Threshold: {summary.get('critical_threshold', 'Not found')}%\n"
        )

        # Save report
        with open("BFT_TEST_RESULTS.md", "w") as f:
            f.write(report)

        print(f"✅ Report saved to BFT_TEST_RESULTS.md")

        return True


if __name__ == "__main__":
    executor = BFTTestExecutor()
    success = executor.run_full_test()
    sys.exit(0 if success else 1)
