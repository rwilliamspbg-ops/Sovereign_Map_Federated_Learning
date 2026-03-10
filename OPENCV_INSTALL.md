# OpenCV Installation Guide

## Overview

The camera capture and SLAM feature extraction modules require OpenCV 4.x installation. These components are fully implemented but require native OpenCV libraries to compile.

Since v1.1.0, all OpenCV-dependent packages use Go build constraints (`//go:build opencv`) to ensure clean compilation when OpenCV is not installed. Standard builds skip these packages automatically.

For compatibility with the Ubuntu/OpenCV toolchain currently used by this repository, `gocv.io/x/gocv` is pinned to `v0.32.0`. This pin is intentional: OpenCV `4.6.x` is compatible with GoCV `v0.32.0`, while newer GoCV releases expect newer ArUco APIs that are not available in OpenCV `4.6.x`.

## Packages Requiring OpenCV

- `sensors/camera/frame_capture.go` - Camera frame capture with GoCV
- `sensors/slam/` - SLAM feature extraction (ORB/SIFT/AKAZE)
- `storage/map_tiles/` - Tile encoding from Mat format

All files are marked with build constraint tags to prevent compilation failures in environments without OpenCV.

## Installation

### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install -y \
    libopencv-dev \
    libopencv-contrib-dev \
    pkg-config
```

### macOS (Homebrew)

```bash
brew install opencv pkg-config
```

### Verify Installation

```bash
pkg-config --modversion opencv4
```

Expected output: `4.x.x`

### Verify Repository Compatibility

This repository is validated against the following combination:

```bash
go list -m gocv.io/x/gocv
pkg-config --modversion opencv4
```

Expected output:
- GoCV: `v0.32.0`
- OpenCV: `4.6.x`

## Build Without OpenCV (Recommended for CI/Testing)

Since v1.1.0, standard builds automatically skip OpenCV-dependent packages:

```bash
# Build everything except OpenCV packages (SUCCEEDS without OpenCV)
go build ./...

# Build specific non-OpenCV packages
go build ./internal/consensus ./internal/batch ./internal/convergence
```

## Build With OpenCV (Optional for Full Features)

To include camera, SLAM, and map tile packages:

```bash
# Install OpenCV first
sudo apt-get install -y libopencv-dev libopencv-contrib-dev pkg-config

# Build all packages with opencv tag
go build -tags opencv ./...

# Test OpenCV packages
go test -tags opencv -v ./sensors/camera ./sensors/slam ./storage/map_tiles
```

If you upgrade OpenCV beyond `4.6.x`, revalidate the GoCV pin before changing it in `go.mod`.

## Alternative: Docker Build

For production deployments, use the Docker image which includes OpenCV:

```dockerfile
FROM opencv/opencv:4.8.0

RUN apt-get update && apt-get install -y golang-go

WORKDIR /app
COPY . .
RUN go build -o sovereign-node ./cmd/sovereign-node/
```

## Status

- ✅ Build constraints properly applied (v1.1.0+)
- ✅ Standard builds work without OpenCV installed
- ✅ All non-OpenCV packages test successfully
- ✅ OpenCV packages load with `-tags opencv` flag
- ✅ Main `sovereign-node` binary builds without OpenCV dependency
- ⚠️ OpenCV packages require 4.x native libraries to compile in tagged builds
- ⚠️ Repository is intentionally pinned to GoCV `v0.32.0` for OpenCV `4.6.x` compatibility

## Build Constraint Implementation

Files with `//go:build opencv` tag:
- `sensors/camera/frame_capture.go`
- `sensors/slam/feature_extraction.go`
- `sensors/slam/orbslam_bridge.go`
- `storage/map_tiles/tile_cache.go`
- `storage/map_tiles/tile_encoder.go`

When building without `-tags opencv`, these files are excluded automatically.

## CI/Testing Without OpenCV

Standard CI workflows build without OpenCV by default:

```bash
# CI build (automatic, no OpenCV required)
go build ./...
go test ./internal/... ./cmd/...
```

For full testing with OpenCV features, use GitHub Actions runners with pre-installed OpenCV or Docker:

```bash
# Docker build with OpenCV
docker build -f Dockerfile.backend -t sovereign-map:latest .
```
