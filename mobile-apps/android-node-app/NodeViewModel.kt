package io.sovereignmap.node

import android.os.Handler
import android.os.Looper
import android.util.Log
import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import java.nio.charset.StandardCharsets
import kotlin.math.min
import kotlin.random.Random

/**
 * ViewModel for Sovereign Node App
 * Manages node state, training loop, and metrics
 */
class NodeViewModel : ViewModel() {
    private val TAG = "SovereignNode"
    private val handler = Handler(Looper.getMainLooper())
    private val signer: MobileHardwareSigner = StrongBoxSigner()
    private val signerAlias = "mohawk.mobile.identity"
    private var updateRunnable: Runnable? = null
    
    // UI State
    private val _isConnected = MutableLiveData(false)
    val isConnected: LiveData<Boolean> = _isConnected
    
    private val _isTraining = MutableLiveData(false)
    val isTraining: LiveData<Boolean> = _isTraining
    
    private val _accuracy = MutableLiveData(0f)
    val accuracy: LiveData<Float> = _accuracy
    
    private val _loss = MutableLiveData(0f)
    val loss: LiveData<Float> = _loss
    
    private val _round = MutableLiveData(0)
    val round: LiveData<Int> = _round
    
    private val _statusMessage = MutableLiveData("Ready to join")
    val statusMessage: LiveData<String> = _statusMessage
    
    private val _nodeID = MutableLiveData(Random.nextInt(1000, 10000))
    val nodeID: LiveData<Int> = _nodeID
    
    fun joinNetwork() {
        _isConnected.value = true
        _isTraining.value = true
        _statusMessage.value = "Training in progress..."

        try {
            val warmup = "join-node-${_nodeID.value}".toByteArray(StandardCharsets.UTF_8)
            val signature = signer.sign(signerAlias, warmup)
            _statusMessage.value = "Hardware signer active (${signature.size}B signature)"
        } catch (err: Exception) {
            _statusMessage.value = "Signer unavailable: ${err.message ?: "unknown"}"
        }

        Log.i(TAG, "Node ${_nodeID.value} joined network")
        
        startTrainingLoop()
    }
    
    fun leaveNetwork() {
        _isConnected.value = false
        _isTraining.value = false
        _statusMessage.value = "Left network"
        _round.value = 0
        _accuracy.value = 0f
        _loss.value = 0f
        
        updateRunnable?.let { handler.removeCallbacks(it) }
        Log.i(TAG, "Node ${_nodeID.value} left network")
    }
    
    private fun startTrainingLoop() {
        updateRunnable = object : Runnable {
            override fun run() {
                val currentRound = (_round.value ?: 0) + 1
                _round.value = currentRound
                
                val newAccuracy = min(
                    0.99f,
                    0.65f + (currentRound * 0.02f) + Random.nextFloat(-0.01f, 0.01f)
                )
                val newLoss = (3.5f - (currentRound * 0.1f) + Random.nextFloat(-0.05f, 0.05f)).coerceAtLeast(0.1f)
                
                _accuracy.value = newAccuracy
                _loss.value = newLoss

                try {
                    val payload = "node=${_nodeID.value};round=$currentRound;ts=${System.currentTimeMillis()}"
                        .toByteArray(StandardCharsets.UTF_8)
                    val signature = signer.sign(signerAlias, payload)
                    _statusMessage.value = "Round $currentRound signed (${signature.size}B)"
                } catch (err: Exception) {
                    _statusMessage.value = "Round $currentRound unsigned: ${err.message ?: "unknown"}"
                }
                
                Log.i(TAG, "Round $currentRound: Accuracy=${String.format("%.2f", newAccuracy*100)}%, Loss=${String.format("%.4f", newLoss)}")
                
                handler.postDelayed(this, 5000) // Update every 5 seconds
            }
        }
        updateRunnable?.let { handler.post(it) }
    }
    
    override fun onCleared() {
        super.onCleared()
        updateRunnable?.let { handler.removeCallbacks(it) }
    }
}
