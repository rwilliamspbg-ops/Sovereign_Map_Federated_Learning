import subprocess

NODES = 2500
SERVICE_NAME = "federated-learning" 

def run_simulation():
    print(f"🚀 Scaling {SERVICE_NAME} to {NODES} nodes (Port-Free Mode)...")
    try:
        # Use 'docker compose' V2 with internal networking only
        subprocess.run(["docker", "compose", "up", "-d", "--scale", f"{SERVICE_NAME}={NODES}"], check=True)
        print(f"✅ {NODES} Nodes Active. Bypassing host port limits.")
    except subprocess.CalledProcessError as e:
        print(f"❌ Simulation failed: {e}")

if __name__ == "__main__":
    run_simulation()
