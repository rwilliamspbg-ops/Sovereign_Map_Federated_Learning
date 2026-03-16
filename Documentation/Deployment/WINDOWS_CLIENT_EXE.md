# Windows EXE Client Launcher

This guide packages a Windows executable that can register and connect your laptop as a federated learning client.

## What it does

- Performs join bootstrap against the backend Join API
- Saves certificate bundle and registration output locally
- Prints acceleration diagnostics (NPU, GPU, CPU)
- Starts FL client and connects to Flower aggregator

## Build on Windows

Open PowerShell in the repository root and run:

```powershell
powershell -ExecutionPolicy Bypass -File windows/build_windows_client_exe.ps1
```

Output executable:

- dist/SovereignMapClient.exe

## Run the EXE

With an invite code:

```powershell
.\dist\SovereignMapClient.exe `
  --backend-url http://YOUR_COORDINATOR_HOST:8000 `
  --participant-name laptop-client `
  --invite-code YOUR_INVITE_CODE
```

For local dev bootstrap (admin token path):

```powershell
.\dist\SovereignMapClient.exe `
  --backend-url http://localhost:8000 `
  --participant-name laptop-client `
  --admin-token local-dev-admin-token
```

Skip join bootstrap and connect directly:

```powershell
.\dist\SovereignMapClient.exe --skip-bootstrap --aggregator YOUR_HOST:8080 --node-id 77
```

## NPU and GPU diagnostics

At startup the EXE prints:

- torch version
- CUDA availability and device names
- NPU availability if torch.npu is present
- selected device and probe result

If acceleration is unavailable, it falls back to CPU and prints a warning.

## Output artifacts on laptop

Default output directory:

- %USERPROFILE%\SovereignMapClient\<participant-name>\

Generated files:

- join-registration.json
- certs/node-cert.pem
- certs/node-key.pem
- certs/ca-cert.pem
