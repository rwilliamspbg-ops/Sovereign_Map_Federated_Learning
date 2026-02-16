import json

# Simulation data from your Round 41 run
data = {
    "rounds": list(range(1, 26)),
    "accuracy": [85, 88, 90, 92, 94, 95, 96, 96.5, 96.9] + [88.2]*1 + [92.4, 93.1, 94.7, 95.8, 96.2, 96.9],
    "event": "Byzantine Breach (55.6%) at Round 10"
}

with open('plot_data.json', 'w') as f:
    json.dump(data, f)
print("Data summary generated for local plotting.")
