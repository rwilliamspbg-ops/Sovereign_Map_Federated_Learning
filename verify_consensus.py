import subprocess
import sys
import time


def verify():
    print("--- Starting Robust Consensus Verification ---")

    # Check 10 times, waiting 20s between checks (Total ~3.3 minutes)
    for attempt in range(1, 11):
        print(f"Attempt {attempt}/10: Checking aggregator logs...")
        try:
            # Get the last 50 lines of the aggregator logs
            logs = subprocess.check_output(
                ["docker", "logs", "--tail", "50", "aggregator-200"],
                stderr=subprocess.STDOUT,
            ).decode("utf-8")

            # Look for your specific success indicators
            if "Global model updated" in logs or "Round 1 complete" in logs:
                print("✅ SUCCESS: BFT Consensus reached and model updated!")
                return True

        except Exception as e:
            print(f"⚠️ Container not ready yet: {e}")

        print("Waiting 20 seconds for next check...")
        time.sleep(20)

    return False


if __name__ == "__main__":
    if verify():
        sys.exit(0)  # Pass
    else:
        print("❌ TIMEOUT: Consensus was not reached within the time limit.")
        sys.exit(1)  # Fail the GitHub Action
