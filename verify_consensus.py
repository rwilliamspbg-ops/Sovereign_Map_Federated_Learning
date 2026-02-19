import subprocess
import sys
import time

def verify():
    print("--- Starting Consensus Verification ---")
    # Give it a moment to ensure logs are flushed
    time.sleep(5)
    
    try:
        # Check aggregator logs for the BFT success signal
        logs = subprocess.check_output(
            ["docker", "logs", "aggregator-200"], 
            stderr=subprocess.STDOUT
        ).decode('utf-8')
        
        if "Global model updated" in logs or "Round 1 complete" in logs:
            print("✅ SUCCESS: BFT Consensus reached!")
            return True
        else:
            print("⚠️ PENDING: Aggregator is still processing...")
            return False
            
    except Exception as e:
        print(f"❌ Error accessing logs: {e}")
        return False

if __name__ == "__main__":
    success = verify()
    if not success:
        sys.exit(1)
