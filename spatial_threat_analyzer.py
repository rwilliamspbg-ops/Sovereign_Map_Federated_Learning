"""
Spatial Threat Analyzer: Gemini-Powered Byzantine Threat Detection
Real-time threat analysis and autonomous defense recommendations
"""

import json
import os
from typing import Dict, List, Optional
from dataclasses import dataclass
from enum import Enum
import asyncio


class ThreatLevel(Enum):
    SAFE = "safe"
    WARNING = "warning"
    ALERT = "alert"
    CRITICAL = "critical"


@dataclass
class BFTMetrics:
    """Byzantine Fault Tolerance metrics for analysis"""
    byzantine_percentage: float  # 0-100
    amplification_factor: float  # >1.0 = amplified
    recovery_time_rounds: int
    convergence_rate: float  # 0-100%
    active_node_count: int
    total_node_count: int
    attack_patterns: List[str]
    throughput: float  # updates/sec
    network_partition_detected: bool = False
    cascading_failure_detected: bool = False


@dataclass
class ThreatAnalysis:
    """Result of threat analysis"""
    threat_level: ThreatLevel
    severity_score: float  # 0-100
    risk_factors: List[str]
    immediate_actions: List[str]
    mitigation_strategy: str
    estimated_recovery_time: float
    confidence: float  # 0-100%
    recommended_defense: str
    ai_insights: str


