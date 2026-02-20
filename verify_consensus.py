import json
import numpy as np

def monitor_live_convergence(file_path="./Live_100_Node/live_stats.json"):
    with open(file_path, 'r') as f:
        data = json.load(f)
        # Get the last 5 accuracy scores
        recent_acc = [r['accuracy'] for r in data['history']][-5:]
    
    if len(recent_acc) < 5:
        return "CONVERGING"

    # Calculate variance (Standard Deviation squared)
    std_dev = np.std(recent_acc)
    
    # If variance is < 0.01 for 5 rounds, we have plateaued
    if std_dev < 0.001:
        return "PLATEAU_REACHED"
    return "CONVERGING"

if __name__ == "__main__":
    status = monitor_live_convergence()
    print(status)
