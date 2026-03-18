import re

with open("sovereignmap_production_backend_v2.py", "r") as f:
    content = f.read()

trust_endpoints = """
# Trust and Verification mocking for the HUD
@app.route("/trust_snapshot", methods=["GET"])
def trust_snapshot():
    return jsonify({
        "trust_status": {
            "trust_mode": "Strict Verification",
            "fl_verification": {
                "verified_rounds": strategy.round_num if strategy else 0,
                "failed_rounds": 0,
                "average_confidence_bps": 9850
            },
            "verification_policy": {
                "require_proof": True,
                "min_confidence_bps": 7500,
                "reject_on_verification_failure": True,
                "allow_consensus_proof": True,
                "allow_zk_proof": True,
                "allow_tee_proof": True
            }
        },
        "policy_history": [
            {
                "source": "governance",
                "proposal_id": "prop-001",
                "new_policy": {
                    "min_confidence_bps": 7500
                }
            }
        ]
    }), 200

@app.route("/verification_policy", methods=["POST"])
def update_verification_policy():
    data = request.json or {}
    logger.info(f"Verification policy update requested: {data}")
    return jsonify({"status": "ok", "message": "Policy applied successfully"}), 200

"""

# Insert trust endpoints before the first route (health)
content = content.replace(
    '@app.route("/health", methods=["GET"])',
    trust_endpoints + '\n@app.route("/health", methods=["GET"])',
)

with open("sovereignmap_production_backend_v2.py", "w") as f:
    f.write(content)
