"""
Flask API for TPM-Attested BFT Testing
======================================
"""

from flask import Flask, jsonify, request
import json
import logging
from bft_with_tpm import BFTTestWithAttestations
import threading

app = Flask(__name__)

# Global test instance
current_test = None
test_thread = None
test_results = {'status': 'idle', 'results': [], 'summary': None}

@app.route('/health')
def health():
    return jsonify({"status": "healthy", "mode": "BFT_WITH_TPM"}), 200

@app.route('/start_bft_tpm_test', methods=['POST'])
def start_bft_tpm_test():
    """Start BFT test with TPM attestations"""
    global current_test, test_thread, test_results
    
    if test_thread and test_thread.is_alive():
        return jsonify({"error": "Test already running"}), 400
    
    test_results = {'status': 'running', 'results': [], 'summary': None}
    current_test = BFTTestWithAttestations()
    
    def run_test():
        global test_results
        try:
            results = current_test.run_full_test()
            test_results['results'] = [
                {k: v for k, v in r.items() if k not in ['accuracy_curve', 'loss_curve']}
                for r in results
            ]
            test_results['summary'] = current_test.get_summary()
            test_results['status'] = 'completed'
            
            # Save report
            report = current_test.generate_report()
            with open('BFT_TPM_TEST_RESULTS.md', 'w') as f:
                f.write(report)
                
        except Exception as e:
            test_results['status'] = 'error'
            test_results['error'] = str(e)
    
    test_thread = threading.Thread(target=run_test, daemon=True)
    test_thread.start()
    
    return jsonify({
        "status": "test_started",
        "configurations": 12,
        "rounds_per_config": 200,
        "total_rounds": 2400,
        "features": ["BFT", "TPM_ATTESTATIONS", "MULTI_KRUM_AGGREGATION"]
    })

@app.route('/bft_tpm_status', methods=['GET'])
def bft_tpm_status():
    """Get test status"""
    return jsonify(test_results)

@app.route('/bft_tpm_summary', methods=['GET'])
def bft_tpm_summary():
    """Get test summary with attestation stats"""
    if test_results['summary']:
        return jsonify(test_results['summary'])
    return jsonify({"error": "No summary available"}), 404

@app.route('/attestation_metrics', methods=['GET'])
def attestation_metrics():
    """Get attestation verification metrics"""
    if not current_test:
        return jsonify({"error": "No test running"}), 400
    
    return jsonify({
        'total_quotes_generated': current_test.metrics['total_quotes'],
        'verified_quotes': current_test.metrics['verified_quotes'],
        'failed_verifications': current_test.metrics['failed_verifications'],
        'verification_success_rate': (
            current_test.metrics['verified_quotes'] / current_test.metrics['total_quotes']
            if current_test.metrics['total_quotes'] > 0 else 0
        ),
        'nodes_total': current_test.NUM_NODES,
        'configurations': len(current_test.BYZANTINE_PERCENTAGES) * len(current_test.AGGREGATION_METHODS)
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=False)
