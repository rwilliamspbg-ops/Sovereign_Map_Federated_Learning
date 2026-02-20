import json
import numpy as np
import os


def monitor_live_convergence(file_path="./Live_100_Node/live_stats.json"):
    # Check if the file exists before attempting to open it
    if not os.path.exists(file_path):
        # Return a status that indicates we are waiting for AWS to start
        return "WAITING_FOR_AWS_DATA"

    try:
        with open(file_path, "r") as f:
            data = json.load(f)

        # Get the last 5 accuracy scores
        if "history" not in data or len(data["history"]) < 5:
            return "CONVERGING"

        recent_acc = [r["accuracy"] for r in data["history"]][-5:]

        # Calculate variance (Standard Deviation squared)
        std_dev = np.std(recent_acc)

        # If variance is < 0.001 for 5 rounds, we have plateaued
        if std_dev < 0.001:
            return "PLATEAU_REACHED"
        return "CONVERGING"

    except (json.JSONDecodeError, KeyError):
        return "DATA_CORRUPT_OR_EMPTY"


if __name__ == "__main__":
    status = monitor_live_convergence()
    print(status)
