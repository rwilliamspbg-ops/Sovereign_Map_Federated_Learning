//
//  ContentView.swift
//  SovereignNodeApp (iOS)
//
//  A simple 1-tap node app for joining Sovereign Map federated learning
//

import SwiftUI
import Network
import os.log

// MARK: - View Models

class NodeViewModel: ObservableObject {
    @Published var isConnected = false
    @Published var isTraining = false
    @Published var accuracy: Float = 0.0
    @Published var loss: Float = 0.0
    @Published var round: Int = 0
    @Published var statusMessage = "Ready to join"
    @Published var nodeID: Int = Int.random(in: 1000...9999)
    
    private let logger = Logger()
    private var updateTimer: Timer?
    
    func joinNetwork() {
        isConnected = true
        isTraining = true
        statusMessage = "Training in progress..."
        logger.info("Node \(self.nodeID) joined network")
        
        // Start training simulation
        startTrainingLoop()
    }
    
    func leaveNetwork() {
        isConnected = false
        isTraining = false
        statusMessage = "Left network"
        updateTimer?.invalidate()
        logger.info("Node \(self.nodeID) left network")
    }
    
    private func startTrainingLoop() {
        updateTimer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { [weak self] _ in
            guard let self = self else { return }
            
            self.round += 1
            self.accuracy = min(0.99, 0.65 + Float(self.round) * 0.02 + Float.random(in: -0.01...0.01))
            self.loss = max(0.1, 3.5 - Float(self.round) * 0.1 + Float.random(in: -0.05...0.05))
            
            self.logger.info("Round \(self.round): Accuracy=\(String(format: "%.2f", self.accuracy*100))%, Loss=\(String(format: "%.4f", self.loss))")
        }
    }
}

// MARK: - Main View

struct ContentView: View {
    @StateObject private var viewModel = NodeViewModel()
    
    var body: some View {
        ZStack {
            // Background gradient
            LinearGradient(
                gradient: Gradient(colors: [Color.blue.opacity(0.1), Color.purple.opacity(0.1)]),
                startPoint: .topLeadingPoint,
                endPoint: .bottomTrailingPoint
            )
            .ignoresSafeArea()
            
            VStack(spacing: 30) {
                // Header
                VStack(spacing: 8) {
                    Image(systemName: "network")
                        .font(.system(size: 40))
                        .foregroundColor(.blue)
                    
                    Text("Sovereign Node")
                        .font(.title)
                        .fontWeight(.bold)
                    
                    Text("Federated Learning on Mobile")
                        .font(.caption)
                        .foregroundColor(.gray)
                }
                .padding()
                
                // Status Card
                ZStack {
                    RoundedRectangle(cornerRadius: 12)
                        .fill(Color.white)
                        .shadow(radius: 4)
                    
                    VStack(spacing: 16) {
                        HStack {
                            Text("Node ID")
                                .foregroundColor(.gray)
                            Spacer()
                            Text("#\(viewModel.nodeID)")
                                .fontWeight(.bold)
                                .font(.system(.body, design: .monospaced))
                        }
                        
                        Divider()
                        
                        HStack {
                            Text("Status")
                                .foregroundColor(.gray)
                            Spacer()
                            HStack(spacing: 4) {
                                Circle()
                                    .fill(viewModel.isConnected ? Color.green : Color.gray)
                                    .frame(width: 8, height: 8)
                                Text(viewModel.isConnected ? "Connected" : "Offline")
                                    .fontWeight(.semibold)
                            }
                        }
                        
                        Divider()
                        
                        Text(viewModel.statusMessage)
                            .font(.caption)
                            .foregroundColor(.blue)
                            .lineLimit(2)
                    }
                    .padding()
                }
                .padding(.horizontal)
                
                // Metrics
                if viewModel.isConnected {
                    VStack(spacing: 12) {
                        MetricRow(label: "Round", value: "\(viewModel.round)")
                        MetricRow(label: "Accuracy", value: String(format: "%.2f%%", viewModel.accuracy * 100))
                        MetricRow(label: "Loss", value: String(format: "%.4f", viewModel.loss))
                    }
                    .padding()
                    .background(Color.gray.opacity(0.05))
                    .cornerRadius(12)
                    .padding(.horizontal)
                    .transition(.opacity)
                }
                
                Spacer()
                
                // Main Action Button
                VStack(spacing: 12) {
                    Button(action: {
                        if viewModel.isConnected {
                            viewModel.leaveNetwork()
                        } else {
                            viewModel.joinNetwork()
                        }
                    }) {
                        HStack {
                            Image(systemName: viewModel.isConnected ? "stop.circle.fill" : "play.circle.fill")
                                .font(.system(size: 20))
                            
                            Text(viewModel.isConnected ? "Leave Network" : "Join Network")
                                .fontWeight(.semibold)
                        }
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(viewModel.isConnected ? Color.red : Color.blue)
                        .foregroundColor(.white)
                        .cornerRadius(12)
                    }
                    .padding(.horizontal)
                    
                    // Settings Button
                    NavigationLink(destination: SettingsView(viewModel: viewModel)) {
                        HStack {
                            Image(systemName: "gear")
                            Text("Settings")
                        }
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.gray.opacity(0.2))
                        .foregroundColor(.blue)
                        .cornerRadius(12)
                    }
                    .padding(.horizontal)
                }
                .padding(.bottom, 30)
            }
            .padding()
        }
    }
}

// MARK: - Metric Row

struct MetricRow: View {
    let label: String
    let value: String
    
    var body: some View {
        HStack {
            Text(label)
                .foregroundColor(.gray)
                .font(.caption)
            Spacer()
            Text(value)
                .fontWeight(.semibold)
                .font(.system(.caption, design: .monospaced))
        }
    }
}

// MARK: - Settings View

struct SettingsView: View {
    @ObservedObject var viewModel: NodeViewModel
    @Environment(\.presentationMode) var presentationMode
    
    @State private var serverURL = "api.sovereignmap.io:8080"
    @State private var byzantineMode = false
    @State private var epochs = 3
    
    var body: some View {
        Form {
            Section(header: Text("Network")) {
                TextField("Server URL", text: $serverURL)
                    .disabled(viewModel.isConnected)
            }
            
            Section(header: Text("Training")) {
                Stepper("Epochs: \(epochs)", value: $epochs, in: 1...10)
                    .disabled(viewModel.isConnected)
                
                Toggle("Byzantine Mode", isOn: $byzantineMode)
                    .disabled(viewModel.isConnected)
            }
            
            Section(header: Text("Node")) {
                HStack {
                    Text("Node ID")
                    Spacer()
                    Text("#\(viewModel.nodeID)")
                        .fontWeight(.bold)
                }
            }
            
            Section {
                Button(action: { presentationMode.wrappedValue.dismiss() }) {
                    HStack {
                        Spacer()
                        Text("Close")
                        Spacer()
                    }
                }
                .foregroundColor(.blue)
            }
        }
        .navigationTitle("Settings")
    }
}

// MARK: - Preview

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        NavigationView {
            ContentView()
        }
    }
}
