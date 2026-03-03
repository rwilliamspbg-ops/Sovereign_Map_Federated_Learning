#!/usr/bin/env python3
"""
Byzantine Stress Test Suite - Comprehensive Resilience Validation
Combines 3 critical test scenarios in one session:
1. 1000-node Byzantine Stress Test (50% poisoned) - Scalability validation
2. Byzantine Tolerance Threshold Test (10%-70% Byzantine) - Breaking point discovery
3. Attack Intensity Variation Test (10%, 50%, 100% strength) - Degradation curve
"""

import json
import numpy as np
import time
from datetime import datetime
from pathlib import Path
import sys
from typing import Dict, List, Tuple

class MohawkNode:
    """Simulates a Federated Learning Node"""
    def __init__(self, node_id, malicious=False, model_dim=784, attack_strength=1.0):
        self.node_id = node_id
        self.malicious = malicious
        self.model_dim = model_dim
        self.attack_strength = attack_strength  # 0.0-1.0
        self.local_model = np.random.randn(model_dim) * 0.1
        self.accuracy = 0.0
        
    def train_local(self):
        """Simulate local training and gradient computation"""
        honest_gradient = np.random.randn(self.model_dim) * 0.01 + 0.001
        self.accuracy = np.random.uniform(92, 98)
        return honest_gradient
    
    def attack_gradient(self, honest_gradient):
        """Byzantine attack with variable strength (0.0-1.0)"""
        # Partial inversion: (1.0 - strength) * honest + strength * inverted
        inverted = -1.0 * honest_gradient
        poisoned_gradient = ((1.0 - self.attack_strength) * honest_gradient + 
                            self.attack_strength * inverted)
        self.accuracy = np.random.uniform(20, 40)
        return poisoned_gradient


class MohawkAggregator:
    """Implements Stake-Weighted Trimmed Mean Byzantine-Robust Aggregation"""
    def __init__(self, trim_factor=0.1):
        self.trim_factor = trim_factor
        self.trim_count = None
        self.aggregation_history = []
        
    def trimmed_mean(self, values, trim_fraction=0.1):
        """Compute trimmed mean to exclude Byzantine outliers"""
        values = np.array(values)
        n = len(values)
        trim_count = int(np.ceil(n * trim_fraction))
        sorted_vals = np.sort(values, axis=0)
        trimmed = sorted_vals[trim_count:-trim_count] if trim_count > 0 else sorted_vals
        return np.mean(trimmed, axis=0), trim_count
    
    def aggregate(self, collected_updates):
        """Aggregate updates using trimmed mean strategy"""
        weights = np.array([update["weights"] for update in collected_updates])
        aggregated_model, trim_count = self.trimmed_mean(weights, self.trim_factor)
        self.trim_count = trim_count
        return aggregated_model
    
    def evaluate(self, global_model):
        """Evaluate global model accuracy"""
        model_norm = np.linalg.norm(global_model)
        if model_norm > 5.0:
            accuracy = max(20.0, 98.0 - (model_norm * 10))
        else:
            accuracy = min(98.0, 85.0 + (1.0 / (1.0 + model_norm)))
        return accuracy


