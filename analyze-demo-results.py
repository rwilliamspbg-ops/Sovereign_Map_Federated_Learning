#!/usr/bin/env python3
"""
Sovereign Map Demo Results Analyzer and Visualizer
Generates charts and reports from demo execution results
"""

import json
import sys
import re
from datetime import datetime
from pathlib import Path

def analyze_demo_results(results_dir):
    """Parse and analyze all results from a demo run"""
    
    results_path = Path(results_dir)
    if not results_path.exists():
        print(f"ERROR: Results directory not found: {results_dir}")
        sys.exit(1)
    
    print(f"\n{'='*60}")
    print(f"Sovereign Map Demo Results Analysis")
    print(f"{'='*60}")
    print(f"Results Directory: {results_dir}")
    print(f"Scan Time: {datetime.now()}")
    print(f"{'='*60}\n")
    
    # Parse log file
    log_file = results_path / "demo.log"
    if log_file.exists():
        print("📋 EXECUTION LOG SUMMARY")
        print("-" * 60)
        
        with open(log_file, 'r') as f:
            lines = f.readlines()
            for line in lines[:10]:  # First 10 lines
                print(line.rstrip())
            if len(lines) > 20:
                print(f"... ({len(lines) - 20} lines omitted)")
            for line in lines[-10:]:  # Last 10 lines
                print(line.rstrip())
        print()
    
    # Analyze metrics iterations
    metrics_files = sorted(results_path.glob("metrics-iteration-*.txt"))
    if metrics_files:
        print("📊 METRICS ITERATIONS COLLECTED")
        print("-" * 60)
        print(f"Total iterations: {len(metrics_files)}")
        
        # Parse first and last iteration
        first_metrics = parse_metrics_file(metrics_files[0])
        last_metrics = parse_metrics_file(metrics_files[-1])
        
        print(f"\nFirst Iteration (metrics-iteration-1.txt):")
        print_metrics(first_metrics)
        
        if len(metrics_files) > 1:
            print(f"\nFinal Iteration (metrics-iteration-{len(metrics_files)}.txt):")
            print_metrics(last_metrics)
        print()
    
    # Analyze Prometheus ranges
    prometheus_files = sorted(results_path.glob("prometheus-range-*.json"))
    if prometheus_files:
        print("📈 PROMETHEUS METRICS EXPORTED")
        print("-" * 60)
        print(f"Time series snapshots: {len(prometheus_files)}")
        print(f"Files: {', '.join([f.name for f in prometheus_files[:3]])}...")
        print()
    
    # Read final state
    final_state = results_path / "final-state.txt"
    if final_state.exists():
        print("🔍 FINAL SYSTEM STATE")
        print("-" * 60)
        with open(final_state, 'r') as f:
            content = f.read()
            # Extract key info
            if "node-agent-" in content:
                agent_count = len(re.findall(r'node-agent-\d+', content))
                print(f"Node agents found in final state: {agent_count}")
            print(content[:500] + "..." if len(content) > 500 else content)
        print()
    
    # Read summary
    summary_file = results_path / "DEMO_REPORT.md"
    if summary_file.exists():
        print("📄 DEMO REPORT")
        print("-" * 60)
        with open(summary_file, 'r') as f:
            print(f.read()[:1000] + "..." if len(f.read()) > 1000 else f.read())
        print()
    
    print(f"{'='*60}")
    print("✅ Analysis complete. Review files above.")
    print(f"{'='*60}\n")

def parse_metrics_file(filepath):
    """Extract metrics from iteration file"""
    metrics = {}
    with open(filepath, 'r') as f:
        content = f.read()
        
        # Extract node count
        if "Running:" in content:
            match = re.search(r'Running: (\d+), Exited: (\d+)', content)
            if match:
                metrics['running'] = int(match.group(1))
                metrics['exited'] = int(match.group(2))
    
    return metrics

def print_metrics(metrics):
    """Pretty print metrics"""
    for key, value in metrics.items():
        print(f"  {key}: {value}")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        results_dir = sys.argv[1]
    else:
        # Auto-detect latest demo results
        base_dir = Path("test-results/demo-1000nodes")
        if base_dir.exists():
            dirs = sorted([d for d in base_dir.iterdir() if d.is_dir()])
            if dirs:
                results_dir = str(dirs[-1])
            else:
                print("ERROR: No demo results found")
                sys.exit(1)
        else:
            print(f"ERROR: Results directory not found: {base_dir}")
            sys.exit(1)
    
    analyze_demo_results(results_dir)
