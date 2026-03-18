"""
Phase 3D Backend API - Serves real federated training metrics to frontend
Provides endpoints for training management and real-time metric streaming
"""

from flask import Flask, jsonify, request
from flask_cors import CORS
from threading import Thread, Lock
import json
from pathlib import Path
from datetime import datetime
import logging
from phase3d_training import FederatedLearningTrainer, TrainingConfig

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

# Global state
training_state = {
    'status': 'idle',  # idle, training, completed, error
    'current_round': 0,
    'total_rounds': 0,
    'metrics_history': [],
    'current_metrics': None,
    'error': None,
    'trainer': None,
    'config': TrainingConfig(),
    'lock': Lock()
}

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'timestamp': datetime.utcnow().isoformat(),
        'backend_version': '3.0.0'
    })

@app.route('/training/config', methods=['GET', 'POST'])
def training_config():
    """Get or create training configuration"""
    with training_state['lock']:
        if request.method == 'POST':
            config_data = request.get_json() or {}
            try:
                config = TrainingConfig(**config_data)
                training_state['config'] = config
                return jsonify({
                    'status': 'success',
                    'config': {
                        'num_rounds': config.num_rounds,
                        'num_clients': config.num_clients,
                        'local_epochs': config.local_epochs,
                        'batch_size': config.batch_size,
                        'learning_rate': config.learning_rate,
                        'epsilon': config.epsilon,
                        'compression_bits': config.compression_bits,
                        'dataset': config.dataset,
                        'device': config.device,
                        'multi_gpu': config.multi_gpu
                    }
                })
            except Exception as e:
                return jsonify({'status': 'error', 'message': str(e)}), 400
        
        # GET: Return current config or defaults
        config = training_state.get('config', TrainingConfig())
        return jsonify({
            'num_rounds': config.num_rounds,
            'num_clients': config.num_clients,
            'local_epochs': config.local_epochs,
            'batch_size': config.batch_size,
            'learning_rate': config.learning_rate,
            'epsilon': config.epsilon,
            'compression_bits': config.compression_bits,
            'dataset': config.dataset,
            'device': config.device,
            'multi_gpu': config.multi_gpu,
            'use_compression': config.use_compression,
            'use_privacy': config.use_privacy
        })

@app.route('/training/start', methods=['POST'])
def start_training():
    """Start a federated learning training session"""
    with training_state['lock']:
        if training_state['status'] == 'training':
            return jsonify({
                'status': 'error',
                'message': 'Training already in progress'
            }), 400
        
        # Get config from request or use defaults
        config_data = request.get_json() or {}
        try:
            config = TrainingConfig(**config_data)
        except Exception as e:
            return jsonify({'status': 'error', 'message': str(e)}), 400
        
        # Initialize trainer and state
        training_state['trainer'] = FederatedLearningTrainer(config)
        training_state['config'] = config
        training_state['status'] = 'training'
        training_state['current_round'] = 0
        training_state['total_rounds'] = config.num_rounds
        training_state['metrics_history'] = []
        training_state['error'] = None
    
    # Start training in background thread
    thread = Thread(target=_run_training)
    thread.daemon = True
    thread.start()
    
    return jsonify({
        'status': 'started',
        'message': 'Training session initiated',
        'total_rounds': training_state['total_rounds']
    })

def _run_training():
    """Background task for federated learning training"""
    try:
        trainer = training_state['trainer']
        client_loaders, test_loader = trainer.load_data()
        
        logger.info(f"Starting {training_state['total_rounds']} rounds of training")
        
        for round_num in range(training_state['total_rounds']):
            # Check if training was cancelled
            if training_state['status'] == 'idle':
                break
            
            # Run one round
            metrics = trainer.train_round(client_loaders, test_loader)
            
            with training_state['lock']:
                training_state['current_round'] = round_num + 1
                training_state['metrics_history'].append(metrics)
                training_state['current_metrics'] = metrics
            
            logger.info(f"Round {round_num + 1}/{training_state['total_rounds']}: "
                       f"Acc={metrics['accuracy']:.4f}, Loss={metrics['round_loss']:.4f}")
        
        with training_state['lock']:
            training_state['status'] = 'completed'
            logger.info("Training completed successfully")
            
    except Exception as e:
        with training_state['lock']:
            training_state['status'] = 'error'
            training_state['error'] = str(e)
            logger.error(f"Training error: {e}")

