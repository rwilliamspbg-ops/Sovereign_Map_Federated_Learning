import json
import os


def create_dashboard(title, uid, panels):
    return {
        "annotations": {"list": []},
        "editable": True,
        "fiscalYearStartMonth": 0,
        "graphTooltip": 0,
        "id": None,
        "links": [],
        "liveNow": False,
        "panels": panels,
        "refresh": "5s",
        "schemaVersion": 38,
        "style": "dark",
        "tags": [title.lower().replace(" ", "-")],
        "templating": {"list": []},
        "time": {"from": "now-6h", "to": "now"},
        "timepicker": {},
        "timezone": "",
        "title": title,
        "uid": uid,
        "version": 2,
    }


def build_thresholds(steps):
    return {
        "mode": "absolute",
        "steps": [{"color": color, "value": value} for value, color in steps],
    }


def create_row_panel(title, pid, y):
    return {
        "collapsed": False,
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": y},
        "id": pid,
        "panels": [],
        "title": title,
        "type": "row",
    }


def create_metric_panel(
    title,
    pid,
    grid_pos,
    targets,
    panel_type="timeseries",
    unit="none",
    min_value=None,
    max_value=None,
    decimals=None,
    thresholds=None,
):
    defaults = {
        "color": {"mode": "thresholds" if thresholds else "palette-classic"},
        "mappings": [],
        "unit": unit,
    }
    if min_value is not None:
        defaults["min"] = min_value
    if max_value is not None:
        defaults["max"] = max_value
    if decimals is not None:
        defaults["decimals"] = decimals
    if thresholds:
        defaults["thresholds"] = build_thresholds(thresholds)

    panel = {
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "fieldConfig": {
            "defaults": defaults,
            "overrides": [],
        },
        "gridPos": grid_pos,
        "id": pid,
        "targets": targets,
        "title": title,
        "type": panel_type,
    }

    if panel_type == "timeseries":
        panel["fieldConfig"]["defaults"]["custom"] = {
            "axisCenteredZero": False,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {"legend": False, "tooltip": False, "viz": False},
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 3,
            "scaleDistribution": {"type": "linear"},
            "showPoints": "auto",
            "spanNulls": True,
            "stacking": {"group": "A", "mode": "none"},
            "thresholdsStyle": {"mode": "line"},
        }
        panel["options"] = {
            "legend": {
                "calcs": ["lastNotNull", "max", "min"],
                "displayMode": "table",
                "placement": "bottom",
                "showLegend": True,
            },
            "tooltip": {"mode": "multi", "sort": "desc"},
        }
    elif panel_type == "stat":
        panel["options"] = {
            "colorMode": "value",
            "graphMode": "area",
            "justifyMode": "auto",
            "orientation": "auto",
            "reduceOptions": {
                "calcs": ["lastNotNull"],
                "fields": "",
                "values": False,
            },
            "textMode": "auto",
        }
    elif panel_type == "gauge":
        panel["options"] = {
            "orientation": "auto",
            "reduceOptions": {
                "calcs": ["lastNotNull"],
                "fields": "",
                "values": False,
            },
            "showThresholdLabels": True,
            "showThresholdMarkers": True,
        }

    return panel


def apply_y_offset(panels, y_offset):
    shifted = []
    for panel in panels:
        panel_copy = dict(panel)
        grid_pos = dict(panel_copy.get("gridPos", {}))
        if "y" in grid_pos:
            grid_pos["y"] = grid_pos["y"] + y_offset
        panel_copy["gridPos"] = grid_pos
        if panel_copy.get("type") == "row" and "panels" in panel_copy:
            panel_copy["panels"] = list(panel_copy["panels"])
        shifted.append(panel_copy)
    return shifted


