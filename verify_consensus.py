import subprocess
import sys
import time


def get_logs(container):
    try:
        return subprocess.check_output(
            ["docker", "logs", "--tail", "50", container], stderr=subprocess.STDOUT
        ).decode("utf-8")
    except Exception:
        return "Could not retrieve logs."


def verify():
    print("--- Starting CI Consensus Verification ---")

    # Check 10 times, waiting 20s between checks
    for attempt in range(1, 11):
        print(f"Attempt {attempt}/10: Checking aggregator logs...")
        logs = get_logs("aggregator-ci")

        # Matches the actual logs: {"message": "FL round X completed", ...}
        if "round" in logs.lower() and "completed" in logs.lower():
            print("✅ SUCCESS: Federated Learning rounds are completing!")
            return True

        print("Waiting 20 seconds...")
        time.sleep(20)

    print("❌ TIMEOUT: Printing diagnostics...")
    print(f"--- Aggregator Last Logs ---\n{get_logs('aggregator-ci')}")
    return False


if __name__ == "__main__":
    # Exit with 0 on success, 1 on failure
    sys.exit(0 if verify() else 1)
