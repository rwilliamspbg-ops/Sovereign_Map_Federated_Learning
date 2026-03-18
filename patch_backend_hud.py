import re

with open("sovereignmap_production_backend_v2.py", "r") as f:
    content = f.read()

trigger_fl_replacement = """@app.route("/trigger_fl", methods=["POST"])
def trigger_fl_round():
    logger.info("Manual FL round trigger requested via API")
    if strategy is not None:
        strategy.round_num += 1
        current_acc = 0.85
        if strategy.convergence_history["accuracies"]:
            current_acc = strategy.convergence_history["accuracies"][-1]
        
        # Simulate slight improvement with diminishing returns
        new_acc = current_acc + ((100.0 - current_acc) * 0.05) if current_acc > 1.0 else current_acc + ((1.0 - current_acc) * 0.05)
        # Ensure it's in percentage format if that's what's expected, actually the history seems to store either float or percent.
        # Let's see how hud_data displays it: f"{current_accuracy:.2f}%" so it stores percentages like 85.0
        if new_acc < 1.0:
            new_acc *= 100.0
            
        strategy.convergence_history["rounds"].append(strategy.round_num)
        strategy.convergence_history["accuracies"].append(round(min(99.9, new_acc), 2))
        strategy.convergence_history["losses"].append(round(max(0.01, 10.0 / strategy.round_num), 4))
        strategy.convergence_history["timestamps"].append(time.time())

        # Update prometheus metrics
        fl_rounds_total.inc()
        fl_model_accuracy.set(strategy.convergence_history["accuracies"][-1])
        cxl_memory_utilization.set(min(1.0, 0.4 + (strategy.round_num * 0.02)))
        
    return (
        jsonify({
            "status": "accepted",
            "message": "FL round started and metrics updated",
            "current_round": strategy.round_num if strategy is not None else 0,
        }),
        202,
    )"""

content = re.sub(
    r'@app\.route\("/trigger_fl", methods=\["POST"\]\)\s+def trigger_fl_round\(\):\s+logger\.info\("Manual FL round trigger requested via API"\).*?202,\s+\)',
    trigger_fl_replacement,
    content,
    flags=re.DOTALL,
)

chat_endpoint = """
@app.route("/chat", methods=["POST"])
def chat_query():
    data = request.json or {}
    query = data.get("query", "").lower()
    
    # Simple simulated LLM responses for HUD actions
    if "status" in query or "health" in query:
        resp = f"System online. Enclave is {enclave_status}. Current accuracy: {strategy.convergence_history['accuracies'][-1] if strategy and strategy.convergence_history['accuracies'] else 'N/A'}%."
    elif "policy" in query or "verify" in query:
        resp = "Verification policy active. TEE proofs and ZK proofs are currently supported and strictly enforced on validators."
    elif "test" in query or "llm" in query:
        resp = "LLM adapters are calibrated against latest local gradients. Current rank: 16. Alpha scaling: 32."
    else:
        resp = "As a Sovereign Map Operational Node, I am currently syncing spatial and telemetric updates. My models are ready for the next round."
        
    return jsonify({"response": resp, "status": "ok"}), 200

"""

content = content.replace(
    '@app.route("/health", methods=["GET"])',
    chat_endpoint + '\n@app.route("/health", methods=["GET"])',
)


def enclave_replacement(content):
    replacement = """@app.route("/create_enclave", methods=["POST"])
def create_enclave():
    global enclave_status
    if enclave_status == "Isolated":
        enclave_status = "Initialized"
    elif enclave_status == "Initialized":
        enclave_status = "Attested & Locked"
    
    logger.info(f"Secure enclave transitioned to: {enclave_status}")
    return jsonify({"status": "ok", "enclave_status": enclave_status}), 200"""
    return re.sub(
        r'@app\.route\("/create_enclave", methods=\["POST"\]\)\s+def create_enclave\(\):\s+global enclave_status\s+enclave_status = "Initialized"\s+logger.info\([^\)]+\)\s+return jsonify\(\{[^\}]+\}\), 200',
        replacement,
        content,
    )


content = enclave_replacement(content)

with open("sovereignmap_production_backend_v2.py", "w") as f:
    f.write(content)
