import argparse
import subprocess
import sys
import time

def run_test(malicious_fraction):
    print(f"\n🚀 Starting Sovereign-Mohawk Stress Test...")
    print(f"📊 Target Malicious Load: {malicious_fraction * 100}%")
    
    # Update the collector script with the new input
    sed_cmd = f"sed -i 's/malicious_fraction = .*/malicious_fraction = {malicious_fraction}/' sovereign_map_test_collector.py"
    subprocess.run(sed_cmd, shell=True)
    
    # Run the simulation
    print("⏳ Processing 10M-node hierarchy updates...")
    result = subprocess.run(["python3", "sovereign_map_test_collector.py"], capture_output=True, text=True)
    
    # Live Visual Feedback
    if "bft_safe: True" in result.stdout:
        print("✅ STATUS: [SAFE] - Hierarchical Multi-Krum successfully filtered updates.")
    else:
        print("⚠️ STATUS: [BREACHED] - Security Threshold Exceeded. Island Mode active.")
    
    print("-" * 50)
    print("Top 5 Convergence Rounds (Post-Breach Recovery):")
    # Grabbing the recovery rounds from the generated report
    subprocess.run("grep -A 5 'recovery_accuracy_values' sovereign_test_report_*.txt", shell=True)
    print("-" * 50)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Sovereign Map Live Test Controller')
    parser.add_argument('--malicious', type=float, default=0.556, help='Fraction of malicious nodes (0.0 to 1.0)')
    args = parser.parse_args()
    
    run_test(args.malicious)
