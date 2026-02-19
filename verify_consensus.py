import subprocess
import time
import sys

def check_logs():
    print("Searching logs for Consensus Success...")
    # This command looks at the aggregator logs for the success signal
    cmd = "docker logs aggregator-200"
    result = subprocess.check_output(cmd, shell=True).decode('utf-8')
    
    if "Global model updated" in result or "Round 1 complete" in result:
        print("✅ SUCCESS: Consensus reached and model updated!")
        return True
    return False

# Give the cluster time to run a round
time.sleep(30) 
if not check_logs():
    sys.exit(1) # This fails the GitHub Action if consensus isn't found
