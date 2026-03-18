package com.sovereignmap.sdk;

import android.util.Log;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;

/**
 * Handles communication with the Sovereign Map aggregator with built-in Phase 3 Self-Healing.
 */
public class SelfHealingClient {
    private static final String TAG = "Sovereign-Network";
    private String aggregatorUrl;
    
    // Dynamic Phase 3 metrics initialized
    private int currentBatchSize = 64; 
    private long backoffIntervalMs = 1000L;

    public SelfHealingClient(String aggregatorUrl) {
        this.aggregatorUrl = aggregatorUrl;
    }

    private void handleSelfHealingBackoff(int responseCode) {
        if (responseCode == 429 || responseCode == 503) {
            // Aggregator metrics overloaded. Cut batch sizes down dynamically
            currentBatchSize = Math.max(16, currentBatchSize / 2);
            backoffIntervalMs *= 2;
            Log.w(TAG, String.format("Aggregator Overload (HTTP %d). Self-healing triggered: Reduced batch size to %d, backoff=%d ms", responseCode, currentBatchSize, backoffIntervalMs));
            
            try {
                Thread.sleep(backoffIntervalMs);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        } else if (responseCode == 200 && currentBatchSize < 64) {
            // Gradually restore throughput
            currentBatchSize = Math.min(64, currentBatchSize + 8);
            backoffIntervalMs = 1000L;
            Log.d(TAG, "Aggregator stabilized. Restoring throughput pool.");
        }
    }

    public boolean submitPayload(byte[] int8Gradients) {
        try {
            URL url = new URL(aggregatorUrl + "/api/v1/update");
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("POST");
            conn.setRequestProperty("Content-Type", "application/octet-stream");
            conn.setDoOutput(true);

            Log.d(TAG, "Transmitting " + int8Gradients.length + " bytes of Int8 compressed gradients.");

            try (OutputStream os = conn.getOutputStream()) {
                os.write(int8Gradients);
            }

            int responseCode = conn.getResponseCode();
            handleSelfHealingBackoff(responseCode);

            return responseCode == 200;
        } catch (Exception e) {
            Log.e(TAG, "Network connection broken. Engaging maximum failover.", e);
            handleSelfHealingBackoff(503);
            return false;
        }
    }
}