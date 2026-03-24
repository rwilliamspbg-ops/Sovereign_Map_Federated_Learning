#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Sovereign Map Testing Cleanup Script
    Cleans up Docker environment after testing phase completion

.DESCRIPTION
    Removes test containers, volumes, and temporary artifacts.
    Keeps production code and configuration intact.

.PARAMETER CleanupLevel
    full     - Remove everything (default)
    selective - Keep monitoring stack
    archive  - Archive first, then clean

.EXAMPLE
    .\tests/scripts/powershell/cleanup-testing.ps1 -CleanupLevel full
    .\tests/scripts/powershell/cleanup-testing.ps1 -CleanupLevel selective
    .\tests/scripts/powershell/cleanup-testing.ps1 -CleanupLevel archive
#>

param(
    [ValidateSet("full", "selective", "archive")]
    [string]$CleanupLevel = "full",
    
    [switch]$Force = $false,
    [switch]$DryRun = $false
)

$ErrorActionPreference = "Stop"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"

function Write-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "=" * 80
    Write-Host $Message
    Write-Host "=" * 80
}

function Write-Section {
    param([string]$Message)
    Write-Host ""
    Write-Host ">>> $Message"
    Write-Host ""
}

function Write-Success {
    param([string]$Message)
    Write-Host "    [OK] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "    [!] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "    [ERROR] $Message" -ForegroundColor Red
}

function Execute-Command {
    param(
        [string]$Command,
        [string]$Description
    )
    
    Write-Host "    $Description..." -NoNewline
    
    if ($DryRun) {
        Write-Host " [DRY RUN]" -ForegroundColor Cyan
        Write-Host "    Would execute: $Command" -ForegroundColor Gray
        return $true
    }
    
    try {
        Invoke-Expression $Command | Out-Null
        Write-Success "Done"
        return $true
    }
    catch {
        Write-Error-Custom "Failed: $_"
        if ($Force) {
            return $true
        }
        return $false
    }
}

# Main cleanup logic
Write-Header "SOVEREIGN MAP - TESTING PHASE CLEANUP"
Write-Host "Cleanup Level: $CleanupLevel"
Write-Host "Dry Run: $($DryRun ? 'YES' : 'NO')"
Write-Host "Timestamp: $timestamp"

# Confirmation
if (-not $Force -and -not $DryRun) {
    Write-Host ""
    Write-Warning "This will remove Docker containers and volumes"
    $confirmation = Read-Host "Continue? (yes/no)"
    if ($confirmation -ne "yes") {
        Write-Host "Cleanup cancelled."
        exit 0
    }
}

Write-Section "Phase 1: Pre-Cleanup Verification"

# Check Docker
Write-Host "    Checking Docker availability..." -NoNewline
try {
    $dockerVersion = docker version --format "{{.Server.Version}}"
    Write-Success "Docker $dockerVersion found"
} catch {
    Write-Error-Custom "Docker not available"
    exit 1
}

# Count current resources
$containerCount = (docker ps -a --format "{{.Names}}" | Measure-Object -Line).Lines
$volumeCount = (docker volume ls -q | Measure-Object -Line).Lines
$imageCount = (docker images -q | Measure-Object -Line).Lines

Write-Host "    Current resources:"
Write-Host "        - Containers: $containerCount"
Write-Host "        - Volumes: $volumeCount"
Write-Host "        - Images: $imageCount"

Write-Section "Phase 2: Docker Compose Down"

$composeFile = "docker-compose.full.yml"
if (Test-Path $composeFile) {
    if ($CleanupLevel -eq "selective") {
        Write-Host "    Selective cleanup: Stopping specific services"
        Execute-Command "docker compose -f $composeFile down node-agent backend mongo redis --remove-orphans" `
            "Stopping node agents and backend services"
    } else {
        Execute-Command "docker compose -f $composeFile down --remove-orphans" `
            "Stopping all Docker Compose services"
    }
} else {
    Write-Warning "Compose file not found: $composeFile"
}

Write-Section "Phase 3: Volume Cleanup"

if ($CleanupLevel -ne "selective") {
    $orphanedVolumes = docker volume ls -f dangling=true -q
    if ($orphanedVolumes) {
        Write-Host "    Found $($orphanedVolumes.Count) orphaned volumes"
        Execute-Command "docker volume prune -f" "Removing orphaned volumes"
    } else {
        Write-Success "No orphaned volumes found"
    }
}