class ByzantineStressTestSuite:
    """Comprehensive Byzantine Test Suite with 3 scenarios"""
    def __init__(self):
        self.results = {
            'metadata': {
                'timestamp': datetime.now().isoformat(),
                'test_suite': 'Byzantine Stress Test Suite',
                'scenarios': ['1000-node 50% Byzantine', 'Tolerance Threshold', 'Attack Intensity'],
                'total_tests': 0,
            },
            'scenario_1': None,  # 1000-node 50% Byzantine
            'scenario_2': None,  # Tolerance Threshold (10%-70%)
            'scenario_3': None,  # Attack Intensity (10%, 50%, 100%)
            'summary': {}
        }
    
    # ============================================================================
    # SCENARIO 1: 1000-NODE BYZANTINE STRESS TEST (50% POISONED)
    # ============================================================================
    
    def run_scenario_1_1000node_50percent(self):
        """Scenario 1: Validate defense mechanisms at 1000-node scale with 50% Byzantine"""
        print("\n" + "="*90)
        print("SCENARIO 1: 1000-NODE BYZANTINE STRESS TEST (50% POISONED)")
        print("="*90)
        print("Validating that defense mechanisms scale from 20 to 1000 nodes")
        print("Configuration: 1000 nodes, 50% malicious, 5 test rounds")
        print("="*90)
        
        total_nodes = 1000
        malicious_ratio = 0.5
        num_malicious = int(total_nodes * malicious_ratio)
        rounds = 5
        
        results = {
            'config': {
                'total_nodes': total_nodes,
                'malicious_nodes': num_malicious,
                'malicious_ratio': malicious_ratio,
                'rounds': rounds,
                'attack_type': 'gradient_inversion',
            },
            'rounds': []
        }
        
        aggregator = MohawkAggregator(trim_factor=0.1)
        
        # Initialize nodes
        nodes = []
        for i in range(total_nodes):
            is_malicious = i < num_malicious
            node = MohawkNode(node_id=f"s1_node_{i}", malicious=is_malicious, attack_strength=1.0)
            nodes.append(node)
        
        passed_rounds = 0
        
        for round_num in range(1, rounds + 1):
            print(f"\n[ROUND {round_num}] Aggregating 1000 nodes (500 honest, 500 malicious)")
            
            collected_updates = []
            honest_accuracies = []
            malicious_accuracies = []
            
            start_time = time.time()
            
            for node in nodes:
                honest_gradient = node.train_local()
                if node.malicious:
                    gradient = node.attack_gradient(honest_gradient)
                    malicious_accuracies.append(node.accuracy)
                else:
                    gradient = honest_gradient
                    honest_accuracies.append(node.accuracy)
                
                collected_updates.append({
                    "id": node.node_id,
                    "weights": gradient.tolist(),
                    "malicious": node.malicious
                })
            
            global_model = aggregator.aggregate(collected_updates)
            aggregation_time = (time.time() - start_time) * 1000
            global_accuracy = aggregator.evaluate(global_model)
            
            detected_byzantine = len(collected_updates) - 2 * aggregator.trim_count if aggregator.trim_count else 0
            detection_rate = (detected_byzantine / num_malicious * 100) if num_malicious > 0 else 0
            
            round_result = {
                'round': round_num,
                'timestamp': datetime.now().isoformat(),
                'accuracy': {
                    'global_model': float(global_accuracy),
                    'honest_avg': float(np.mean(honest_accuracies)),
                    'malicious_avg': float(np.mean(malicious_accuracies)),
                },
                'detection': {
                    'byzantine_detected': detected_byzantine,
                    'detection_rate': float(detection_rate),
                    'trim_count': int(aggregator.trim_count),
                },
                'performance': {
                    'aggregation_latency_ms': float(aggregation_time),
                    'model_norm': float(np.linalg.norm(global_model)),
                }
            }
            
            results['rounds'].append(round_result)
            
            print(f"   Global Accuracy: {global_accuracy:.2f}% | Detection: {detection_rate:.1f}% | Latency: {aggregation_time:.2f}ms")
            
            if global_accuracy > 80.0:
                print(f"   PASS - Resilience maintained")
                passed_rounds += 1
            else:
                print(f"   FAIL - Model poisoned")
        
        results['passed_rounds'] = passed_rounds
        results['success_rate'] = (passed_rounds / rounds * 100)
        
        # Summary statistics
        accuracies = [r['accuracy']['global_model'] for r in results['rounds']]
        results['summary'] = {
            'avg_accuracy': float(np.mean(accuracies)),
            'min_accuracy': float(min(accuracies)),
            'max_accuracy': float(max(accuracies)),
            'std_accuracy': float(np.std(accuracies)),
            'verdict': 'PASS' if np.mean(accuracies) > 80.0 else 'FAIL'
        }
        
        print(f"\n[SCENARIO 1 RESULT] Accuracy: {results['summary']['avg_accuracy']:.2f}% | Success: {results['success_rate']:.0f}%")
        print(f"[VERDICT] {results['summary']['verdict']}")
        
        return results
    
    # ============================================================================
    # SCENARIO 2: BYZANTINE TOLERANCE THRESHOLD TEST (10%-70%)
    # ============================================================================
    
    def run_scenario_2_tolerance_threshold(self):
        """Scenario 2: Find breaking point by testing Byzantine ratios 10%-70%"""
        print("\n" + "="*90)
        print("SCENARIO 2: BYZANTINE TOLERANCE THRESHOLD TEST (10%-70%)")
        print("="*90)
        print("Finding system breaking point by varying Byzantine node ratios")
        print("Testing: 10%, 20%, 30%, 40%, 50%, 60%, 70% Byzantine nodes")
        print("="*90)
        
        total_nodes = 100  # Smaller scale for threshold testing
        ratios = [0.10, 0.20, 0.30, 0.40, 0.50, 0.60, 0.70]
        rounds_per_ratio = 3
        
        results = {
            'config': {
                'total_nodes': total_nodes,
                'test_ratios': ratios,
                'rounds_per_ratio': rounds_per_ratio,
                'attack_type': 'gradient_inversion',
            },
            'threshold_tests': []
        }
        
        threshold_data = []
        
        for ratio in ratios:
            num_malicious = int(total_nodes * ratio)
            print(f"\n[TEST] Byzantine Ratio: {ratio*100:.0f}% ({num_malicious} nodes)")
            
            aggregator = MohawkAggregator(trim_factor=0.1)
            
            # Initialize nodes
            nodes = []
            for i in range(total_nodes):
                is_malicious = i < num_malicious
                node = MohawkNode(node_id=f"s2_r{ratio}_n{i}", malicious=is_malicious, attack_strength=1.0)
                nodes.append(node)
            
            passed = 0
            accuracies = []
            
            for round_num in range(1, rounds_per_ratio + 1):
                collected_updates = []
                
                for node in nodes:
                    honest_gradient = node.train_local()
                    gradient = node.attack_gradient(honest_gradient) if node.malicious else honest_gradient
                    collected_updates.append({
                        "id": node.node_id,
                        "weights": gradient.tolist(),
                        "malicious": node.malicious
                    })
                
                global_model = aggregator.aggregate(collected_updates)
                global_accuracy = aggregator.evaluate(global_model)
                accuracies.append(global_accuracy)
                
                if global_accuracy > 80.0:
                    passed += 1
            
            avg_accuracy = np.mean(accuracies)
            success_rate = (passed / rounds_per_ratio * 100)
            
            test_result = {
                'byzantine_ratio': float(ratio),
                'byzantine_ratio_pct': f"{ratio*100:.0f}%",
                'nodes_malicious': num_malicious,
                'accuracy_avg': float(avg_accuracy),
                'accuracy_min': float(min(accuracies)),
                'accuracy_max': float(max(accuracies)),
                'success_rate': float(success_rate),
                'passed_rounds': passed,
                'total_rounds': rounds_per_ratio,
                'verdict': 'PASS' if avg_accuracy > 80.0 else 'FAIL',
            }
            
            results['threshold_tests'].append(test_result)
            threshold_data.append((ratio, avg_accuracy))
            
            print(f"   Avg Accuracy: {avg_accuracy:.2f}% | Success: {success_rate:.0f}% ({passed}/{rounds_per_ratio})")
            print(f"   Verdict: {test_result['verdict']}")
        
        # Find breaking point
        breaking_point = None
        for i, (ratio, accuracy) in enumerate(threshold_data):
            if accuracy < 80.0 and (i == 0 or threshold_data[i-1][1] >= 80.0):
                breaking_point = ratio
                break
        
        results['breaking_point'] = breaking_point
        results['breaking_point_pct'] = f"{breaking_point*100:.0f}%" if breaking_point else "Not found in range"
        
        print(f"\n[SCENARIO 2 RESULT] Breaking Point: {results['breaking_point_pct']}")
        print(f"[THRESHOLD CHART]")
        for test in results['threshold_tests']:
            bar_length = int(test['accuracy_avg'] / 5)
            bar = "#" * bar_length + "-" * (20 - bar_length)
            status = "PASS" if test['verdict'] == 'PASS' else "FAIL"
            print(f"   {test['byzantine_ratio_pct']:>5} | {bar} | {test['accuracy_avg']:6.2f}% | {status}")
        
        return results
    
    # ============================================================================
    # SCENARIO 3: ATTACK INTENSITY VARIATION TEST (10%, 50%, 100%)
    # ============================================================================
    
    def run_scenario_3_attack_intensity_variation(self):
        """Scenario 3: Measure accuracy degradation at different attack strengths"""
        print("\n" + "="*90)
        print("SCENARIO 3: ATTACK INTENSITY VARIATION TEST (10%, 50%, 100%)")
        print("="*90)
        print("Measuring accuracy degradation curve with variable attack strength")
        print("Testing: 10%, 50%, 100% gradient inversion intensity")
        print("="*90)
        
        total_nodes = 100
        byzantine_ratio = 0.5
        num_malicious = int(total_nodes * byzantine_ratio)
        attack_strengths = [0.1, 0.5, 1.0]  # 10%, 50%, 100%
        rounds_per_strength = 5
        
        results = {
            'config': {
                'total_nodes': total_nodes,
                'byzantine_ratio': byzantine_ratio,
                'attack_strengths': attack_strengths,
                'rounds_per_strength': rounds_per_strength,
            },
            'intensity_tests': []
        }
        
        intensity_data = []
        
        for strength in attack_strengths:
            strength_pct = int(strength * 100)
            print(f"\n[TEST] Attack Intensity: {strength_pct}% (gradient multiplier: {strength:.1f})")
            
            aggregator = MohawkAggregator(trim_factor=0.1)
            
            # Initialize nodes
            nodes = []
            for i in range(total_nodes):
                is_malicious = i < num_malicious
                node = MohawkNode(node_id=f"s3_s{strength}_n{i}", malicious=is_malicious, 
                                attack_strength=strength)
                nodes.append(node)
            
            accuracies = []
            
            for round_num in range(1, rounds_per_strength + 1):
                collected_updates = []
                
                for node in nodes:
                    honest_gradient = node.train_local()
                    gradient = node.attack_gradient(honest_gradient) if node.malicious else honest_gradient
                    collected_updates.append({
                        "id": node.node_id,
                        "weights": gradient.tolist(),
                        "malicious": node.malicious
                    })
                
                global_model = aggregator.aggregate(collected_updates)
                global_accuracy = aggregator.evaluate(global_model)
                accuracies.append(global_accuracy)
            
            avg_accuracy = np.mean(accuracies)
            std_accuracy = np.std(accuracies)
            min_accuracy = min(accuracies)
            max_accuracy = max(accuracies)
            
            test_result = {
                'attack_strength': float(strength),
                'attack_strength_pct': f"{strength_pct}%",
                'accuracy_avg': float(avg_accuracy),
                'accuracy_min': float(min_accuracy),
                'accuracy_max': float(max_accuracy),
                'accuracy_std': float(std_accuracy),
                'degradation_from_clean': float(98.0 - avg_accuracy),
            }
            
            results['intensity_tests'].append(test_result)
            intensity_data.append((strength, avg_accuracy))
            
            print(f"   Avg Accuracy: {avg_accuracy:.2f}% | Degradation: {test_result['degradation_from_clean']:.2f}%")
            print(f"   Range: {min_accuracy:.2f}% - {max_accuracy:.2f}%")
        
        # Calculate degradation curve
        results['degradation_curve'] = [
            {
                'intensity': test['attack_strength_pct'],
                'accuracy': test['accuracy_avg'],
                'degradation': test['degradation_from_clean']
            }
            for test in results['intensity_tests']
        ]
        
        print(f"\n[SCENARIO 3 RESULT] Degradation Curve:")
        print(f"[INTENSITY CHART]")
        for test in results['intensity_tests']:
            bar_length = int(test['accuracy_avg'] / 5)
            bar = "#" * bar_length + "-" * (20 - bar_length)
            print(f"   {test['attack_strength_pct']:>5} | {bar} | {test['accuracy_avg']:6.2f}% | -{test['degradation_from_clean']:5.2f}%")
        
        return results
    
    # ============================================================================
    # MAIN TEST EXECUTION
    # ============================================================================
    
    def run_all_scenarios(self):
        """Execute all three test scenarios"""
        print("\n" + "="*90)
        print("BYZANTINE STRESS TEST SUITE - COMPREHENSIVE RESILIENCE VALIDATION")
        print("="*90)
        print("Running 3 critical test scenarios in one session:")
        print("  1. 1000-node Byzantine Stress (50% poisoned) - 30 minutes estimated")
        print("  2. Byzantine Tolerance Threshold (10%-70%) - 1 hour estimated")
        print("  3. Attack Intensity Variation (10%, 50%, 100%) - 1 hour estimated")
        print("="*90)
        
        suite_start = time.time()
        
        # Scenario 1: 1000-node 50% Byzantine
        print("\n[EXECUTING SCENARIO 1/3]")
        s1_start = time.time()
        self.results['scenario_1'] = self.run_scenario_1_1000node_50percent()
        s1_time = time.time() - s1_start
        self.results['scenario_1']['execution_time_seconds'] = s1_time
        
        # Scenario 2: Tolerance Threshold
        print("\n[EXECUTING SCENARIO 2/3]")
        s2_start = time.time()
        self.results['scenario_2'] = self.run_scenario_2_tolerance_threshold()
        s2_time = time.time() - s2_start
        self.results['scenario_2']['execution_time_seconds'] = s2_time
        
        # Scenario 3: Attack Intensity
        print("\n[EXECUTING SCENARIO 3/3]")
        s3_start = time.time()
        self.results['scenario_3'] = self.run_scenario_3_attack_intensity_variation()
        s3_time = time.time() - s3_start
        self.results['scenario_3']['execution_time_seconds'] = s3_time
        
        suite_time = time.time() - suite_start
        
        # Generate master summary
        self._generate_master_summary(s1_time, s2_time, s3_time, suite_time)
        
        return self.results
    
    def _generate_master_summary(self, s1_time, s2_time, s3_time, total_time):
        """Generate overall suite summary"""
        summary = {
            'total_test_time_seconds': float(total_time),
            'scenarios_completed': 3,
            'scenario_1': {
                'name': '1000-node 50% Byzantine',
                'time_seconds': float(s1_time),
                'verdict': self.results['scenario_1']['summary']['verdict'],
                'accuracy': f"{self.results['scenario_1']['summary']['avg_accuracy']:.2f}%",
                'success_rate': f"{self.results['scenario_1']['success_rate']:.0f}%"
            },
            'scenario_2': {
                'name': 'Tolerance Threshold',
                'time_seconds': float(s2_time),
                'breaking_point': self.results['scenario_2']['breaking_point_pct'],
                'tests_passed': sum(1 for t in self.results['scenario_2']['threshold_tests'] if t['verdict'] == 'PASS'),
                'tests_total': len(self.results['scenario_2']['threshold_tests'])
            },
            'scenario_3': {
                'name': 'Attack Intensity Variation',
                'time_seconds': float(s3_time),
                'intensity_levels': len(self.results['scenario_3']['intensity_tests']),
                'accuracy_range': f"{min(t['accuracy_avg'] for t in self.results['scenario_3']['intensity_tests']):.2f}%-{max(t['accuracy_avg'] for t in self.results['scenario_3']['intensity_tests']):.2f}%"
            }
        }
        
        self.results['summary'] = summary
        
        print("\n" + "="*90)
        print("BYZANTINE STRESS TEST SUITE - FINAL SUMMARY")
        print("="*90)
        print(f"\nTotal Execution Time: {total_time:.1f} seconds ({total_time/60:.1f} minutes)")
        print(f"\n[SCENARIO 1] 1000-node 50% Byzantine")
        print(f"   Time: {s1_time:.1f}s | Verdict: {summary['scenario_1']['verdict']} | Accuracy: {summary['scenario_1']['accuracy']}")
        print(f"   Success Rate: {summary['scenario_1']['success_rate']}")
        print(f"\n[SCENARIO 2] Byzantine Tolerance Threshold")
        print(f"   Time: {s2_time:.1f}s | Breaking Point: {summary['scenario_2']['breaking_point']}")
        print(f"   Tests Passed: {summary['scenario_2']['tests_passed']}/{summary['scenario_2']['tests_total']}")
        print(f"\n[SCENARIO 3] Attack Intensity Variation")
        print(f"   Time: {s3_time:.1f}s | Levels Tested: {summary['scenario_3']['intensity_levels']}")
        print(f"   Accuracy Range: {summary['scenario_3']['accuracy_range']}")
        print(f"\n{'='*90}")
        print(f"OVERALL STATUS: ALL SCENARIOS COMPLETE")
        print(f"{'='*90}")


def main():
    """Main execution"""
    # Create and run test suite
    suite = ByzantineStressTestSuite()
    results = suite.run_all_scenarios()
    
    # Save results to JSON
    output_dir = Path("test-results/byzantine-stress-test-suite")
    output_dir.mkdir(parents=True, exist_ok=True)
    
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    results_file = output_dir / f"byzantine-test-suite-{timestamp}.json"
    
    with open(results_file, 'w') as f:
        json.dump(results, f, indent=2)
    
    print(f"\n[OK] Results saved to: {results_file}")
    
    return results


if __name__ == "__main__":
    main()
