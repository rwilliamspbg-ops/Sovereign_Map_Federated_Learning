import re

with open("sovereignmap_production_backend_v2.py", "r") as f:
    content = f.read()

phase3d_endpoints = """
# Phase 3D Training Mock Endpoints
@app.route("/training/start", methods=["POST"])
def start_training():
    return jsonify({"status": "training", "message": "Phase 3D hardware training started via HUD"}), 200

@app.route("/training/stop", methods=["POST"])
def stop_training():
    return jsonify({"status": "idle", "message": "Training halted"}), 200

@app.route("/training/status", methods=["GET"])
def training_status():
    return jsonify({
        "status": "idle", 
        "round": strategy.round_num if strategy else 0,
        "total_rounds": 50,
        "current_metrics": {
            "accuracy": strategy.convergence_history["accuracies"][-1] if strategy and strategy.convergence_history["accuracies"] else 0.5,
            "loss": 0.5,
            "latency_ms": 125,
            "bandwidth_kb": 25.4,
            "compression_ratio": 4.1
        }
    }), 200

"""

# Insert endpoints
content = content.replace(
    '@app.route("/health", methods=["GET"])',
    phase3d_endpoints + '\n@app.route("/health", methods=["GET"])',
)

with open("sovereignmap_production_backend_v2.py", "w") as f:
    f.write(content)
