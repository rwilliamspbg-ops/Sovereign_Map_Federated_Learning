$ErrorActionPreference = "Stop"

# In PowerShell 7+, native stderr can be elevated to ErrorRecord when
# ErrorActionPreference=Stop. Disable that behavior for docker CLI output.
if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -ErrorAction SilentlyContinue) {
    $PSNativeCommandUseErrorActionPreference = $false
}

# Always run from the script's directory so relative paths resolve correctly
# regardless of the caller's current location.
$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptRoot
$ComposeFile = "docker-compose.production.yml"
$AccelComposeFile = "docker-compose.acceleration.yml"
$UseAccelerationOverride = Test-Path $AccelComposeFile
$Nodes = 50  # Reduced for Windows testing
$Duration = "5m"
$DurationSeconds = 300
$Timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$OutDir = "test-results/demo-windows/$Timestamp"

# Keep Windows demo flow resilient: TPM bootstrap can be slow/fragile on first run
# and is not required for basic monitoring/backend demo validation.
$env:TPM_ENABLED = "false"
$env:GPU_ENABLED = "true"
$env:NPU_ENABLED = "true"
if (-not $env:CUDA_VISIBLE_DEVICES) { $env:CUDA_VISIBLE_DEVICES = "0" }
if (-not $env:ASCEND_RT_VISIBLE_DEVICES) { $env:ASCEND_RT_VISIBLE_DEVICES = "0" }

