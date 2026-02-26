# Sovereign Map Go Mobile Client

Complete Go implementation of Sovereign Map federated learning client for iOS, Android, and CLI tools.

## 📱 Mobile Support

### iOS

Build iOS framework for Swift integration:

```bash
# Build framework
gomobile bind -target=ios -o SovereignMap.xcframework ./pkg/client

# Use in Xcode
# 1. Drag SovereignMap.xcframework into project
# 2. Link in Build Phases
# 3. Import in Swift
```

**Swift Example:**

```swift
import SovereignMap

let client = SovereignMapClient("backend.example.com:8080")
try client.connect()

// Load training data
let features: [[Float]] = loadMNISTData()
let labels: [Int32] = loadLabels()
try client.loadData(features, labels: labels)

// Train locally
let (weights, metrics, err) = client.trainLocal()

// Check metrics
let status = client.status()
print("Accuracy: \(status["last_accuracy"] ?? 0.0)")
```

### Android

Build Android library for Kotlin integration:

```bash
# Build AAR
gomobile bind -target=android -androidapi 21 -o app/libs/sovereignmap.aar ./pkg/client

# Add to gradle
dependencies {
    implementation fileTree(dir: "libs", include: ["*.aar"])
}
```

**Kotlin Example:**

```kotlin
import sovereignmapclient.Client

val client = Client("backend.example.com:8080")
client.connect()

val metrics = client.trainLocal()
Log.d("FL", "Accuracy: ${metrics.lastAccuracy}")
```

## 🖥️ CLI Tool

Command-line tool for managing deployments and monitoring.

### Build

```bash
go build -o sovereignmap-cli ./cmd/cli
```

### Usage

```bash
# Get convergence data
./sovereignmap-cli convergence --server http://localhost:8000

# Watch real-time updates
./sovereignmap-cli convergence --server http://localhost:8000 --watch --interval 5

# Check health
./sovereignmap-cli health --server http://localhost:8000

# Export metrics
./sovereignmap-cli metrics export --format prometheus --output metrics.txt

# Scale deployment
./sovereignmap-cli scale --nodes 100 --server http://localhost:8000

# Add Byzantine node
./sovereignmap-cli node add --byzantine --server http://localhost:8000

# List connected nodes
./sovereignmap-cli node list --server http://localhost:8000

# Get node status
./sovereignmap-cli node status --id node-1 --server http://localhost:8000
```

## 📦 Package Structure

```
go-mobile/sovereignmapclient/
├── go.mod                          # Go module definition
├── pkg/
│   └── client/
│       └── client.go               # Core client implementation
├── cmd/
│   └── cli/
│       └── main.go                 # CLI tool
└── README.md                        # This file
```

## 🔧 Client API

### Initialization

```go
import "sovereignmap/client"

// Create client
c := client.NewSovereignMapClient("localhost:8080")

// Connect to aggregator
ctx := context.Background()
err := c.Connect(ctx)
```

### Data Loading

```go
// Load training data
features := [][]float32{...}  // MNIST features
labels := []int32{...}        // MNIST labels
err := c.LoadData(features, labels)
```

### Training

```go
// Train locally
weights, metrics, err := c.TrainLocal(ctx)

// Check metrics
metrics := c.GetMetrics()
fmt.Printf("Accuracy: %.2f%%\n", metrics.Accuracy*100)
```

### Configuration

```go
// Set training parameters
c.SetEpochs(5)
c.SetBatchSize(32)

// Enable Byzantine mode
c.SetByzantine(true)

// Check status
status := c.Status()
```

### Disconnection

```go
err := c.Disconnect()
```

## 🔒 Security

### TLS/mTLS Support

```go
// Load client certificate
cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
creds := credentials.NewTLS(&tls.Config{
    Certificates: []tls.Certificate{cert},
})

// Connect with mTLS
conn, err := grpc.Dial("backend:8080", grpc.WithTransportCredentials(creds))
```

### Privacy

```go
// Differential privacy configuration
c.privacyBudget = 1.0  // epsilon
// In production: use actual DP mechanism
```

## 📊 Metrics

Available metrics from training:

