param(
  [string]$VenvDir = ".venv-windows-client",
  [string]$ExeName = "SovereignMapClient",
  [switch]$OneDir
)

$ErrorActionPreference = "Stop"

Write-Host "Building Windows client EXE..."

if (-not (Test-Path $VenvDir)) {
  py -3.11 -m venv $VenvDir
}

$PythonExe = Join-Path $VenvDir "Scripts/python.exe"
$PipExe = Join-Path $VenvDir "Scripts/pip.exe"

& $PipExe install --upgrade pip setuptools wheel
& $PipExe install pyinstaller
& $PipExe install flwr==1.7.0 torch==2.1.0 torchvision==0.16.0 opacus==1.4.0 numpy==1.24.3

$mode = "--onefile"
if ($OneDir) {
  $mode = "--onedir"
}

& $PythonExe -m PyInstaller $mode --clean --noconfirm `
  --name $ExeName `
  --hidden-import src.client `
  --collect-submodules flwr `
  --collect-submodules opacus `
  windows/client_launcher.py

Write-Host "Build complete. Output in dist/$ExeName"
Write-Host "Example run:"
Write-Host "  .\\dist\\$ExeName.exe --backend-url http://<host>:8000 --participant-name laptop --invite-code <code>"