tokenomics_section_panels = [
    create_row_panel("Economy", 1, 0),
    create_metric_panel(
        "Mint Rate / min",
        2,
        {"h": 5, "w": 6, "x": 0, "y": 1},
        [
            {
                "expr": "tokenomics_mint_rate_per_min or sovereignmap_token_mint_rate_per_min or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=2,
        thresholds=[(None, "red"), (50, "yellow"), (120, "green")],
    ),
    create_metric_panel(
        "Total Supply",
        3,
        {"h": 5, "w": 6, "x": 6, "y": 1},
        [
            {
                "expr": "tokenomics_token_supply_total or sovereignmap_token_supply_total or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=0,
    ),
    create_metric_panel(
        "Minted Supply",
        4,
        {"h": 5, "w": 6, "x": 12, "y": 1},
        [
            {
                "expr": "tokenomics_token_supply_minted or sovereignmap_token_supply_minted or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=0,
    ),
    create_metric_panel(
        "Minted Share (%)",
        5,
        {"h": 5, "w": 6, "x": 18, "y": 1},
        [
            {
                "expr": "((tokenomics_token_supply_minted or sovereignmap_token_supply_minted or vector(0)) / clamp_min((tokenomics_token_supply_total or sovereignmap_token_supply_total or vector(1)), 1)) * 100",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        unit="percent",
        min_value=0,
        max_value=100,
        decimals=2,
        thresholds=[(None, "red"), (25, "yellow"), (45, "green")],
    ),
    create_metric_panel(
        "Mint and Supply Trend",
        6,
        {"h": 7, "w": 24, "x": 0, "y": 6},
        [
            {
                "expr": "tokenomics_mint_rate_per_min or sovereignmap_token_mint_rate_per_min or vector(0)",
                "legendFormat": "mint_per_min",
                "refId": "A",
            },
            {
                "expr": "tokenomics_token_supply_minted or sovereignmap_token_supply_minted or vector(0)",
                "legendFormat": "minted_supply",
                "refId": "B",
            },
            {
                "expr": "tokenomics_token_supply_total or sovereignmap_token_supply_total or vector(0)",
                "legendFormat": "total_supply",
                "refId": "C",
            },
        ],
        unit="short",
        decimals=2,
    ),
    create_row_panel("Bridge", 7, 13),
    create_metric_panel(
        "Net Bridge Flow / min",
        8,
        {"h": 5, "w": 6, "x": 0, "y": 14},
        [
            {
                "expr": "tokenomics_bridge_net_flow_per_min or sovereignmap_bridge_net_flow_per_min or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=2,
        thresholds=[(None, "red"), (0, "yellow"), (20, "green")],
    ),
    create_metric_panel(
        "Collateral Ratio",
        9,
        {"h": 5, "w": 6, "x": 6, "y": 14},
        [
            {
                "expr": "tokenomics_bridge_collateral_ratio_percent or sovereignmap_bridge_collateral_ratio_percent or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        unit="percent",
        min_value=0,
        max_value=220,
        decimals=2,
        thresholds=[(None, "red"), (120, "yellow"), (150, "green")],
    ),
    create_metric_panel(
        "Settlement Share",
        10,
        {"h": 5, "w": 6, "x": 12, "y": 14},
        [
            {
                "expr": "tokenomics_bridge_settlement_share_percent or sovereignmap_bridge_settlement_share_percent or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        unit="percent",
        min_value=0,
        max_value=100,
        decimals=2,
        thresholds=[(None, "green"), (45, "yellow"), (70, "red")],
    ),
    create_metric_panel(
        "Escrow Total",
        11,
        {"h": 5, "w": 6, "x": 18, "y": 14},
        [
            {
                "expr": "tokenomics_bridge_escrow_total or sovereignmap_bridge_escrow_total or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=0,
    ),
    create_metric_panel(
        "Bridge Flow Dynamics",
        12,
        {"h": 7, "w": 12, "x": 0, "y": 19},
        [
            {
                "expr": "tokenomics_bridge_inflow_per_min or sovereignmap_bridge_inflow_per_min or vector(0)",
                "legendFormat": "inflow",
                "refId": "A",
            },
            {
                "expr": "tokenomics_bridge_outflow_per_min or sovereignmap_bridge_outflow_per_min or vector(0)",
                "legendFormat": "outflow",
                "refId": "B",
            },
            {
                "expr": "tokenomics_bridge_net_flow_per_min or sovereignmap_bridge_net_flow_per_min or vector(0)",
                "legendFormat": "net",
                "refId": "C",
            },
        ],
        unit="short",
        decimals=2,
    ),
    create_metric_panel(
        "Bridge Liquidity Trend",
        13,
        {"h": 7, "w": 12, "x": 12, "y": 19},
        [
            {
                "expr": "tokenomics_bridge_escrow_total or sovereignmap_bridge_escrow_total or vector(0)",
                "legendFormat": "escrow_total",
                "refId": "A",
            },
            {
                "expr": "tokenomics_bridge_volume_24h or sovereignmap_bridge_volume_24h or vector(0)",
                "legendFormat": "volume_24h",
                "refId": "B",
            },
        ],
        unit="short",
        decimals=2,
    ),
    create_row_panel("Validator", 14, 26),
    create_metric_panel(
        "Active Validators",
        15,
        {"h": 5, "w": 6, "x": 0, "y": 27},
        [
            {
                "expr": "tokenomics_validator_count or sovereignmap_validators_count or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=0,
        thresholds=[(None, "red"), (5, "yellow"), (15, "green")],
    ),
    create_metric_panel(
        "Stake Participation",
        16,
        {"h": 5, "w": 6, "x": 6, "y": 27},
        [
            {
                "expr": "(tokenomics_stake_participation_ratio or sovereignmap_stake_participation_ratio or vector(0)) * 100",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        unit="percent",
        min_value=0,
        max_value=100,
        decimals=2,
        thresholds=[(None, "red"), (50, "yellow"), (75, "green")],
    ),
    create_metric_panel(
        "Stake Concentration (Gini)",
        17,
        {"h": 5, "w": 6, "x": 12, "y": 27},
        [
            {
                "expr": "tokenomics_stake_concentration_gini or sovereignmap_stake_concentration_gini or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        min_value=0,
        max_value=1,
        decimals=3,
        thresholds=[(None, "green"), (0.4, "yellow"), (0.65, "red")],
    ),
    create_metric_panel(
        "Validator Security Trend",
        18,
        {"h": 7, "w": 24, "x": 0, "y": 32},
        [
            {
                "expr": "tokenomics_validator_count or sovereignmap_validators_count or vector(0)",
                "legendFormat": "validators",
                "refId": "A",
            },
            {
                "expr": "(tokenomics_stake_participation_ratio or sovereignmap_stake_participation_ratio or vector(0)) * 100",
                "legendFormat": "stake_participation_percent",
                "refId": "B",
            },
            {
                "expr": "(tokenomics_stake_concentration_gini or sovereignmap_stake_concentration_gini or vector(0)) * 100",
                "legendFormat": "gini_percent",
                "refId": "C",
            },
        ],
        unit="short",
        decimals=2,
    ),
    create_row_panel("Wallets", 19, 39),
    create_metric_panel(
        "Unique Wallets",
        20,
        {"h": 5, "w": 6, "x": 0, "y": 40},
        [
            {
                "expr": "tokenomics_unique_wallets_count or sovereignmap_unique_wallets_count or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=0,
    ),
    create_metric_panel(
        "Average Wallet Balance",
        21,
        {"h": 5, "w": 6, "x": 6, "y": 40},
        [
            {
                "expr": "tokenomics_wallet_average_balance or sovereignmap_wallet_average_balance or vector(0)",
                "refId": "A",
            }
        ],
        panel_type="stat",
        unit="short",
        decimals=2,
    ),
    create_metric_panel(
        "Top 10 Holder Concentration",
        22,
        {"h": 5, "w": 6, "x": 12, "y": 40},
        [
            {
                "expr": "(tokenomics_top_10_holder_concentration or sovereignmap_top_10_holder_concentration or vector(0)) * 100",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        unit="percent",
        min_value=0,
        max_value=100,
        decimals=2,
        thresholds=[(None, "green"), (20, "yellow"), (35, "red")],
    ),
    create_metric_panel(
        "Wallet Liquidity Ratio",
        23,
        {"h": 5, "w": 6, "x": 18, "y": 40},
        [
            {
                "expr": "(tokenomics_wallet_liquidity_ratio or sovereignmap_wallet_liquidity_ratio or vector(0)) * 100",
                "refId": "A",
            }
        ],
        panel_type="gauge",
        unit="percent",
        min_value=0,
        max_value=100,
        decimals=2,
        thresholds=[(None, "red"), (25, "yellow"), (45, "green")],
    ),
    create_metric_panel(
        "Wallet Buckets",
        24,
        {"h": 7, "w": 12, "x": 0, "y": 45},
        [
            {
                "expr": "tokenomics_wallets_by_balance_bucket_large or sovereignmap_wallets_by_balance_bucket_large or vector(0)",
                "legendFormat": "large",
                "refId": "A",
            },
            {
                "expr": "tokenomics_wallets_by_balance_bucket_medium or sovereignmap_wallets_by_balance_bucket_medium or vector(0)",
                "legendFormat": "medium",
                "refId": "B",
            },
            {
                "expr": "tokenomics_wallets_by_balance_bucket_small or sovereignmap_wallets_by_balance_bucket_small or vector(0)",
                "legendFormat": "small",
                "refId": "C",
            },
        ],
        unit="short",
        decimals=0,
    ),
    create_metric_panel(
        "Distribution Risk and Liquidity",
        25,
        {"h": 7, "w": 12, "x": 12, "y": 45},
        [
            {
                "expr": "(tokenomics_top_10_holder_concentration or sovereignmap_top_10_holder_concentration or vector(0)) * 100",
                "legendFormat": "top10_concentration_percent",
                "refId": "A",
            },
            {
                "expr": "(tokenomics_wallet_liquidity_ratio or sovereignmap_wallet_liquidity_ratio or vector(0)) * 100",
                "legendFormat": "wallet_liquidity_percent",
                "refId": "B",
            },
        ],
        unit="percent",
        decimals=2,
    ),
]

EXECUTIVE_SCORE_PROFILES = {
    "conservative": {
        "economy_mint_weight": 0.30,
        "economy_minted_weight": 0.70,
        "bridge_collateral_weight": 0.80,
        "bridge_net_flow_weight": 0.20,
        "validator_participation_weight": 0.60,
        "validator_decentralization_weight": 0.40,
        "wallet_liquidity_weight": 0.35,
        "wallet_distribution_weight": 0.65,
    },
    "balanced": {
        "economy_mint_weight": 0.40,
        "economy_minted_weight": 0.60,
        "bridge_collateral_weight": 0.70,
        "bridge_net_flow_weight": 0.30,
        "validator_participation_weight": 0.70,
        "validator_decentralization_weight": 0.30,
        "wallet_liquidity_weight": 0.50,
        "wallet_distribution_weight": 0.50,
    },
    "growth": {
        "economy_mint_weight": 0.60,
        "economy_minted_weight": 0.40,
        "bridge_collateral_weight": 0.55,
        "bridge_net_flow_weight": 0.45,
        "validator_participation_weight": 0.80,
        "validator_decentralization_weight": 0.20,
        "wallet_liquidity_weight": 0.65,
        "wallet_distribution_weight": 0.35,
    },
}


def build_tokenomics_summary_panels(profile_name):
    profile = EXECUTIVE_SCORE_PROFILES[profile_name]
    return [
        create_row_panel("Executive Summary", 100, 0),
        create_metric_panel(
            "Economic Growth Score",
            101,
            {"h": 5, "w": 6, "x": 0, "y": 1},
            [
                {
                    "expr": f"clamp_max(clamp_min((((tokenomics_mint_rate_per_min or sovereignmap_token_mint_rate_per_min or vector(0)) / 150) * 100 * {profile['economy_mint_weight']}) + ((((tokenomics_token_supply_minted or sovereignmap_token_supply_minted or vector(0)) / clamp_min((tokenomics_token_supply_total or sovereignmap_token_supply_total or vector(1)), 1)) * 100) * {profile['economy_minted_weight']}), 100), 0)",
                    "refId": "A",
                }
            ],
            panel_type="stat",
            unit="percent",
            decimals=1,
            thresholds=[(None, "red"), (50, "yellow"), (70, "green")],
        ),
        create_metric_panel(
            "Bridge Health Score",
            102,
            {"h": 5, "w": 6, "x": 6, "y": 1},
            [
                {
                    "expr": f"clamp_max(clamp_min((((tokenomics_bridge_collateral_ratio_percent or sovereignmap_bridge_collateral_ratio_percent or vector(0)) / 150) * 100 * {profile['bridge_collateral_weight']}) + ((clamp_min(clamp_max((tokenomics_bridge_net_flow_per_min or sovereignmap_bridge_net_flow_per_min or vector(0)) + 20, 0), 40) / 40) * 100 * {profile['bridge_net_flow_weight']}), 100), 0)",
                    "refId": "A",
                }
            ],
            panel_type="stat",
            unit="percent",
            decimals=1,
            thresholds=[(None, "red"), (50, "yellow"), (70, "green")],
        ),
        create_metric_panel(
            "Validator Health Score",
            103,
            {"h": 5, "w": 6, "x": 12, "y": 1},
            [
                {
                    "expr": f"clamp_max(clamp_min((((tokenomics_stake_participation_ratio or sovereignmap_stake_participation_ratio or vector(0)) * 100) * {profile['validator_participation_weight']}) + (((1 - (tokenomics_stake_concentration_gini or sovereignmap_stake_concentration_gini or vector(0))) * 100) * {profile['validator_decentralization_weight']}), 100), 0)",
                    "refId": "A",
                }
            ],
            panel_type="stat",
            unit="percent",
            decimals=1,
            thresholds=[(None, "red"), (50, "yellow"), (70, "green")],
        ),
        create_metric_panel(
            "Wallet Health Score",
            104,
            {"h": 5, "w": 6, "x": 18, "y": 1},
            [
                {
                    "expr": f"clamp_max(clamp_min((((tokenomics_wallet_liquidity_ratio or sovereignmap_wallet_liquidity_ratio or vector(0)) * 100) * {profile['wallet_liquidity_weight']}) + ((100 - ((tokenomics_top_10_holder_concentration or sovereignmap_top_10_holder_concentration or vector(0)) * 100)) * {profile['wallet_distribution_weight']}), 100), 0)",
                    "refId": "A",
                }
            ],
            panel_type="stat",
            unit="percent",
            decimals=1,
            thresholds=[(None, "red"), (50, "yellow"), (70, "green")],
        ),
    ]


tokenomics_panels = build_tokenomics_summary_panels("balanced") + apply_y_offset(
    tokenomics_section_panels, 6
)

performance_panels = [
    create_metric_panel(
        "Edge Compression Ratio",
        1,
        {"h": 8, "w": 12, "x": 0, "y": 0},
        [{"expr": "fl_compression_ratio", "refId": "A"}],
    ),
    create_metric_panel(
        "Self-Healing Backoff Events",
        2,
        {"h": 8, "w": 12, "x": 12, "y": 0},
        [{"expr": "fl_self_healing_events_total", "refId": "A"}],
    ),
    create_metric_panel(
        "Aggregator Bottlenecks (429/503)",
        3,
        {"h": 8, "w": 12, "x": 0, "y": 8},
        [{"expr": "fl_aggregator_throttles_total", "refId": "A"}],
    ),
]

os.makedirs("grafana/dashboards", exist_ok=True)
os.makedirs("grafana/provisioning/dashboards", exist_ok=True)

with open("grafana/dashboards/fl_tokenomics.json", "w", encoding="utf-8") as f:
    json.dump(
        create_dashboard("Tokenomics Overview", "tok-1", tokenomics_panels),
        f,
        indent=2,
    )

with open(
    "grafana/provisioning/dashboards/fl_tokenomics.json", "w", encoding="utf-8"
) as f:
    json.dump(
        create_dashboard("Tokenomics Overview", "tok-1", tokenomics_panels),
        f,
        indent=2,
    )

for profile_name in EXECUTIVE_SCORE_PROFILES:
    profile_uid = f"tok-1-{profile_name}"
    profile_title = f"Tokenomics Overview ({profile_name.title()} Profile)"
    profile_panels = build_tokenomics_summary_panels(profile_name) + apply_y_offset(
        tokenomics_section_panels, 6
    )

    with open(
        f"grafana/provisioning/dashboards/fl_tokenomics_{profile_name}.json",
        "w",
        encoding="utf-8",
    ) as f:
        json.dump(
            create_dashboard(profile_title, profile_uid, profile_panels),
            f,
            indent=2,
        )

    with open(
        f"grafana/dashboards/fl_tokenomics_{profile_name}.json",
        "w",
        encoding="utf-8",
    ) as f:
        json.dump(
            create_dashboard(profile_title, profile_uid, profile_panels),
            f,
            indent=2,
        )

with open("grafana/dashboards/fl_performance.json", "w", encoding="utf-8") as f:
    json.dump(
        create_dashboard("FL Performance", "perf-1", performance_panels),
        f,
        indent=2,
    )

print("Dashboards regenerated.")
