import subprocess
import sys
import time


def verify():
    print("--- Starting CI Consensus Verification ---")

    # Check 10 times, waiting 20s between checks (Total ~3.3 minutes)
    for attempt in range(1, 11):
        print(f"Attempt {attempt}/10: Checking aggregator logs...")
        try:
            # We look for aggregator-ci now to match the new compose file
            logs = subprocess.check_output(
                ["docker", "logs", "--tail", "50", "aggregator-ci"],
                stderr=subprocess.STDOUT,
            ).decode("utf-8")

            if "Global model updated" in logs or "Round 1 complete" in logs:
                print("✅ SUCCESS: BFT Consensus reached!")
                return True

        except Exception as e:
            print(f"⚠️ Container not ready yet: {e}")

        print("Waiting 20 seconds...")
        time.sleep(20)

    return False


if __name__ == "__main__":
    if verify():
        sys.exit(0)
    else:
        print("❌ TIMEOUT: Consensus not reached.")
        sys.exit(1)
