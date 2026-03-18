"""
FL Metrics Translator: Convert FL metrics to Spatial Coordinates
Maps Byzantine Fault Tolerance metrics to 3D visualization space
"""

import numpy as np
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
from enum import Enum
import json
from datetime import datetime


class NodeStatus(Enum):
    HEALTHY = "healthy"
    DEGRADED = "degraded"
    CRITICAL = "critical"
    FAILED = "failed"


@dataclass
class FLMetric:
    """Federated Learning metric from a single node"""

    node_id: int
    accuracy: float  # 0-100%
    byzantine_level: float  # 0-100%
    convergence: float  # 0-100%
    throughput: float  # updates/sec
    recovery_time: float  # rounds
    amplification_factor: float
    timestamp: datetime
    active: bool = True


@dataclass
class SpatialCoordinate:
    """3D spatial coordinate for node visualization"""

    node_id: int
    x: float
    y: float
    z: float
    color: str  # RGB hex
    size: float  # relative size (0.1-2.0)
    intensity: float  # pulsing intensity (0.0-1.0)
    status: NodeStatus
    accuracy: float
    byzantine_level: float
    convergence: float
    amplification_factor: float


class HilbertCurve:
    """Hilbert curve for locality-preserving 3D mapping"""

    @staticmethod
    def d2xy(n: int, d: int) -> Tuple[int, int]:
        """Convert distance along Hilbert curve to (x,y) coordinates"""
        x = y = 0
        s = 1
        while s < n:
            rx = 1 & (d >> 1)
            ry = 1 & (d ^ rx)
            if ry == 0:
                if rx == 1:
                    x = s - 1 - x
                    y = s - 1 - y
                x, y = y, x
            x += s * rx
            y += s * ry
            d >>= 2
            s <<= 1
        return x, y

    @staticmethod
    def xy2d(n: int, x: int, y: int) -> int:
        """Convert (x,y) coordinates to distance along Hilbert curve"""
        d = 0
        s = n // 2
        while s > 0:
            rx = 1 if (x & s) > 0 else 0
            ry = 1 if (y & s) > 0 else 0
            d += s * s * ((3 * rx) ^ ry)
            if ry == 0:
                if rx == 1:
                    x = n - 1 - x
                    y = n - 1 - y
                x, y = y, x
            s //= 2
        return d