class SpatialThreatAnalyzer:
    """Analyze Byzantine threats using Gemini AI"""

    def __init__(self, api_key: Optional[str] = None):
        """Initialize threat analyzer with Gemini API"""
        self.api_key = api_key or os.getenv("GEMINI_API_KEY")
        if not self.api_key:
            print("[WARNING] GEMINI_API_KEY not set. Using mock analysis.")
            self.mock_mode = True
        else:
            self.mock_mode = False
            try:
                import google.generativeai as genai

                genai.configure(api_key=self.api_key)
                self.model = genai.GenerativeModel("gemini-3-pro")
            except Exception as e:
                print(
                    f"[WARNING] Failed to initialize Gemini: {e}. Using mock mode."
                )
                self.mock_mode = True

        self.analysis_history: List[ThreatAnalysis] = []
        self.last_analysis: Optional[ThreatAnalysis] = None

    async def analyze_threats(self, metrics: BFTMetrics) -> ThreatAnalysis:
        """Analyze Byzantine threats and generate recommendations"""

        if self.mock_mode:
            return self._mock_analysis(metrics)

        try:
            return await self._gemini_analysis(metrics)
        except Exception as e:
            print(f"[WARNING] Gemini analysis failed: {e}. Using mock analysis.")
            return self._mock_analysis(metrics)

    async def _gemini_analysis(self, metrics: BFTMetrics) -> ThreatAnalysis:
        """Analyze using Gemini AI"""
        prompt = self._build_prompt(metrics)

        try:
            response = self.model.generate_content(prompt)
            analysis_text = response.text
            return self._parse_gemini_response(analysis_text, metrics)
        except Exception as e:
            print(f"[ERROR] Gemini API error: {e}")
            raise

    def _build_prompt(self, metrics: BFTMetrics) -> str:
        """Build analysis prompt for Gemini"""
        return f"""
Analyze this Byzantine Fault Tolerance (BFT) threat scenario and provide strategic recommendations:

CURRENT METRICS:
- Byzantine Nodes: {metrics.byzantine_percentage:.1f}% (Critical threshold: 50%)
- Amplification Factor: {metrics.amplification_factor:.2f}x (Threshold: 2.5x)
- Recovery Time: {metrics.recovery_time_rounds} rounds
- Convergence Rate: {metrics.convergence_rate:.1f}%
- Active Nodes: {metrics.active_node_count}/{metrics.total_node_count}
- Throughput: {metrics.throughput:.0f} updates/sec
- Attack Patterns: {', '.join(metrics.attack_patterns) if metrics.attack_patterns else 'None detected'}
- Network Partition: {'Yes' if metrics.network_partition_detected else 'No'}
- Cascading Failures: {'Yes' if metrics.cascading_failure_detected else 'No'}

TASK:
1. Assess threat level (safe/warning/alert/critical)
2. Calculate severity score (0-100)
3. Identify key risk factors
4. Recommend immediate actions
5. Suggest mitigation strategy
6. Estimate recovery time
7. Recommend automated defense protocol

RESPONSE FORMAT (JSON):
{{
    "threat_level": "critical|alert|warning|safe",
    "severity_score": 0-100,
    "risk_factors": ["factor1", "factor2", ...],
    "immediate_actions": ["action1", "action2", ...],
    "mitigation_strategy": "strategy description",
    "estimated_recovery_time": hours,
    "confidence": 0-100,
    "recommended_defense": "defense protocol name",
    "ai_insights": "detailed analysis explanation"
}}

Prioritize actionable, specific recommendations that can be executed autonomously.
"""

    def _parse_gemini_response(
        self, response_text: str, metrics: BFTMetrics
    ) -> ThreatAnalysis:
        """Parse Gemini response into ThreatAnalysis"""
        try:
            # Extract JSON from response
            json_str = response_text
            if "```json" in response_text:
                json_str = response_text.split("```json")[1].split("```")[0]
            elif "{" in response_text:
                # Try to extract JSON object
                start = response_text.find("{")
                end = response_text.rfind("}") + 1
                json_str = response_text[start:end]

            data = json.loads(json_str)

            return ThreatAnalysis(
                threat_level=ThreatLevel[data.get("threat_level", "alert").upper()],
                severity_score=float(data.get("severity_score", 50)),
                risk_factors=data.get("risk_factors", []),
                immediate_actions=data.get("immediate_actions", []),
                mitigation_strategy=data.get("mitigation_strategy", "Standard defense protocol"),
                estimated_recovery_time=float(data.get("estimated_recovery_time", 1.0)),
                confidence=float(data.get("confidence", 70)),
                recommended_defense=data.get("recommended_defense", "Hierarchical aggregation with trim"),
                ai_insights=data.get("ai_insights", "Analysis complete"),
            )
        except Exception as e:
            print(f"[WARNING] Failed to parse Gemini response: {e}")
            return self._mock_analysis(metrics)

    def _mock_analysis(self, metrics: BFTMetrics) -> ThreatAnalysis:
        """Generate mock analysis (for testing without Gemini)"""

        # Determine threat level
        if metrics.byzantine_percentage > 55:
            threat_level = ThreatLevel.CRITICAL
            severity_score = 90
        elif metrics.byzantine_percentage > 50:
            threat_level = ThreatLevel.ALERT
            severity_score = 75
        elif metrics.amplification_factor > 2.5:
            threat_level = ThreatLevel.ALERT
            severity_score = 70
        elif metrics.byzantine_percentage > 40:
            threat_level = ThreatLevel.WARNING
            severity_score = 50
        else:
            threat_level = ThreatLevel.SAFE
            severity_score = 20

        # Risk factors
        risk_factors = []
        if metrics.byzantine_percentage > 45:
            risk_factors.append("Byzantine nodes approaching critical threshold")
        if metrics.amplification_factor > 2.0:
            risk_factors.append("Coordinated Byzantine attack detected")
        if metrics.recovery_time_rounds > 10:
            risk_factors.append("Slow recovery from Byzantine perturbations")
        if metrics.network_partition_detected:
            risk_factors.append(
                "Network partition detected - Byzantine vectors multiplied"
            )
        if metrics.cascading_failure_detected:
            risk_factors.append("Cascading failures in progress")

        # Immediate actions
        immediate_actions = []
        if threat_level in [ThreatLevel.CRITICAL, ThreatLevel.ALERT]:
            immediate_actions.append("Activate hierarchical aggregation defense")
            immediate_actions.append("Increase Byzantine trim percentage to 20%")
            immediate_actions.append("Enable real-time Byzantine detection")
        if metrics.amplification_factor > 2.5:
            immediate_actions.append("Isolate suspected Byzantine nodes")
            immediate_actions.append("Log all Byzantine attack patterns")
        if metrics.network_partition_detected:
            immediate_actions.append("Activate partition recovery protocol")

        return ThreatAnalysis(
            threat_level=threat_level,
            severity_score=severity_score,
            risk_factors=risk_factors or ["No critical risk factors identified"],
            immediate_actions=immediate_actions or ["Continue monitoring"],
            mitigation_strategy="Use hierarchical aggregation with adaptive trim percentage based on Byzantine level",
            estimated_recovery_time=max(1.0, metrics.recovery_time_rounds * 0.5),
            confidence=85.0,
            recommended_defense="Hierarchical aggregation (26% faster, more stable)",
            ai_insights=f"System at {threat_level.value} threat level. Byzantine nodes: {metrics.byzantine_percentage:.1f}%. Amplification: {metrics.amplification_factor:.2f}x. Recovery time: {metrics.recovery_time_rounds} rounds.",
        )

    def get_defense_protocol(self, analysis: ThreatAnalysis) -> Dict:
        """Get automated defense protocol based on analysis"""

        protocol = {
            "protocol_id": f"defense_{analysis.threat_level.value}_{int(analysis.severity_score)}",
            "threat_level": analysis.threat_level.value,
            "activation": analysis.threat_level != ThreatLevel.SAFE,
            "parameters": {
                "aggregation_method": "hierarchical",
                "trim_percentage": min(0.25, 0.10 + (analysis.severity_score / 500)),
                "timeout_per_round": 10 + (analysis.severity_score / 5),
                "node_isolation_threshold": 2.5,
                "cascading_failure_detection": "enabled",
            },
            "actions": analysis.immediate_actions,
            "monitoring": {
                "amplification_factor_threshold": 2.5,
                "convergence_threshold": 50,
                "recovery_time_limit": 15,
            },
        }

        return protocol

    def store_analysis(self, analysis: ThreatAnalysis) -> None:
        """Store analysis for historical tracking"""
        self.analysis_history.append(analysis)
        self.last_analysis = analysis

    def get_threat_trend(self) -> Dict:
        """Analyze threat trend over time"""
        if len(self.analysis_history) < 2:
            return {"trend": "insufficient_data"}

        recent = self.analysis_history[-10:]
        severity_scores = [a.severity_score for a in recent]
        threat_levels = [a.threat_level.value for a in recent]

        return {
            "recent_analyses": len(recent),
            "avg_severity": sum(severity_scores) / len(severity_scores),
            "severity_trend": "increasing"
            if severity_scores[-1] > severity_scores[0]
            else "decreasing",
            "threat_levels": threat_levels,
            "latest_threat_level": threat_levels[-1],
            "max_severity": max(severity_scores),
            "min_severity": min(severity_scores),
        }


