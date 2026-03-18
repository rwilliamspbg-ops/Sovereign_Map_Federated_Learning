import re

with open('tokenomics_metrics_exporter.py', 'r') as f:
    content = f.read()

replacement = """
def run_simulation(exporter):
    import time, random
    last_escrow = 85000450
    last_wallets = 8540
    last_balance = 14630
    while True:
        last_escrow += random.randint(-50000, 100000)
        last_wallets += random.randint(0, 5)
        last_balance += random.randint(-10, 15)
        payload = {
          "mint_rate_per_min": random.uniform(140.0, 160.0),
          "token_supply_total": 1000000000,
          "token_supply_minted": 125000000 + random.randint(1000, 5000),
          "bridge_inflow_per_min": random.uniform(400.0, 500.0),
          "bridge_outflow_per_min": random.uniform(250.0, 350.0),
          "bridge_escrow_total": last_escrow,
          "bridge_collateral_ratio_percent": random.uniform(150.0, 160.0),
          "unique_wallets_count": last_wallets,
          "wallet_average_balance": last_balance,
          "top_10_holder_concentration_percent": 12.4,
          "wallet_liquidity_ratio_percent": random.uniform(34.0, 35.0)
        }
        exporter.ingest_event(payload)
        time.sleep(10)

def create_app(source_file: str):
    app = Flask(__name__)
    exporter = TokenomicsMetricsExporter(source_file=source_file)
    
    import threading
    t = threading.Thread(target=run_simulation, args=(exporter,), daemon=True)
    t.start()
"""

# Replace the create_app definition
content = re.sub(r'def create_app\(source_file: str\):(.*?)exporter = TokenomicsMetricsExporter\(source_file=source_file\)', replacement, content, flags=re.DOTALL)

with open('tokenomics_metrics_exporter.py', 'w') as f:
    f.write(content)

