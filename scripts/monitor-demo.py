#!/usr/bin/env python3
"""
Real-time monitoring for Sovereign Map Demo
Fetches live metrics from Prometheus and Grafana
"""

import requests
import json
import time
import sys
from datetime import datetime


def check_services():
    """Check all monitoring services are running"""

    services = {
        "Prometheus": "http://localhost:9090/-/healthy",
        "Grafana": "http://localhost:3001/api/health",
        "Backend": "http://localhost:8000/health",
    }

    print(f"\n{'='*60}")
    print(f"Sovereign Map Monitoring Status")
    print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"{'='*60}\n")

    all_healthy = True

    for service_name, endpoint in services.items():
        try:
            response = requests.get(endpoint, timeout=5)
            status = (
                "✅ UP"
                if response.status_code == 200
                else f"⚠️  Code {response.status_code}"
            )
        except Exception as e:
            status = f"❌ DOWN: {str(e)[:30]}"
            all_healthy = False

        print(f"{service_name:15} {status}")

    print()

    # Query Prometheus for active metrics
    try:
        response = requests.get(
            "http://localhost:9090/api/v1/label/__name__/values", timeout=5
        )
        if response.status_code == 200:
            metrics = response.json().get("data", [])
            print(f"📊 Prometheus Metrics Available: {len(metrics)}")
            print(f"   Sample: {', '.join(metrics[:5])}")
        else:
            print(f"⚠️  Prometheus metrics unavailable (code {response.status_code})")
    except Exception as e:
        print(f"❌ Prometheus query failed: {str(e)[:40]}")

    print()

    # Query running containers
    try:
        response = requests.get(
            "http://localhost:9090/api/v1/query?query=up", timeout=5
        )
        if response.status_code == 200:
            data = response.json().get("data", {}).get("result", [])
            print(f"🐳 Containers Reporting Metrics: {len(data)}")
        else:
            print(f"⚠️  Container metrics unavailable")
    except Exception as e:
        print(f"❌ Container query failed: {str(e)[:40]}")

    print(f"\n{'='*60}")
    print(f"Monitoring URLs:")
    print(f"  • Grafana:     http://localhost:3001")
    print(f"  • Prometheus: http://localhost:9090")
    print(f"  • Backend:    http://localhost:8000")
    print(f"{'='*60}\n")

    return all_healthy


def get_prometheus_query(query):
    """Execute a Prometheus query"""
    try:
        response = requests.get(
            "http://localhost:9090/api/v1/query", params={"query": query}, timeout=5
        )
        if response.status_code == 200:
            return response.json().get("data", {}).get("result", [])
        return None
    except Exception as e:
        print(f"Query failed: {e}")
        return None


if __name__ == "__main__":
    # Run once or continuously
    continuous = "--continuous" in sys.argv or "-c" in sys.argv
    interval = 30

    if continuous:
        print("Continuous monitoring mode (Ctrl+C to stop)")
        try:
            while True:
                check_services()
                time.sleep(interval)
        except KeyboardInterrupt:
            print("\nMonitoring stopped.")
            sys.exit(0)
    else:
        check_services()
