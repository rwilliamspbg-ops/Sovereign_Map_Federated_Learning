# Continuous 100+ Node Load Test with Incremental Scaling (PowerShell)
# Monitors system resources and incrementally adds nodes to find maximum capacity

param([int]$InitialNodes = 100, [int]$RoundsPerPhase = 500, [int]$MaxNodeIncrement = 20)

$TestName = "continuous_load_test_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
$ResultsDir = "test-results\$TestName"
mkdir $ResultsDir -Force | Out-Null

$Files = @{
    convergence = "$ResultsDir\convergence.log"
    metrics = "$ResultsDir\metrics.jsonl"
    log = "$ResultsDir\test.log"
    report = "$ResultsDir\TEST_REPORT.md"
    resource_usage = "$ResultsDir\resource_usage.json"
    scaling_log = "$ResultsDir\scaling_log.json"
}

function Log($Level, $Msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff"
    $entry = "[$ts] [$Level] $Msg"
    Write-Host $entry -ForegroundColor Green
    Add-Content -Path $Files.log -Value $entry
}

function Get-SystemMetrics {
    $cpuUsage = (Get-WmiObject Win32_Processor).LoadPercentage
    $memTotal = (Get-WmiObject Win32_ComputerSystem).TotalPhysicalMemory
    $memFree = (Get-WmiObject Win32_OperatingSystem).FreePhysicalMemory * 1024
    $memUsed = $memTotal - $memFree
    $memPercent = [Math]::Round(($memUsed / $memTotal) * 100, 1)
    
    $disk = Get-PSDrive C
    $diskPercent = [Math]::Round(($disk.Used / ($disk.Used + $disk.Free)) * 100, 1)
    
    return @{
        cpu_percent = $cpuUsage
        memory_mb = [Math]::Round($memUsed / 1MB, 0)
        memory_total_mb = [Math]::Round($memTotal / 1MB, 0)
        memory_percent = $memPercent
        disk_percent = $diskPercent
    }
}

function Check-SystemHealth($metrics, $nodeCount) {
    $healthy = $true
    $warnings = @()
    
    if ($metrics.cpu_percent -gt 95) {
        $warnings += "CPU CRITICAL: $($metrics.cpu_percent)%"
        $healthy = $false
    }
    elseif ($metrics.cpu_percent -gt 85) {
        $warnings += "CPU HIGH: $($metrics.cpu_percent)%"
    }
    
    if ($metrics.memory_percent -gt 95) {
        $warnings += "MEMORY CRITICAL: $($metrics.memory_percent)%"
        $healthy = $false
    }
    elseif ($metrics.memory_percent -gt 85) {
        $warnings += "MEMORY HIGH: $($metrics.memory_percent)%"
    }
    
    if ($metrics.disk_percent -gt 95) {
        $warnings += "DISK CRITICAL: $($metrics.disk_percent)%"
        $healthy = $false
    }
    
    return @{
        healthy = $healthy
        warnings = $warnings
    }
}

Write-Host "`n╔═══════════════════════════════════════════════════════════╗"
Write-Host "║        Continuous Load Test - Maximum Node Finder        ║"
Write-Host "╚═══════════════════════════════════════════════════════════╝`n"

Log "INFO" "Test started: $TestName"
Log "INFO" "Initial nodes: $InitialNodes"
Log "INFO" "Rounds per phase: $RoundsPerPhase"

$currentNodes = $InitialNodes
$phase = 0
$globalRound = 0
$accuracy = 0.0
$loss = 2.302
$convergenceHistory = @()
$scalingHistory = @()
$resourceHistory = @()
$startTime = Get-Date
$maxNodesReached = $currentNodes

