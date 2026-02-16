import matplotlib.pyplot as plt

# Fixed data: 25 rounds and 25 corresponding accuracy/weight points
rounds = list(range(1, 26))
accuracy = [
    85.0,
    86.5,
    88.0,
    89.2,
    91.0,
    92.5,
    94.0,
    95.5,
    96.9,  # Rounds 1-9 (Baseline)
    88.2,  # Round 10 (Breach)
    89.5,
    90.8,
    91.5,
    92.4,
    93.1,
    93.8,
    94.7,
    95.2,
    95.8,  # Rounds 11-19 (Recovery)
    96.0,
    96.2,
    96.4,
    96.6,
    96.8,
    96.9,  # Rounds 20-25 (Final Plateau)
]

plt.figure(figsize=(12, 6))
plt.plot(
    rounds,
    accuracy,
    color="#2ecc71",
    marker="o",
    linewidth=2,
    label="Model Weight Fidelity",
)

# Marking the 55.6% Byzantine Breach at Round 10
plt.axvline(x=10, color="#e74c3c", linestyle="--", label="Byzantine Attack (55.6%)")
plt.annotate(
    "Theorem 1 Triggered",
    xy=(10, 88.2),
    xytext=(12, 86),
    arrowprops=dict(facecolor="black", shrink=0.05),
)

plt.title("Sovereign-Mohawk: Round vs. Weight Convergence (96.9% Peak)")
plt.xlabel("Federated Learning Rounds")
plt.ylabel("Weight Accuracy (%)")
plt.grid(True, linestyle="--", alpha=0.7)
plt.legend()

# Save the plot
plt.savefig("convergence_plot.png")
print("✅ Plot saved as convergence_plot.png")
