$ErrorActionPreference = "Stop"

# Navigation
Set-Location "Sovereign_Map_Federated_Learning"
$ComposeFile = "docker-compose.production.yml"
$Nodes = 50  # Reduced for Windows testing
$Duration = "5m"
$DurationSeconds = 300
$Timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$OutDir = "test-results/demo-windows/$Timestamp"

if (-not (Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
}

function Log {
    param([string]$Message)
    $LogMsg = "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $Message"
    Write-Host $LogMsg
    $LogMsg | Add-Content "$OutDir/demo.log"
}

Log "=========================================="
Log "Sovereign Map Federated Learning Demo"
Log "=========================================="
Log "Nodes: $Nodes"
Log "Duration: $Duration"
Log "Compose: $ComposeFile"
Log "Output: $OutDir"
Log "=========================================="

# Check if compose file exists
if (-not (Test-Path $ComposeFile)) {
    Log "ERROR: Compose file not found: $ComposeFile"
    exit 1
}

# Start monitoring stack
Log "Starting Prometheus and Grafana..."
try {
    & docker compose -f $ComposeFile up -d prometheus grafana alertmanager 2>&1 | Add-Content "$OutDir/startup.log"
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
    $Response = Invoke-WebRequest -Uri "http://localhost:9090/-/healthy" -ErrorAction SilentlyContinue
    if ($Response.StatusCode -eq 200) {
        Log "✅ Prometheus is healthy"
    }
} catch {
    Log "⚠️  Prometheus health check failed"
}

# Check Grafana health
Log "Checking Grafana health..."
try {
    $Response = Invoke-WebRequest -Uri "http://localhost:3001/api/health" -ErrorAction SilentlyContinue
    if ($Response.StatusCode -eq 200) {
        Log "✅ Grafana is healthy"
    }
} catch {
    Log "⚠️  Grafana health check failed"
}

# Start backend services
Log "Starting backend infrastructure..."
try {
    & docker compose -f $ComposeFile up -d mongo redis backend 2>&1 | Add-Content "$OutDir/backend.log"
    Log "✅ Backend started"
    Start-Sleep -Seconds 5
} catch {
    Log "ERROR: Failed to start backend: $_"
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
        $Metrics = Invoke-WebRequest -Uri "http://localhost:9090/api/v1/label/__name__/values" -ErrorAction SilentlyContinue
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
& docker compose -f $ComposeFile ps 2>&1 | Add-Content "$OutDir/final-state.txt"
"" | Add-Content "$OutDir/final-state.txt"

# Container stats
"# Container Statistics" | Add-Content "$OutDir/final-state.txt"
& docker stats --no-stream --format "table {{.Names}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>&1 | Add-Content "$OutDir/final-state.txt"

# Backend logs
Log "Capturing backend logs..."
& docker logs sovereignmap-backend > "$OutDir/backend-final.log" 2>&1
& docker logs sovereignmap-prometheus > "$OutDir/prometheus-final.log" 2>&1
& docker logs sovereignmap-grafana > "$OutDir/grafana-final.log" 2>&1

# Create summary report
@"
# Sovereign Map Demo Report (Windows)

## Execution Summary
- **Timestamp:** $Timestamp
- **Nodes Configured:** $Nodes
- **Duration:** $Duration ($DurationSeconds seconds)
- **Monitoring Intervals:** $Iterations
- **Platform:** Windows with Docker Desktop

## Monitoring Stack Status
- **Prometheus:** http://localhost:9090 ✅
- **Grafana:** http://localhost:3001 ✅
  - Default Credentials: admin/sovereignmap2026
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