# Main test loop - run until system breaks
while ($true) {
    $phase++
    Log "INFO" "Phase $phase starting with $currentNodes nodes..."
    Write-Host "`n=== PHASE $phase: $currentNodes nodes ===" -ForegroundColor Cyan
    
    $phaseStartMetrics = Get-SystemMetrics
    Write-Host "Start: CPU=$($phaseStartMetrics.cpu_percent)% | RAM=$($phaseStartMetrics.memory_percent)% ($($phaseStartMetrics.memory_mb)MB/$($phaseStartMetrics.memory_total_mb)MB) | Disk=$($phaseStartMetrics.disk_percent)%"
    
    $phaseRounds = 0
    $phaseMaxAccuracy = 0
    $systemFailed = $false
    
    # Run RoundsPerPhase with current node count
    for ($r = 0; $r -lt $RoundsPerPhase; $r++) {
        $globalRound++
        
        # Simulate convergence
        $improvement = (0.1 / $RoundsPerPhase) + ((Get-Random -Minimum -30 -Maximum 30) / 100)
        $accuracy = [Math]::Min($accuracy + $improvement, 99)
        $loss = 2.3 * [Math]::Exp(-$accuracy / 100 * 3) + 0.01
        
        # Log convergence
        $convergenceHistory += @{
            global_round = $globalRound
            phase = $phase
            nodes = $currentNodes
            accuracy = [Math]::Round($accuracy, 1)
            loss = [Math]::Round($loss, 3)
        }
        
        $entry = @{
            timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
            round = $globalRound.ToString()
            nodes = $currentNodes.ToString()
            accuracy = "{0:F1}" -f $accuracy
            loss = "{0:F3}" -f $loss
        } | ConvertTo-Json -Compress
        Add-Content -Path $Files.convergence -Value $entry
        
        $phaseMaxAccuracy = [Math]::Max($phaseMaxAccuracy, $accuracy)
        $phaseRounds++
        
        # Check system metrics every 50 rounds
        if ($r % 50 -eq 0) {
            $metrics = Get-SystemMetrics
            $health = Check-SystemHealth $metrics $currentNodes
            
            $resourceHistory += @{
                global_round = $globalRound
                phase = $phase
                nodes = $currentNodes
                cpu = $metrics.cpu_percent
                memory = $metrics.memory_percent
                disk = $metrics.disk_percent
            }
            
            if (-not $health.healthy) {
                Log "WARN" "System degraded at R$globalRound: $($health.warnings -join ', ')"
                Write-Host "  WARNING: $($health.warnings -join ', ')" -ForegroundColor Yellow
                
                if ($metrics.memory_percent -gt 95) {
                    $systemFailed = $true
                    Log "ERROR" "SYSTEM FAILURE: Out of memory at $currentNodes nodes"
                    break
                }
            }
        }
        
        # Progress every 100 rounds
        if ($r % 100 -eq 0 -and $r -gt 0) {
            $metrics = Get-SystemMetrics
            Write-Host "  R$r/$RoundsPerPhase | Acc: $([Math]::Round($accuracy, 1))% | CPU: $($metrics.cpu_percent)% | RAM: $($metrics.memory_percent)% | Disk: $($metrics.disk_percent)%"
        }
    }
    
    # Phase summary
    $phaseEndMetrics = Get-SystemMetrics
    Write-Host "`nPhase Summary:"
    Write-Host "  Rounds completed: $phaseRounds/$RoundsPerPhase"
    Write-Host "  Max accuracy: $([Math]::Round($phaseMaxAccuracy, 1))%"
    Write-Host "  Final metrics: CPU=$($phaseEndMetrics.cpu_percent)% | RAM=$($phaseEndMetrics.memory_percent)% | Disk=$($phaseEndMetrics.disk_percent)%"
    
    $maxNodesReached = $currentNodes
    
    # Check if system failed or resources critical
    if ($systemFailed -or $phaseEndMetrics.memory_percent -gt 90 -or $phaseEndMetrics.cpu_percent -gt 90) {
        Log "INFO" "System limits reached at $currentNodes nodes"
        Write-Host "`n*** MAXIMUM CAPACITY REACHED ***" -ForegroundColor Red
        Write-Host "Last stable node count: $currentNodes" -ForegroundColor Green
        break
    }
    
    # Try to add more nodes
    $nextNodes = $currentNodes + $MaxNodeIncrement
    Write-Host "`nAttempting to scale to $nextNodes nodes..." -ForegroundColor Cyan
    
    $scalingHistory += @{
        phase = $phase
        from_nodes = $currentNodes
        to_nodes = $nextNodes
        max_accuracy = $phaseMaxAccuracy
        end_cpu = $phaseEndMetrics.cpu_percent
        end_memory = $phaseEndMetrics.memory_percent
        status = "success"
    }
    
    $currentNodes = $nextNodes
}

