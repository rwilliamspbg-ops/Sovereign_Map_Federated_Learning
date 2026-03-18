import requests
import time
import sys
import os

# Load Aggregator IP from environment or command line
AGGREGATOR_IP = os.getenv('AGGREGATOR_IP')
if len(sys.argv) > 1:
    AGGREGATOR_IP = sys.argv[1]

if not AGGREGATOR_IP:
    print("❌ Error: AGGREGATOR_IP not set. Usage: python3 verify-mesh.py <IP>")
    sys.exit(1)

URL = f"http://{AGGREGATOR_IP}:8081/status"
TARGET_NODES = 200

print(f"📡 Pinging Aggregator Mesh: {URL}")
print(f"🎯 Target: {TARGET_NODES} active nodes...")

def check_mesh():
    try:
        response = requests.get(URL, timeout=5)
        data = response.json()
        
        # Adjust 'active_nodes' based on your specific JSON response keys
        active_count = data.get('active_nodes', 0)
        peers = data.get('peers', [])
        
        # Calculate health percentage
        health = (active_count / TARGET_NODES) * 100
        
        print(f"\n--- Mesh Status [{time.strftime('%H:%M:%S')}] ---")
        print(f"✅ Nodes Online: {active_count} / {TARGET_NODES} ({health:.1f}%)")
        
        if active_count < TARGET_NODES:
            print(f"⚠️  Missing: {TARGET_NODES - active_count} nodes.")
            # Identify missing hosts if peer data is available
            hosts = {}
            for p in peers:
                host_ip = p.split('_')[1].replace('-', '.') # Reconstruct IP from NodeID
                hosts[host_ip] = hosts.get(host_ip, 0) + 1
            
            print("📍 Node Distribution per Host:")
            for ip, count in hosts.items():
                status = "🟢 OK" if count == 25 else "🔴 FAILING"
                print(f"   {ip}: {count}/25 nodes {status}")

        return active_count >= TARGET_NODES

    except Exception as e:
        print(f"❌ Connection Failed: {e}")
        return False

# Run a quick 3-attempt loop to wait for cold-start containers
for i in range(3):
    if check_mesh():
        print("\n🚀 ALL SYSTEMS GO: 200 nodes verified. Start Phase 4.")
        break
    if i < 2:
        print("\n⏳ Waiting 10s for nodes to warm up...")
        time.sleep(10)
