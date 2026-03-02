#!/usr/bin/env python3
"""
Generate comprehensive visual reports from Sovereign Map Demo results
Creates charts, statistics, and summary documents
"""

import json
import re
from datetime import datetime
from pathlib import Path
import statistics as stats

class DemoResultsReporter:
    def __init__(self, results_dir):
        self.results_dir = Path(results_dir)
        self.report_data = {
            'timestamp': datetime.now().isoformat(),
            'metrics_by_iteration': [],
            'container_stats': [],
            'performance_summary': {},
        }
    
    def parse_log(self):
        """Extract key information from demo log"""
        log_file = self.results_dir / "demo.log"
        if not log_file.exists():
            return None
        
        with open(log_file, 'r') as f:
            lines = f.readlines()
        
        summary = {
            'total_lines': len(lines),
            'start_time': None,
            'end_time': None,
            'key_events': []
        }
        
        for line in lines[:5]:
            if 'Nodes:' in line:
                match = re.search(r'Nodes: (\d+)', line)
                if match:
                    summary['nodes'] = int(match.group(1))
            if 'Duration:' in line:
                match = re.search(r'Duration: (\S+)', line)
                if match:
                    summary['duration'] = match.group(1)
        
        for line in lines:
            if '✅' in line or '❌' in line or 'ERROR' in line:
                summary['key_events'].append(line.strip())
        
        return summary
    
    def parse_metrics_iterations(self):
        """Parse all metrics iteration files"""
        metrics_files = sorted(self.results_dir.glob("metrics-iteration-*.txt"))
        
        iterations_data = []
        for mf in metrics_files:
            with open(mf, 'r') as f:
                content = f.read()
            
            iteration_num = int(mf.stem.split('-')[-1])
            
            metrics = {'iteration': iteration_num}
            
            # Extract running containers
            if 'Running:' in content:
                match = re.search(r'Running: (\d+), Total: (\d+)', content)
                if match:
                    metrics['running'] = int(match.group(1))
                    metrics['total'] = int(match.group(2))
            
            # Extract metric count
            if 'Available metrics:' in content:
                match = re.search(r'Available metrics: (\d+)', content)
                if match:
                    metrics['prometheus_metrics'] = int(match.group(1))
            
            iterations_data.append(metrics)
        
        return iterations_data
    
    def parse_final_state(self):
        """Parse final system state"""
        final_file = self.results_dir / "final-state.txt"
        if not final_file.exists():
            return None
        
        with open(final_file, 'r') as f:
            content = f.read()
        
        state = {
            'container_count': len(re.findall(r'sovereignmap-', content)),
            'timestamp': datetime.now().isoformat(),
        }
        
        return state
    
    def generate_report(self):
        """Generate comprehensive report"""
        log_summary = self.parse_log()
        metrics = self.parse_metrics_iterations()
        final_state = self.parse_final_state()
        
        report_md = self._build_markdown_report(log_summary, metrics, final_state)
        report_json = self._build_json_report(log_summary, metrics, final_state)
        
        return report_md, report_json
    
    def _build_markdown_report(self, log_summary, metrics, final_state):
        """Build markdown report"""
        
        md_lines = [
            "# Sovereign Map Federated Learning Demo - Results Report",
            "",
            f"**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
            "",
            "## Executive Summary",
            "",
        ]
        
        if log_summary:
            md_lines.extend([
                f"- **Nodes Deployed:** {log_summary.get('nodes', 'N/A')}",
                f"- **Duration:** {log_summary.get('duration', 'N/A')}",
                f"- **Total Log Entries:** {log_summary.get('total_lines', 0)}",
                "",
            ])
        
        md_lines.extend([
            "## Metrics Collection",
            "",
        ])
        
        if metrics:
            md_lines.append(f"**Total Iterations Collected:** {len(metrics)}")
            md_lines.append("")
            
            # Calculate statistics
            if metrics and all('running' in m for m in metrics):
                running_vals = [m['running'] for m in metrics]
                md_lines.extend([
                    "### Container Statistics",
                    f"- **Min Running:** {min(running_vals)}",
                    f"- **Max Running:** {max(running_vals)}",
                    f"- **Avg Running:** {stats.mean(running_vals):.1f}",
                    f"- **Final Running:** {metrics[-1].get('running', 'N/A')}",
                    "",
                ])
            
            # Prometheus metrics
            if metrics and any('prometheus_metrics' in m for m in metrics):
                prom_vals = [m.get('prometheus_metrics', 0) for m in metrics if 'prometheus_metrics' in m]
                if prom_vals:
                    md_lines.extend([
                        "### Prometheus Metrics",
                        f"- **Average Time Series:** {stats.mean(prom_vals):.0f}",
                        f"- **Peak:** {max(prom_vals)}",
                        "",
                    ])
        
        md_lines.extend([
            "## Key Events",
            "",
        ])
        
        if log_summary and log_summary.get('key_events'):
            for event in log_summary['key_events'][:10]:
                md_lines.append(f"- {event}")
            if len(log_summary['key_events']) > 10:
                md_lines.append(f"- ... and {len(log_summary['key_events']) - 10} more")
        
        md_lines.extend([
            "",
            "## Final System State",
            "",
        ])
        
        if final_state:
            md_lines.append(f"- **Containers:** {final_state.get('container_count', 'N/A')}")
            md_lines.append(f"- **State Snapshot:** {final_state.get('timestamp', 'N/A')}")
        
        md_lines.extend([
            "",
            "## Data Files",
            "",
            "- Metrics iterations: `metrics-iteration-N.txt`",
            "- Docker stats: `final-state.txt`",
            "- Service logs: `*-final.log`",
            "- This report: `RESULTS_REPORT.md` and `RESULTS_REPORT.json`",
            "",
            "---",
            f"Report generated at {datetime.now().isoformat()}",
        ])
        
        return "\n".join(md_lines)
    
    def _build_json_report(self, log_summary, metrics, final_state):
        """Build JSON report"""
        return {
            'generated': datetime.now().isoformat(),
            'summary': log_summary or {},
            'metrics_iterations': metrics,
            'final_state': final_state or {},
            'statistics': {
                'total_iterations': len(metrics),
                'avg_containers': stats.mean([m.get('running', 0) for m in metrics if 'running' in m]) if metrics else 0,
            } if metrics else {}
        }
    
    def save_reports(self):
        """Save reports to files"""
        report_md, report_json = self.generate_report()
        
        md_path = self.results_dir / "RESULTS_REPORT.md"
        json_path = self.results_dir / "RESULTS_REPORT.json"
        
        with open(md_path, 'w') as f:
            f.write(report_md)
        
        with open(json_path, 'w') as f:
            json.dump(report_json, f, indent=2)
        
        print(f"✅ Markdown report: {md_path}")
        print(f"✅ JSON report: {json_path}")
        print(f"\nReport summary:")
        print(report_md[:500] + "...")

if __name__ == "__main__":
    import sys
    
    # Find latest demo results
    base_dir = Path("Sovereign_Map_Federated_Learning/test-results/demo-windows")
    
    if not base_dir.exists():
        print(f"ERROR: Results directory not found: {base_dir}")
        sys.exit(1)
    
    dirs = sorted([d for d in base_dir.iterdir() if d.is_dir()])
    if not dirs:
        print("ERROR: No demo results found")
        sys.exit(1)
    
    latest_dir = dirs[-1]
    print(f"Analyzing results from: {latest_dir}")
    print()
    
    reporter = DemoResultsReporter(latest_dir)
    reporter.save_reports()