# Example usage
if __name__ == "__main__":
    import asyncio

    async def main():
        # Create analyzer
        analyzer = SpatialThreatAnalyzer()

        # Test metrics
        metrics = BFTMetrics(
            byzantine_percentage=52.0,
            amplification_factor=2.3,
            recovery_time_rounds=5,
            convergence_rate=74.0,
            active_node_count=98000,
            total_node_count=100000,
            attack_patterns=["coordinated_flip", "amplification"],
            throughput=74000,
            network_partition_detected=False,
            cascading_failure_detected=False,
        )

        # Analyze
        analysis = await analyzer.analyze_threats(metrics)
        analyzer.store_analysis(analysis)

        # Display results
        print(f"Threat Level: {analysis.threat_level.value}")
        print(f"Severity Score: {analysis.severity_score:.1f}/100")
        print(f"Confidence: {analysis.confidence:.1f}%")
        print(f"\nRisk Factors:")
        for factor in analysis.risk_factors:
            print(f"  - {factor}")
        print(f"\nImmediate Actions:")
        for action in analysis.immediate_actions:
            print(f"  - {action}")
        print(f"\nMitigation Strategy: {analysis.mitigation_strategy}")
        print(f"Recommended Defense: {analysis.recommended_defense}")

        # Get protocol
        protocol = analyzer.get_defense_protocol(analysis)
        print(f"\nDefense Protocol:")
        print(json.dumps(protocol, indent=2))

    asyncio.run(main())