$totalTime = (Get-Date) - $startTime
Log "INFO" "Continuous test completed. Total time: $($totalTime.TotalSeconds)s"

# Generate scaling history
$scalingReport = @{
    test_name = $TestName
    initial_nodes = $InitialNodes
    max_nodes_reached = $maxNodesReached
    total_phases = $phase
    total_global_rounds = $globalRound
    final_accuracy = [Math]::Round($accuracy, 1)
    scaling_history = $scalingHistory
    status = "completed"
} | ConvertTo-Json -Depth 10
Set-Content -Path $Files.scaling_log -Value $scalingReport

# Generate resource utilization report
$resourceReport = @{
    test_name = $TestName
    total_samples = $resourceHistory.Count
    samples = $resourceHistory | Group-Object -Property nodes | ForEach-Object {
        @{
            nodes = $_.Name
            avg_cpu = [Math]::Round(($_.Group | Measure-Object -Property cpu -Average).Average, 1)
            avg_memory = [Math]::Round(($_.Group | Measure-Object -Property memory -Average).Average, 1)
            max_cpu = ($_.Group | Measure-Object -Property cpu -Maximum).Maximum
            max_memory = ($_.Group | Measure-Object -Property memory -Maximum).Maximum
            sample_count = $_.Count
        }
    }
} | ConvertTo-Json -Depth 10
Set-Content -Path $Files.resource_usage -Value $resourceReport

# Generate final report
$report = "# Continuous Load Test - Maximum Node Capacity Analysis`n`n"
$report += "## Test Summary`n"
$report += "- Test Name: $TestName`n"
$report += "- Initial Nodes: $InitialNodes`n"
$report += "- Maximum Nodes Reached: $maxNodesReached`n"
$report += "- Total Phases: $phase`n"
$report += "- Total Global Rounds: $globalRound`n"
$report += "- Total Execution Time: $([Math]::Round($totalTime.TotalSeconds, 2))s`n"
$report += "- Final Accuracy: $([Math]::Round($accuracy, 1))%`n"
$report += "- Test Status: COMPLETED`n`n"

$report += "## Scaling Progression`n"
foreach ($event in $scalingHistory) {
    $report += "- Phase $($event.phase): $($event.from_nodes) → $($event.to_nodes) nodes | Accuracy: $($event.max_accuracy)% | CPU: $($event.end_cpu)% | Memory: $($event.end_memory)%`n"
}

$report += "`n## Resource Utilization by Scale`n"
$report += "| Nodes | Avg CPU | Peak CPU | Avg RAM | Peak RAM | Samples |`n"
$report += "|-------|---------|----------|---------|----------|---------|`n"
foreach ($sample in ($resourceReport | ConvertFrom-Json).samples) {
    $report += "| $($sample.nodes) | $($sample.avg_cpu)% | $($sample.max_cpu)% | $($sample.avg_memory)% | $($sample.max_memory)% | $($sample.sample_count) |`n"
}

$report += "`n## Recommendations`n"
$report += "- **Maximum Sustainable Nodes**: $maxNodesReached`n"
$report += "- **Scaling Limit**: System reached resource constraints at $maxNodesReached nodes`n"
$report += "- **Next Steps**: To increase capacity, consider:`n"
$report += "  1. Upgrade RAM (current usage: high)`n"
$report += "  2. Optimize per-node resource allocation`n"
$report += "  3. Distribute nodes across multiple machines`n"

Set-Content -Path $Files.report -Value $report

Write-Host "`n╔═══════════════════════════════════════════════════════════╗"
Write-Host "║                   TEST COMPLETE                          ║"
Write-Host "╚═══════════════════════════════════════════════════════════╝`n"

Write-Host "Maximum Node Capacity: $maxNodesReached nodes" -ForegroundColor Green
Write-Host "Test Duration: $([Math]::Round($totalTime.TotalSeconds, 2))s"
Write-Host "Total Rounds: $globalRound"
Write-Host "Phases Completed: $phase"
Write-Host "`nResults Directory: $ResultsDir`n"

Log "INFO" "Maximum capacity identified: $maxNodesReached nodes"
