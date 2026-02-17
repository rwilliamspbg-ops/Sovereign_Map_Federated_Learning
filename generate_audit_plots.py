import json
import matplotlib.pyplot as plt
import glob
import os

def create_plots():
    # Find the most recent result file
    files = glob.glob("audit_results/*.json")
    if not files:
        print("No JSON results found in audit_results/")
        return
    
    latest_file = max(files, key=os.path.getctime)
    with open(latest_file, 'r') as f:
        data = json.load(f)

    # Extract recovery data
    rounds = list(range(1, len(data['recovery_accuracy_values']) + 1))
    accuracy = [val * 100 for val in data['recovery_accuracy_values']]

    plt.figure(figsize=(10, 6))
    plt.plot(rounds, accuracy, marker='o', linestyle='-', color='#2ecc71', linewidth=2)
    
    # Formatting the plot for Sovereign Map Branding
    plt.title(f"Sovereign-Mohawk V0.3.0 Validation\nConvergence at {data['malicious_fraction']*100}% Malicious Load", fontsize=14)
    plt.xlabel("Communication Rounds (Post-Breach)", fontsize=12)
    plt.ylabel("Model Accuracy (%)", fontsize=12)
    plt.grid(True, linestyle='--', alpha=0.7)
    plt.ylim(0, 100)
    
    # Save the plot
    plot_path = "audit_results/convergence_audit_plot.png"
    plt.savefig(plot_path)
    print(f"✅ Plot generated and saved to: {plot_path}")

if __name__ == "__main__":
    create_plots()
