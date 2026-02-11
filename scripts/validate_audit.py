import os
import sys

# Required SGP-001 Audit Standards
REQUIRED_CONFIG = {
    "PRIVACY_BUDGET_EPSILON": "1.0",
    "PRIVACY_DELTA": "1e-5",
    "MIN_PARTICIPANTS": "10",
    "TARGET_THROUGHPUT_TOPS": "85"
}

def validate_env():
    print("🛡️ Starting SGP-001 Audit Validation...")
    errors = []
    
    for var, expected in REQUIRED_CONFIG.items():
        actual = os.getenv(var)
        if actual != expected:
            errors.append(f"Mismatch in {var}: Expected {expected}, got {actual}")
            
    # Check if the audit file is actually mounted
    if not os.path.exists("/app/config/privacy_policy.md"):
        errors.append("Critical Error: /app/config/privacy_policy.md not found (Mount Failure)")

    if errors:
        print("\n❌ Audit Validation Failed:")
        for error in errors:
            print(f"  - {error}")
        print("\n⚠️ Entering 'Independent Island' Mode: Service Startup Aborted.")
        sys.exit(1)

    print("✅ Audit Validation Passed. Initializing Federated Learning Node...")

if __name__ == "__main__":
    validate_env()
