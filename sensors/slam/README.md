# SLAM Module Notes

This directory contains OpenCV-backed SLAM feature extraction and ORB-SLAM bridge integration paths used by Sovereign Map runtime and tests.

All files in this module use `//go:build opencv` build constraints. They are automatically excluded from standard builds and only compiled when the `-tags opencv` flag is used.

This module follows the repository-wide OpenCV compatibility pin: GoCV is pinned to `v0.32.0` because the current OpenCV `4.6.x` environment does not support the newer ArUco API expected by more recent GoCV releases.

## Build Instructions

### Standard Build (Default)
SLAM module is automatically excluded:
```bash
go build ./...
```

### Build with OpenCV
Include SLAM features:
```bash
# Install OpenCV first
sudo apt-get install -y libopencv-dev

# Build with opencv tag
go build -tags opencv ./sensors/slam/
```

Compatibility baseline for this module:
- GoCV: `v0.32.0`
- OpenCV: `4.6.x`

## Release Note

- v1.1.0: ORB-SLAM map SaveMap/LoadMap support, OpenCV build constraint system for CI compatibility.
