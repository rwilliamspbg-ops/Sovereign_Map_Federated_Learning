#!/usr/bin/env python3
"""
Real-Time Convergence Monitoring Dashboard
Sovereign Map Federated Learning System
"""

import json
import requests
import time
import subprocess
from datetime import datetime
from collections import deque

class ConvergenceMonitor:
    def __init__(self):
        self.backend_url = "http://localhost:8081"
        self.history = deque(maxlen=50)
        self.round_times = []
    
    def fetch_convergence(self):
        try:
            resp = requests.get(f"{self.backend_url}/convergence", timeout=5)
            return resp.json()
        except:
            return None
    
    def fetch_metrics_summary(self):
        try:
            resp = requests.get(f"{self.backend_url}/metrics_summary", timeout=5)
            return resp.json()
        except:
            return None
    
    def get_latest_logs(self, n=20):
        try:
            result = subprocess.run(
                ["docker", "logs", "sovereign_map_federated_learning-backend-1", "--tail", str(n)],
                capture_output=True,
                text=True,
                timeout=10
            )
            return result.stdout
        except:
            return ""
    
    def display_header(self):
        print("\033[2J\033[H")  # Clear screen
        print("╔" + "═" * 78 + "╗")
        print("║" + " " * 20 + "🌐 SOVEREIGN MAP CONVERGENCE MONITOR" + " " * 23 + "║")
        print("║" + " " * 15 + "Federated Learning with 50 Node Agents" + " " * 25 + "║")
        print("╚" + "═" * 78 + "╝")
    
    def display_convergence(self, data):
        if not data:
            print("❌ Cannot fetch convergence data")
            return
        
        rounds = data.get('rounds', [])
        accuracies = data.get('accuracies', [])
        losses = data.get('losses', [])
        
        print("\n📊 CONVERGENCE DATA:")
        print("-" * 80)
        
        if rounds:
            print(f"Current Round:     {data.get('current_round', 'N/A')}")
            print(f"Current Accuracy:  {data.get('current_accuracy', 0):.2f}%")
            print(f"Current Loss:      {data.get('current_loss', 0):.4f}")
            
            print("\nRound-by-Round Breakdown:")
            print(f"{'Round':<8} {'Timestamp':<20} {'Accuracy':<12} {'Loss':<10} {'Trend':<10}")
            print("-" * 60)
            
            for i in range(max(0, len(rounds) - 10), len(rounds)):
                round_num = rounds[i]
                acc = accuracies[i] if i < len(accuracies) else 0
                loss = losses[i] if i < len(losses) else 0
                
                # Calculate trend
                if i > 0:
                    prev_acc = accuracies[i-1]
                    delta = acc - prev_acc
                    trend = f"+{delta:.2f}%" if delta > 0 else f"{delta:.2f}%"
                else:
                    trend = "baseline"
                
                print(f"{round_num:<8} {str(datetime.now()):<20} {acc:>10.2f}% {loss:>8.4f}  {trend:<10}")
        else:
            print("⏳ Waiting for first round to complete...")
    
    def display_metrics(self, metrics):
        if not metrics:
            return
        
        print("\n📈 SYSTEM METRICS:")
        print("-" * 80)
        
        fl = metrics.get('federated_learning', {})
        nodes = metrics.get('nodes', {})
        
        print(f"Rounds Completed:  {fl.get('current_round', 0)}")
        print(f"Avg Node Stake:    {nodes.get('avg_stake', 0):.2f} tokens")
        print(f"Total Network Stake: {nodes.get('total_stake', 0):.2f} tokens")
        print(f"Active Nodes:      50 agents")
        
        acc_hist = fl.get('accuracy_history', [])
        if acc_hist:
            print(f"Recent Accuracies: {' -> '.join([f'{a:.1f}%' for a in acc_hist[-5:]])}")
    
    def display_nodes(self):
        print("\n🖥️  NODE CLUSTER STATUS:")
        print("-" * 80)
        
        try:
            result = subprocess.run(
                ["docker", "ps", "-a", "--filter", "label=com.docker.compose.project=sovereign_map_federated_learning", "--format", "{{.Names}}\t{{.Status}}"],
                capture_output=True,
                text=True,
                timeout=5
            )
            
            lines = result.stdout.strip().split('\n')
            running = sum(1 for line in lines if 'Up' in line)
            total = len(lines)
            
            print(f"Total Containers:  {total}")
            print(f"Running:           {running} ✅")
            print(f"Status:            {running}/{total} healthy")
            
            # Count node-agents
            node_agents = [line for line in lines if 'node-agent' in line]
            print(f"Node Agents:       {len(node_agents)}/50")
            
        except:
            print("❌ Could not fetch node status")
    
    def run(self):
        while True:
            try:
                self.display_header()
                
                conv_data = self.fetch_convergence()
                metrics = self.fetch_metrics_summary()
                
                self.display_convergence(conv_data)
                self.display_metrics(metrics)
                self.display_nodes()
                
                print("\n" + "=" * 80)
                print(f"Last updated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')} | Press Ctrl+C to exit")
                print("=" * 80)
                
                time.sleep(5)
                
            except KeyboardInterrupt:
                print("\n\n👋 Monitoring stopped.")
                break
            except Exception as e:
                print(f"\n❌ Error: {e}")
                time.sleep(5)

if __name__ == "__main__":
    monitor = ConvergenceMonitor()
    monitor.run()
