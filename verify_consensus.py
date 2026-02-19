import subprocess
import sys
import time

def verify():
    print("--- Starting Consensus Verification ---")
    
    # We will check 10 times, waiting 20 seconds between each check (total ~3 mins)
    for attempt in range(1, 11):
        print(f"Check {attempt}/10...")
        try:
            logs = subprocess.check_output(
                ["docker", "logs", "aggregator-200"], 
                stderr=subprocess.STDOUT
            ).decode('utf-8')
            
            if "Global model updated" in logs or "Round 1 complete" in logs:
                print("✅ SUCCESS: BFT Consensus reached!")
                return True
        except Exception as e:
            print(f"Error reading logs: {e}")
            
        time.sleep(20) 
    
    print("❌ TIMEOUT: Consensus was not reached within 3 minutes.")
    return False

if __name__ == "__main__":
    if not verify():
        sys.exit(1)