```go
type MetricsUpdate struct {
    Round      int32     // FL round number
    Accuracy   float32   // Model accuracy (0-1)
    Loss       float32   // Training loss
    Timestamp  int64     // Unix timestamp
    NodeID     string    // This node's ID
    Byzantine  bool      // Byzantine mode enabled
}
```

## 🚀 Deployment

### iOS Deployment

1. Build framework: `gomobile bind -target=ios -o SovereignMap.xcframework ./pkg/client`
2. Add to Xcode project
3. Link in Build Settings
4. Import in Swift code

### Android Deployment

1. Build AAR: `gomobile bind -target=android -androidapi 21 -o app/libs/sovereignmap.aar ./pkg/client`
2. Add to build.gradle
3. Import in Kotlin code

### CLI Deployment

```bash
# Build
go build -o sovereignmap-cli ./cmd/cli

# Run
./sovereignmap-cli convergence --server http://backend:8000
```

## 📚 Examples

### iOS Swift App

```swift
import SovereignMap
import SwiftUI

class FLModel: ObservableObject {
    @Published var accuracy: Float = 0.0
    let client: SovereignMapClient
    
    init() {
        self.client = SovereignMapClient("api.sovereignmap.io:443")
    }
    
    func startTraining() {
        DispatchQueue.global().async {
            do {
                try self.client.connect()
                try self.client.loadData(features, labels: labels)
                
                let (_, metrics, _) = self.client.trainLocal()
                
                DispatchQueue.main.async {
                    self.accuracy = metrics.accuracy
                }
                
                try self.client.disconnect()
            } catch {
                print("Training error: \(error)")
            }
        }
    }
}
```

### Android Kotlin App

```kotlin
import sovereignmapclient.Client
import kotlinx.coroutines.*

class FLViewModel : ViewModel() {
    private val client = Client("api.sovereignmap.io:443")
    
    fun trainModel() {
        viewModelScope.launch(Dispatchers.Default) {
            try {
                client.connect()
                client.loadData(features, labels)
                
                val metrics = client.trainLocal()
                
                withContext(Dispatchers.Main) {
                    updateUI(metrics)
                }
                
                client.disconnect()
            } catch (e: Exception) {
                Log.e("FL", "Training failed", e)
            }
        }
    }
}
```

### CLI Usage

```bash
#!/bin/bash

# Monitor convergence every 10 seconds
watch -n 10 'sovereignmap-cli convergence --server http://localhost:8000'

# Export daily metrics
sovereignmap-cli metrics export \
    --format prometheus \
    --output metrics-$(date +%Y%m%d).txt \
    --server http://localhost:8000

# Scale and monitor
sovereignmap-cli scale --nodes 100 --server http://localhost:8000
sleep 60
sovereignmap-cli convergence --server http://localhost:8000 --watch
```

## 🔄 WebSocket Real-Time Updates

Connect via WebSocket for real-time metric updates:

```go
import "github.com/gorilla/websocket"

var dialer = websocket.Dialer{}
ws, _, err := dialer.Dial("ws://localhost:8000/ws", nil)

var update struct {
    Round    int32
    Accuracy float32
    Loss     float32
}

err = ws.ReadJSON(&update)
fmt.Printf("Round %d: Accuracy %.2f%%\n", update.Round, update.Accuracy*100)
```

## 📋 Requirements

- Go 1.21+
- gomobile (for iOS/Android builds)
- gRPC and Protocol Buffers
- For mobile: Xcode (iOS) or Android Studio (Android)

## 🛠️ Development

### Build all targets

```bash
# Build CLI
go build -o sovereignmap-cli ./cmd/cli

# Build iOS framework
gomobile bind -target=ios -o SovereignMap.xcframework ./pkg/client

# Build Android AAR
gomobile bind -target=android -o sovereignmap.aar ./pkg/client

# Run tests
go test ./pkg/client/...
```

### Testing

```bash
go test -v ./pkg/client
go test -race ./...
go test -bench . ./...
```

## 📄 License

MIT License - Same as parent project

## 🤝 Contributing

Contributions welcome! Please submit PRs to main repository.

## 📞 Support

- Issues: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues
- Documentation: See parent repository README
- Examples: See examples/ directory

---

**Version**: 1.0.0  
**Status**: Production Ready  
**Last Updated**: February 2026