@app.route('/training/status', methods=['GET'])
def get_status():
    """Get current training status"""
    with training_state['lock']:
        config = training_state.get('config', TrainingConfig())
        return jsonify({
            'status': training_state['status'],
            'current_round': training_state['current_round'],
            'total_rounds': training_state['total_rounds'],
            'progress_percent': (training_state['current_round'] / max(training_state['total_rounds'], 1)) * 100,
            'current_metrics': training_state['current_metrics'],
            'error': training_state['error'],
            'dataset': config.dataset,
            'device': config.device,
            'multi_gpu': config.multi_gpu
        })

@app.route('/training/metrics', methods=['GET'])
def get_metrics():
    """Get all metrics collected so far"""
    with training_state['lock']:
        return jsonify({
            'metrics': training_state['metrics_history'],
            'total_collected': len(training_state['metrics_history']),
            'status': training_state['status']
        })

@app.route('/training/metrics_summary', methods=['GET'])
def metrics_summary():
    """Get summary of current training metrics (for BrowserFLDemo integration)"""
    with training_state['lock']:
        if not training_state['metrics_history']:
            # Return defaults if no training started
            return jsonify({
                'timestamp': datetime.utcnow().isoformat(),
                'status': training_state['status'],
                'round': training_state['current_round'],
                'accuracy': 0.1,
                'loss': 2.3,
                'compression_ratio': 1.0,
                'privacy_overhead': 0.0,
                'epsilon_used': 0.0
            })
        
        latest = training_state['metrics_history'][-1]
        return jsonify({
            'timestamp': datetime.utcnow().isoformat(),
            'status': training_state['status'],
            'round': training_state['current_round'],
            'accuracy': latest.get('accuracy', 0.1),
            'loss': latest.get('round_loss', 2.3),
            'compression_ratio': latest.get('compression_ratio', 1.0),
            'privacy_overhead': latest.get('privacy_overhead', 0.0),
            'epsilon_used': latest.get('epsilon', 0.0),
            'compression_bits': latest.get('compression_bits', 32)
        })

@app.route('/training/cancel', methods=['POST'])
def cancel_training():
    """Cancel ongoing training"""
    with training_state['lock']:
        if training_state['status'] != 'training':
            return jsonify({
                'status': 'error',
                'message': 'No training in progress'
            }), 400
        
        training_state['status'] = 'idle'
    
    return jsonify({
        'status': 'cancelled',
        'rounds_completed': training_state['current_round']
    })

@app.route('/training/history', methods=['GET'])
def get_history():
    """Get full training history as JSON"""
    with training_state['lock']:
        return jsonify({
            'history': training_state['metrics_history'],
            'config': {
                'num_rounds': training_state['total_rounds'],
                'current_round': training_state['current_round']
            }
        })

@app.route('/training/export', methods=['GET'])
def export_metrics():
    """Export training results as JSON file"""
    with training_state['lock']:
        data = {
            'timestamp': datetime.utcnow().isoformat(),
            'status': training_state['status'],
            'metrics': training_state['metrics_history']
        }
    
    # Save to file
    output_path = Path('./training_export.json')
    with open(output_path, 'w') as f:
        json.dump(data, f, indent=2)
    
    return jsonify({
        'status': 'success',
        'exported_file': str(output_path),
        'metrics_count': len(training_state['metrics_history'])
    })

@app.errorhandler(404)
def not_found(error):
    """Handle 404 errors"""
    return jsonify({
        'status': 'error',
        'message': 'Endpoint not found',
        'path': request.path
    }), 404

@app.errorhandler(500)
def server_error(error):
    """Handle 500 errors"""
    return jsonify({
        'status': 'error',
        'message': 'Internal server error',
        'error': str(error)
    }), 500

if __name__ == '__main__':
    logger.info("Starting Phase 3D training backend on http://localhost:5001")
    app.run(host='0.0.0.0', port=5001, debug=False)
