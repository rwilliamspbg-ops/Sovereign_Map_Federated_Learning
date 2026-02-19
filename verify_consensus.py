import subprocess
import sys
import time

def get_logs(container):
    try:
        return subprocess.check_output(
            ["docker", "logs", "--tail", "20", container],
            stderr=subprocess.STDOUT
        ).decode("utf-8")
    except:
        return "Could not retrieve logs."

def verify():
    for attempt in range(1, 11):
        print(f"Attempt {attempt}/10: Checking aggregator...")
        agg_logs = get_logs("aggregator-ci")
        
        if "Global model updated" in agg_logs:
            print("✅ SUCCESS: Consensus reached!")
            return True
        
        time.sleep(20)

    print("❌ TIMEOUT: Printing diagnostics...")
    print(f"--- Aggregator Last Logs ---\n{agg_logs}")
    print(f"--- Node-1 Last Logs ---\n{get_logs('sovereign_map_federated_learning-node-1-1')}")
    return False

if __name__ == "__main__":
    sys.exit(0 if verify() else 1)