Write-Section "Phase 4: System Cleanup"

if ($CleanupLevel -eq "full" -or $CleanupLevel -eq "archive") {
    Write-Host "    Removing unused images and containers"
    Execute-Command "docker system prune -f" "Pruning Docker system"
}

Write-Section "Phase 5: Archive (if selected)"

if ($CleanupLevel -eq "archive") {
    $archiveFile = "sovereign-map-test-archive-$timestamp.tar.gz"
    Write-Host "    Archiving test results to: $archiveFile"
    
    if (Test-Path "test-results") {
        Execute-Command "tar -czf '$archiveFile' test-results/" `
            "Creating test results archive"
        Write-Success "Archive created: $archiveFile"
    } else {
        Write-Warning "test-results directory not found, skipping archive"
    }
}

Write-Section "Phase 6: Post-Cleanup Verification"

$newContainerCount = (docker ps -a --format "{{.Names}}" 2>/dev/null | Measure-Object -Line).Lines
$newVolumeCount = (docker volume ls -q 2>/dev/null | Measure-Object -Line).Lines

Write-Host "    Resource counts after cleanup:"
Write-Host "        - Containers: $containerCount -> $newContainerCount"
Write-Host "        - Volumes: $volumeCount -> $newVolumeCount"

$freedContainers = $containerCount - $newContainerCount
$freedVolumes = $volumeCount - $newVolumeCount

if ($freedContainers -gt 0) {
    Write-Success "Freed $freedContainers containers"
}
if ($freedVolumes -gt 0) {
    Write-Success "Freed $freedVolumes volumes"
}

# Disk space estimation
Write-Section "Phase 7: Final Summary"

Write-Host "    Cleanup Level: $CleanupLevel"
if ($DryRun) {
    Write-Host "    Mode: DRY RUN (no changes made)" -ForegroundColor Yellow
} else {
    Write-Host "    Mode: LIVE (changes applied)" -ForegroundColor Green
}

Write-Host ""
Write-Host "    Summary:"
Write-Host "        Containers removed: $freedContainers"
Write-Host "        Volumes removed: $freedVolumes"
Write-Host "        Timestamp: $timestamp"

# Create cleanup report
$reportFile = "cleanup-report-$timestamp.txt"
@"
SOVEREIGN MAP TESTING - CLEANUP REPORT
======================================

Timestamp: $timestamp
Cleanup Level: $CleanupLevel
Dry Run: $($DryRun ? 'YES' : 'NO')

PRE-CLEANUP STATE
-----------------
Containers: $containerCount
Volumes: $volumeCount
Images: $imageCount

POST-CLEANUP STATE
------------------
Containers: $newContainerCount (removed: $freedContainers)
Volumes: $newVolumeCount (removed: $freedVolumes)

ACTIONS TAKEN
-------------
- Docker Compose services stopped: $(if($CleanupLevel -eq 'selective') { 'Selective' } else { 'All' })
- Orphaned volumes pruned: $(if($CleanupLevel -ne 'selective') { 'Yes' } else { 'No' })
- Docker system pruned: $(if($CleanupLevel -eq 'full' -or $CleanupLevel -eq 'archive') { 'Yes' } else { 'No' })
- Archive created: $(if($CleanupLevel -eq 'archive') { 'Yes' } else { 'No' })

RECOMMENDATIONS
---------------
1. Keep Docker images for quick restart: yes
2. Retain configuration files: yes
3. Archive test data: recommended
4. Review final disk usage: recommended
5. Update deployment documentation: recommended

STATUS: CLEANUP COMPLETE
========================
"@ | Out-File $reportFile -Encoding UTF8

Write-Success "Cleanup report saved to: $reportFile"

Write-Header "CLEANUP COMPLETE"
Write-Host "Cleanup Level: $CleanupLevel"
Write-Host "Timestamp: $timestamp"
Write-Host "Report: $reportFile"
Write-Host ""

if ($DryRun) {
    Write-Warning "DRY RUN - No changes were actually made"
    Write-Host "Re-run without -DryRun to apply changes"
}

Write-Host "Next Steps:"
Write-Host "  1. Review cleanup report: $reportFile"
Write-Host "  2. Verify Docker state: docker system df"
Write-Host "  3. Check disk usage: df -h"
Write-Host "  4. Archive test results if not done: tar -czf archive.tar.gz test-results/"
Write-Host ""