class FLMetricsTranslator:
    """Translate FL metrics to 3D spatial coordinates"""

    def __init__(self, num_nodes: int = 100000, grid_size: int = 512):
        self.num_nodes = num_nodes
        self.grid_size = grid_size
        self.scale = 100.0  # 3D space scale factor
        self.metrics_cache: Dict[int, FLMetric] = {}
        self.coordinates_cache: Dict[int, SpatialCoordinate] = {}

    def update_metric(self, metric: FLMetric) -> None:
        """Update metric for a node"""
        self.metrics_cache[metric.node_id] = metric

    def get_node_coordinates(self, node_id: int) -> Optional[SpatialCoordinate]:
        """Get 3D coordinates for a node"""
        if node_id not in self.metrics_cache:
            return None

        metric = self.metrics_cache[node_id]

        # Use Hilbert curve for locality-preserving mapping
        hilbert_index = node_id % (self.grid_size * self.grid_size)
        x_grid, y_grid = HilbertCurve.d2xy(self.grid_size, hilbert_index)

        # Map grid coordinates to 3D space
        x = (x_grid / self.grid_size - 0.5) * self.scale
        y = (metric.convergence / 100.0 - 0.5) * self.scale  # Convergence = height
        z = (y_grid / self.grid_size - 0.5) * self.scale

        # Determine node color based on Byzantine level
        color = self._get_node_color(metric.byzantine_level)

        # Determine node status
        status = self._get_node_status(metric)

        # Calculate size based on throughput (normalized 0-1)
        size = 0.5 + (metric.throughput / 100000.0) * 1.5  # 0.5 to 2.0
        size = max(0.1, min(2.0, size))

        # Calculate pulsing intensity based on convergence
        intensity = metric.convergence / 100.0

        return SpatialCoordinate(
            node_id=node_id,
            x=x,
            y=y,
            z=z,
            color=color,
            size=size,
            intensity=intensity,
            status=status,
            accuracy=metric.accuracy,
            byzantine_level=metric.byzantine_level,
            convergence=metric.convergence,
            amplification_factor=metric.amplification_factor,
        )

    def get_all_coordinates(self) -> List[SpatialCoordinate]:
        """Get coordinates for all active nodes"""
        coordinates = []
        for node_id in self.metrics_cache.keys():
            coord = self.get_node_coordinates(node_id)
            if coord:
                coordinates.append(coord)
        return coordinates

    def _get_node_color(self, byzantine_level: float) -> str:
        """Determine node color based on Byzantine threat level"""
        if byzantine_level < 10:
            return "#00ff00"  # Green: Safe
        elif byzantine_level < 30:
            return "#ffff00"  # Yellow: Warning
        elif byzantine_level < 50:
            return "#ff8800"  # Orange: Alert
        else:
            return "#ff0000"  # Red: Critical

    def _get_node_status(self, metric: FLMetric) -> NodeStatus:
        """Determine node status"""
        if not metric.active:
            return NodeStatus.FAILED

        if metric.convergence < 50:
            return NodeStatus.CRITICAL
        elif metric.convergence < 70:
            return NodeStatus.DEGRADED
        else:
            return NodeStatus.HEALTHY

    def export_to_prometheus_format(self) -> str:
        """Export metrics in Prometheus format"""
        lines = []
        lines.append("# HELP sovereign_fl_node_position_3d 3D spatial coordinates")
        lines.append("# TYPE sovereign_fl_node_position_3d gauge")

        for node_id, coord in enumerate(self.get_all_coordinates()):
            labels = f'node_id="{coord.node_id}",status="{coord.status.value}"'
            lines.append(f"sovereign_fl_node_x{{{labels}}} {coord.x}")
            lines.append(f"sovereign_fl_node_y{{{labels}}} {coord.y}")
            lines.append(f"sovereign_fl_node_z{{{labels}}} {coord.z}")
            lines.append(f"sovereign_fl_node_size{{{labels}}} {coord.size}")
            lines.append(f"sovereign_fl_node_intensity{{{labels}}} {coord.intensity}")

        return "\n".join(lines)

    def export_to_json(self) -> str:
        """Export all coordinates as JSON"""
        coordinates = self.get_all_coordinates()
        data = {
            "timestamp": datetime.now().isoformat(),
            "node_count": len(coordinates),
            "nodes": [
                {
                    "id": coord.node_id,
                    "position": {"x": coord.x, "y": coord.y, "z": coord.z},
                    "visual": {
                        "color": coord.color,
                        "size": coord.size,
                        "intensity": coord.intensity,
                    },
                    "metrics": {
                        "accuracy": coord.accuracy,
                        "byzantine_level": coord.byzantine_level,
                        "convergence": coord.convergence,
                        "amplification_factor": coord.amplification_factor,
                    },
                    "status": coord.status.value,
                }
                for coord in coordinates
            ],
        }
        return json.dumps(data, indent=2)

    def get_aggregated_stats(self) -> Dict:
        """Get aggregated statistics for all nodes"""
        if not self.metrics_cache:
            return {}

        accuracies = [m.accuracy for m in self.metrics_cache.values()]
        byzantine_levels = [m.byzantine_level for m in self.metrics_cache.values()]
        convergences = [m.convergence for m in self.metrics_cache.values()]
        amplifications = [m.amplification_factor for m in self.metrics_cache.values()]

        return {
            "total_nodes": len(self.metrics_cache),
            "accuracy": {
                "mean": np.mean(accuracies),
                "min": np.min(accuracies),
                "max": np.max(accuracies),
                "std": np.std(accuracies),
            },
            "byzantine_level": {
                "mean": np.mean(byzantine_levels),
                "min": np.min(byzantine_levels),
                "max": np.max(byzantine_levels),
            },
            "convergence": {
                "mean": np.mean(convergences),
                "min": np.min(convergences),
                "max": np.max(convergences),
            },
            "amplification_factor": {
                "mean": np.mean(amplifications),
                "max": np.max(amplifications),
            },
            "node_status": {
                "healthy": sum(
                    1 for m in self.metrics_cache.values() if m.convergence >= 70
                ),
                "degraded": sum(
                    1 for m in self.metrics_cache.values() if 50 <= m.convergence < 70
                ),
                "critical": sum(
                    1 for m in self.metrics_cache.values() if m.convergence < 50
                ),
                "failed": sum(1 for m in self.metrics_cache.values() if not m.active),
            },
        }


# Example usage
if __name__ == "__main__":
    # Create translator for 100K nodes
    translator = FLMetricsTranslator(num_nodes=100000)

    # Simulate metrics from 100 nodes
    for node_id in range(100):
        metric = FLMetric(
            node_id=node_id,
            accuracy=90 + np.random.randn() * 5,
            byzantine_level=np.random.uniform(0, 50),
            convergence=80 + np.random.randn() * 10,
            throughput=80000 + np.random.randn() * 5000,
            recovery_time=4 + np.random.randn(),
            amplification_factor=1.5 + np.random.uniform(0, 2),
            timestamp=datetime.now(),
            active=np.random.random() > 0.05,
        )
        translator.update_metric(metric)

    # Export and display stats
    stats = translator.get_aggregated_stats()
    print("Aggregated Statistics:")
    print(json.dumps(stats, indent=2))

    # Export JSON
    print("\nJSON Export (first 500 chars):")
    print(translator.export_to_json()[:500])
