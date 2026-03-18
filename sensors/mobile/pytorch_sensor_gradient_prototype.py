import torch
import torch.nn as nn
import torch.nn.functional as F


class SovereignMapFeatureExtractor(nn.Module):
    def __init__(self):
        super().__init__()
        # Camera pipeline (RGB image to feature map)
        self.conv1 = nn.Conv2d(3, 16, kernel_size=3, padding=1)
        # Lidar pipeline (Depth map to spatial map)
        self.conv_depth = nn.Conv2d(1, 16, kernel_size=3, padding=1)

        # Fusion layer (Combines Camera + Lidar + GPS Embeddings)
        self.fusion_fc = nn.Linear(16 * 2 + 2, 128)
        self.output_layer = nn.Linear(128, 64)  # 64-dim Map Feature Vector

    def forward(self, rgb, depth, gps):
        # Process Camera
        c = F.relu(self.conv1(rgb))
        c = F.adaptive_avg_pool2d(c, (1, 1)).view(c.size(0), -1)

        # Process Lidar
        d = F.relu(self.conv_depth(depth))
        d = F.adaptive_avg_pool2d(d, (1, 1)).view(d.size(0), -1)

        # Concat Spatial Data
        combined = torch.cat((c, d, gps), dim=1)

        fused = F.relu(self.fusion_fc(combined))
        return self.output_layer(fused)


def simulate_edge_training():
    print("Initializing Sovereign Map Local PyTorch Prototype...")
    model = SovereignMapFeatureExtractor()
    optimizer = torch.optim.SGD(model.parameters(), lr=0.01)

    # 1. Capture Raw Sensor Data (Mock Edge Inputs)
    batch_size = 1
    camera_rgb = torch.randn(batch_size, 3, 224, 224)  # [1, 3, 224, 224] image
    lidar_depth = torch.randn(batch_size, 1, 224, 224)  # [1, 1, 224, 224] depth
    gps_lat_lon = torch.tensor([[37.7749, -122.4194]])  # SF GPS coordinates

    # 2. Simulate the Global Sovereign Map Tile (Target state we want to compare against)
    # In reality, this is the signed Tile State downloaded over gossip protocol
    global_map_features = torch.randn(batch_size, 64)

    optimizer.zero_grad()

    # 3. Local Model Forward Pass: Sensor -> Vector
    local_features = model(camera_rgb, lidar_depth, gps_lat_lon)

    # 4. Calculate Difference (Semantic Loss against Global Map)
    loss = F.mse_loss(local_features, global_map_features)
    print(f"Computed Sensor-to-Map Loss: {loss.item():.4f}")

    # 5. Backward Pass: Create the Gradients
    loss.backward()

    # 6. Extract Gradients to Array (This feeds into DP Quantization in Java SDK!)
    flattened_grads = []
    for param in model.parameters():
        if param.grad is not None:
            flattened_grads.extend(param.grad.view(-1).tolist())

    print(f"Extracted {len(flattened_grads)} raw gradient floats from sensor inputs.")
    print(f"First 5 gradients: {flattened_grads[:5]}")
    print(
        "\nNext step -> These float arrays are passed to TEEWrapper.executeDPQuantization()"
    )


if __name__ == "__main__":
    simulate_edge_training()