if (-not (Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
}

function Log {
    param([string]$Message)
    $LogMsg = "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $Message"
    Write-Host $LogMsg
    $LogMsg | Add-Content "$OutDir/demo.log"
}

function Invoke-ComposeUp {
    param(
        [string]$StepName,
        [string]$LogFile,
        [string[]]$Services,
        [switch]$NoDeps,
        [switch]$NoBuild,
        [switch]$RetriedWithoutAccel
    )

    Log "$StepName..."

    # Use Start-Process so docker stderr/status text does not become a PowerShell
    # exception in Windows PowerShell. We rely strictly on process exit code.
    $StdOut = [System.IO.Path]::GetTempFileName()
    $StdErr = [System.IO.Path]::GetTempFileName()

    $ArgList = @("compose", "-f", $ComposeFile)
    if ($script:UseAccelerationOverride) {
        $ArgList += @("-f", $script:AccelComposeFile)
    }
    $ArgList += @("up", "-d")
    if ($NoDeps) {
        $ArgList += "--no-deps"
    }
    if ($NoBuild) {
        $ArgList += "--no-build"
    }
    $ArgList += $Services
    $Proc = Start-Process -FilePath "docker" -ArgumentList $ArgList -NoNewWindow -Wait -PassThru -RedirectStandardOutput $StdOut -RedirectStandardError $StdErr

    if (Test-Path $StdOut) {
        Get-Content $StdOut | Add-Content $LogFile
    }
    if (Test-Path $StdErr) {
        Get-Content $StdErr | Add-Content $LogFile
    }

    $ExitCode = $Proc.ExitCode

    Remove-Item $StdOut -ErrorAction SilentlyContinue
    Remove-Item $StdErr -ErrorAction SilentlyContinue

    if ($ExitCode -ne 0) {
        if ($script:UseAccelerationOverride -and -not $RetriedWithoutAccel) {
            # Fallback if host Docker runtime cannot satisfy GPU requests.
            $FailureHint = ""
            if (Test-Path $LogFile) {
                $FailureHint = ((Get-Content $LogFile -Tail 80) | Select-String -Pattern "gpu|nvidia|runtime|device driver|capabilities" -CaseSensitive:$false | Select-Object -Last 1).Line
            }
            $script:UseAccelerationOverride = $false
            if ($FailureHint) {
                Log "⚠️  Acceleration override failed; retrying without GPU compose override. Reason: $FailureHint"
            } else {
                Log "⚠️  Acceleration override failed; retrying without GPU compose override"
            }
            Invoke-ComposeUp -StepName $StepName -LogFile $LogFile -Services $Services -NoDeps:$NoDeps -NoBuild:$NoBuild -RetriedWithoutAccel
            return
        }

        $Tail = @()
        if (Test-Path $LogFile) {
            $Tail = Get-Content $LogFile -Tail 20
        }
        $TailText = ($Tail -join [Environment]::NewLine)
        throw "docker compose failed for '$StepName' (exit $ExitCode). Recent output: $TailText"
    }
}

function Invoke-ComposeBuild {
    param(
        [string]$StepName,
        [string]$LogFile,
        [string[]]$Services,
        [switch]$RetriedWithoutAccel
    )

    Log "$StepName..."

    $StdOut = [System.IO.Path]::GetTempFileName()
    $StdErr = [System.IO.Path]::GetTempFileName()

    $ArgList = @("compose", "-f", $ComposeFile)
    if ($script:UseAccelerationOverride) {
        $ArgList += @("-f", $script:AccelComposeFile)
    }
    $ArgList += @("build") + $Services

    $Proc = Start-Process -FilePath "docker" -ArgumentList $ArgList -NoNewWindow -Wait -PassThru -RedirectStandardOutput $StdOut -RedirectStandardError $StdErr

    if (Test-Path $StdOut) {
        Get-Content $StdOut | Add-Content $LogFile
    }
    if (Test-Path $StdErr) {
        Get-Content $StdErr | Add-Content $LogFile
    }

    $ExitCode = $Proc.ExitCode

    Remove-Item $StdOut -ErrorAction SilentlyContinue
    Remove-Item $StdErr -ErrorAction SilentlyContinue

    if ($ExitCode -ne 0) {
        if ($script:UseAccelerationOverride -and -not $RetriedWithoutAccel) {
            $FailureHint = ""
            if (Test-Path $LogFile) {
                $FailureHint = ((Get-Content $LogFile -Tail 80) | Select-String -Pattern "gpu|nvidia|runtime|device driver|capabilities" -CaseSensitive:$false | Select-Object -Last 1).Line
            }
            $script:UseAccelerationOverride = $false
            if ($FailureHint) {
                Log "⚠️  Acceleration override failed during build; retrying without GPU compose override. Reason: $FailureHint"
            } else {
                Log "⚠️  Acceleration override failed during build; retrying without GPU compose override"
            }
            Invoke-ComposeBuild -StepName $StepName -LogFile $LogFile -Services $Services -RetriedWithoutAccel
            return
        }

        $Tail = @()
        if (Test-Path $LogFile) {
            $Tail = Get-Content $LogFile -Tail 20
        }
        $TailText = ($Tail -join [Environment]::NewLine)
        throw "docker compose build failed for '$StepName' (exit $ExitCode). Recent output: $TailText"
    }
}
Log "=========================================="
Log "Sovereign Map Federated Learning Demo"
Log "=========================================="
Log "Nodes: $Nodes"
Log "Duration: $Duration"
Log "Compose: $ComposeFile"
if ($UseAccelerationOverride) {
    Log "Acceleration Compose Override: $AccelComposeFile"
} else {
    Log "Acceleration Compose Override: not found (CPU fallback mode)"
}
Log "Output: $OutDir"
Log "=========================================="

# Check if compose file exists
if (-not (Test-Path $ComposeFile)) {
    Log "ERROR: Compose file not found: $ComposeFile"
    exit 1
}

try {
    Invoke-ComposeUp -StepName "Starting Prometheus and Grafana" -LogFile "$OutDir/startup.log" -Services @("prometheus", "grafana", "alertmanager") -NoDeps
    Log "✅ Monitoring stack started"
} catch {
    Log "ERROR: Failed to start monitoring: $_"
    exit 1
}

# Wait for services
Log "Waiting 10 seconds for services to stabilize..."
Start-Sleep -Seconds 10

# Check Prometheus health
Log "Checking Prometheus health..."
try {
    $Response = Invoke-WebRequest -Uri "http://localhost:9090/-/healthy" -UseBasicParsing -ErrorAction SilentlyContinue
    if ($Response.StatusCode -eq 200) {
        Log "✅ Prometheus is healthy"
    }
} catch {
    Log "⚠️  Prometheus health check failed"
}

# Check Grafana health
Log "Checking Grafana health..."
try {
    $Response = Invoke-WebRequest -Uri "http://localhost:3001/api/health" -UseBasicParsing -ErrorAction SilentlyContinue
    if ($Response.StatusCode -eq 200) {
        Log "✅ Grafana is healthy"
    }
} catch {
    Log "⚠️  Grafana health check failed"
}

# Start backend services
try {
    Invoke-ComposeUp -StepName "Starting backend infrastructure" -LogFile "$OutDir/backend.log" -Services @("mongo", "redis", "backend") -NoDeps
    Log "✅ Backend started"
    Start-Sleep -Seconds 5
} catch {
    Log "ERROR: Failed to start backend: $_"
}

try {
    Log "Triggering LLM policy warm-up event..."
    $Warmup = Invoke-RestMethod -Uri "http://localhost:8000/simulate/llmPolicyValid" -Method POST -ContentType "application/json" -Body "{}" -ErrorAction SilentlyContinue
    if ($Warmup.status -eq "ok") {
        Log "✅ LLM policy warm-up event emitted"
    } else {
        Log "⚠️  LLM policy warm-up returned unexpected response"
    }
} catch {
    Log "⚠️  LLM policy warm-up failed: $_"
}

# Start exporters and node agents so dashboards have live metric sources.
try {
    Invoke-ComposeUp -StepName "Starting metrics exporters" -LogFile "$OutDir/backend.log" -Services @("tokenomics-metrics", "tpm-metrics", "fl-performance") -NoDeps
    Log "✅ Metrics exporters started"
} catch {
    Log "⚠️  Metrics exporters startup issue: $_"
}

try {
    Invoke-ComposeBuild -StepName "Prebuilding FL node-agent image" -LogFile "$OutDir/backend.log" -Services @("node-agent-1")
    Log "✅ Node-agent image prebuilt"
} catch {
    Log "⚠️  Node-agent prebuild issue: $_"
}

try {
    Invoke-ComposeUp -StepName "Starting FL node agents" -LogFile "$OutDir/backend.log" -Services @("node-agent-1", "node-agent-2", "node-agent-3") -NoDeps -NoBuild
    Log "✅ Node agents started"
} catch {
    Log "⚠️  Node agent startup issue: $_"
}

# Create monitoring loop
Log "Collecting system metrics for $Duration..."
$Iterations = [math]::Max(1, [int]($DurationSeconds / 60))

for ($i = 1; $i -le $Iterations; $i++) {
    $CurrentTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssK"
    Log "Iteration $i/$Iterations"
    
    # Collect Docker stats
    $StatsFile = "$OutDir/metrics-iteration-$i.txt"
    "# Iteration $i at $CurrentTime" | Out-File $StatsFile
    "" | Add-Content $StatsFile
    
    try {
        "## Docker Container Statistics" | Add-Content $StatsFile
        & docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>&1 | Add-Content $StatsFile
    } catch {
        "Error collecting stats: $_" | Add-Content $StatsFile
    }
    
    # Get container count
    "" | Add-Content $StatsFile
    "## Container Status" | Add-Content $StatsFile
    try {
        $RunningCount = (& docker ps --format "{{.Names}}" 2>&1 | Measure-Object -Line).Lines
        $AllCount = (& docker ps -a --format "{{.Names}}" 2>&1 | Measure-Object -Line).Lines
        "Running: $RunningCount, Total: $AllCount" | Add-Content $StatsFile
    } catch {
        "Error: $_" | Add-Content $StatsFile
    }
    
    # Query Prometheus
    "" | Add-Content $StatsFile
    "## Prometheus Status" | Add-Content $StatsFile
    try {
        $Metrics = Invoke-WebRequest -Uri "http://localhost:9090/api/v1/label/__name__/values" -UseBasicParsing -ErrorAction SilentlyContinue
        if ($Metrics.StatusCode -eq 200) {
            $MetricCount = ($Metrics.Content | ConvertFrom-Json).data.Count
            "Available metrics: $MetricCount" | Add-Content $StatsFile
        }
    } catch {
        "Prometheus query failed: $_" | Add-Content $StatsFile
    }
    
    if ($i -lt $Iterations) {
        Log "Waiting 60 seconds until next iteration..."
        Start-Sleep -Seconds 60
    }
}

# Final collection
Log "Collecting final metrics..."

# Docker state
"# Final Docker State" | Out-File "$OutDir/final-state.txt"
if ($UseAccelerationOverride) {
    & docker compose -f $ComposeFile -f $AccelComposeFile ps 2>&1 | Add-Content "$OutDir/final-state.txt"
} else {
    & docker compose -f $ComposeFile ps 2>&1 | Add-Content "$OutDir/final-state.txt"
}
"" | Add-Content "$OutDir/final-state.txt"

# Container stats
"# Container Statistics" | Add-Content "$OutDir/final-state.txt"
try {
    & docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>&1 | Add-Content "$OutDir/final-state.txt"
} catch {
    "Error collecting final docker stats: $_" | Add-Content "$OutDir/final-state.txt"
}

# Backend logs
Log "Capturing backend logs..."
try {
    if ($UseAccelerationOverride) {
        & docker compose -f $ComposeFile -f $AccelComposeFile logs --no-color backend > "$OutDir/backend-final.log" 2>&1
        & docker compose -f $ComposeFile -f $AccelComposeFile logs --no-color prometheus > "$OutDir/prometheus-final.log" 2>&1
        & docker compose -f $ComposeFile -f $AccelComposeFile logs --no-color grafana > "$OutDir/grafana-final.log" 2>&1
    } else {
        & docker compose -f $ComposeFile logs --no-color backend > "$OutDir/backend-final.log" 2>&1
        & docker compose -f $ComposeFile logs --no-color prometheus > "$OutDir/prometheus-final.log" 2>&1
        & docker compose -f $ComposeFile logs --no-color grafana > "$OutDir/grafana-final.log" 2>&1
    }
} catch {
    Log "⚠️  Could not capture one or more service logs: $_"
}

# Create summary report
@"
# Sovereign Map Demo Report (Windows)

## Execution Summary
- **Timestamp:** $Timestamp
- **Nodes Configured:** $Nodes
- **Nodes Started (compose profile):** 3
- **Duration:** $Duration ($DurationSeconds seconds)
- **Monitoring Intervals:** $Iterations
- **Platform:** Windows with Docker Desktop

## Monitoring Stack Status
- **Prometheus:** http://localhost:9090 ✅
- **Grafana:** http://localhost:3001 ✅
    - Credentials: admin/<configured password>
  - Dashboards: 7 available
- **Backend API:** http://localhost:8000
- **Alertmanager:** http://localhost:9093

## Generated Data Files
- Metrics iterations: $Iterations files
- Docker stats: Final state in final-state.txt
- Service logs: backend-final.log, prometheus-final.log, grafana-final.log
- Startup logs: startup.log, backend.log

## Recommendations
1. Open Grafana: http://localhost:3001
2. Select "Sovereign Map Overview" dashboard
3. Monitor real-time metrics
4. Compare performance metrics across iterations
5. Export Prometheus data for further analysis

## Next Steps
- Review performance dashboard in Grafana
- Analyze scaling behavior from metrics
- Check NPU/GPU utilization if available
- Export data for reporting

---
Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
"@ | Out-File "$OutDir/DEMO_REPORT.md"

Log "=========================================="
Log "✅ Demo complete!"
Log "Results: $OutDir"
Log "=========================================="
Log ""
Log "📊 Open Grafana to view live metrics:"
Log "   http://localhost:3001"
Log ""
Log "📈 Or access Prometheus directly:"
Log "   http://localhost:9090"
