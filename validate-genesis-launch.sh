#!/bin/bash
################################################################################
# Genesis Launch Validation Script
# 
# This script performs comprehensive validation of all Genesis launch
# components before production deployment.
#
# Usage: ./validate-genesis-launch.sh
################################################################################

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

################################################################################
# Helper Functions
################################################################################

print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED+=1))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED+=1))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS+=1))
}

################################################################################
# Validation Checks
################################################################################

validate_system_requirements() {
    print_header "System Requirements"
    
    # Check Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | grep -oP '\d+\.\d+' | head -1)
        DOCKER_MAJOR=$(echo "$DOCKER_VERSION" | cut -d. -f1)
        if [[ $DOCKER_MAJOR -ge 24 ]]; then
            check_pass "Docker version $DOCKER_VERSION >= 24.0"
        else
            check_fail "Docker version $DOCKER_VERSION < 24.0 (upgrade required)"
        fi
    else
        check_fail "Docker not installed"
    fi
    
    # Check Docker Compose
    if docker compose version &> /dev/null; then
        COMPOSE_VERSION=$(docker compose version | grep -oP '\d+\.\d+' | head -1)
        COMPOSE_MAJOR=$(echo "$COMPOSE_VERSION" | cut -d. -f1)
        if [[ $COMPOSE_MAJOR -ge 2 ]]; then
            check_pass "Docker Compose version $COMPOSE_VERSION >= 2.0"
        else
            check_fail "Docker Compose version $COMPOSE_VERSION < 2.0 (upgrade required)"
        fi
    else
        check_fail "Docker Compose not installed"
    fi
    
    # Check system resources
    CPU_CORES=$(nproc)
    if [[ $CPU_CORES -ge 8 ]]; then
        check_pass "CPU cores: $CPU_CORES >= 8"
    else
        check_warn "CPU cores: $CPU_CORES < 8 (recommended: 8+)"
    fi
    
    TOTAL_RAM=$(free -m | awk '/^Mem:/{print int($2/1024)}')
    if [[ $TOTAL_RAM -ge 16 ]]; then
        check_pass "RAM: ${TOTAL_RAM}GB >= 16GB"
    else
        check_warn "RAM: ${TOTAL_RAM}GB < 16GB (recommended: 16GB+)"
    fi
    
    DISK_SPACE=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [[ $DISK_SPACE -ge 100 ]]; then
        check_pass "Disk space: ${DISK_SPACE}GB >= 100GB"
    else
        check_warn "Disk space: ${DISK_SPACE}GB < 100GB (recommended: 100GB+)"
    fi
}

validate_repository_files() {
    print_header "Repository Files"
    
    # Essential scripts
    FILES=(
        "genesis-launch.sh"
        "docker-compose.yml"
        "docker-compose.monitoring.yml"
        "prometheus.yml"
        "alertmanager.yml"
    )
    
    for file in "${FILES[@]}"; do
        if [[ -f "$file" ]]; then
            check_pass "$file exists"
        else
            check_fail "$file missing"
        fi
    done
    
    # Check script is executable
    if [[ -x "genesis-launch.sh" ]]; then
        check_pass "genesis-launch.sh is executable"
    else
        check_warn "genesis-launch.sh not executable (run: chmod +x genesis-launch.sh)"
    fi
}

validate_docker_configs() {
    print_header "Docker Configurations"
    
    # Validate docker-compose files
    COMPOSE_FILES=(
        "docker-compose.yml"
        "docker-compose.monitoring.yml"
    )
    
    for file in "${COMPOSE_FILES[@]}"; do
        if docker compose -f "$file" config &> /dev/null; then
            check_pass "$file syntax valid"
        else
            check_fail "$file syntax invalid"
        fi
    done
    
    # Check Docker daemon
    if docker info &> /dev/null; then
        check_pass "Docker daemon running"
    else
        check_fail "Docker daemon not running"
    fi
    
    # Check Docker network
    if docker network ls &> /dev/null; then
        check_pass "Docker networking functional"
    else
        check_fail "Docker networking issues"
    fi
}

validate_monitoring_stack() {
    print_header "Monitoring Stack"
    
    # Check Prometheus config
    if [[ -f "prometheus.yml" ]]; then
        if grep -q "job_name.*sovereign" prometheus.yml; then
            check_pass "Prometheus configuration includes Sovereign jobs"
        else
            check_warn "Prometheus configuration may be incomplete"
        fi
    fi
    
    # Check Grafana dashboards
    DASHBOARDS=(
        "grafana/dashboards/genesis-launch-overview.json"
        "grafana/dashboards/network-performance-health.json"
        "grafana/dashboards/consensus-trust-monitoring.json"
    )
    
    for dashboard in "${DASHBOARDS[@]}"; do
        if [[ -f "$dashboard" ]]; then
            # Validate JSON syntax
            if jq empty "$dashboard" 2> /dev/null; then
                check_pass "$(basename \"$dashboard\") exists and valid JSON"
            else
                check_fail "$(basename \"$dashboard\") invalid JSON"
            fi
        else
            check_fail "$(basename \"$dashboard\") missing"
        fi
    done
    
    # Check Grafana provisioning
    if [[ -f "grafana/provisioning/dashboards/dashboards.yml" ]]; then
        if grep -q "Genesis" grafana/provisioning/dashboards/dashboards.yml; then
            check_pass "Grafana provisioning includes Genesis dashboards"
        else
            check_warn "Grafana provisioning may not include Genesis dashboards"
        fi
    else
        check_fail "Grafana provisioning configuration missing"
    fi
    
    # Check alert rules
    if [[ -f "tpm_alerts.yml" ]]; then
        if grep -q "groups:" tpm_alerts.yml; then
            check_pass "Alert rules configuration exists"
        else
            check_warn "Alert rules may be incomplete"
        fi
    else
        check_warn "Alert rules file missing"
    fi
}

validate_network_ports() {
    print_header "Network Ports"
    
    PORTS=(8000 8080 9090 3000 9093)
    
    for port in "${PORTS[@]}"; do
        if ! lsof -i :"$port" &> /dev/null && ! netstat -tuln 2>/dev/null | grep -q ":$port "; then
            check_pass "Port $port available"
        else
            check_warn "Port $port already in use"
        fi
    done
}

validate_go_packages() {
    print_header "Go Backend Packages"
    
    if command -v go &> /dev/null; then
        GO_PACKAGES=(
            "internal/api"
            "internal/consensus"
            "internal/convergence"
            "internal/island"
            "internal/monitoring"
            "internal/p2p"
            "internal/tpm"
        )
        
        for pkg in "${GO_PACKAGES[@]}"; do
            if [[ -d "$pkg" ]]; then
                if go build -o /dev/null "./$pkg" 2> /dev/null; then
                    check_pass "$pkg compiles successfully"
                else
                    check_fail "$pkg compilation failed"
                fi
            else
                check_warn "$pkg directory not found"
            fi
        done
    else
        check_warn "Go not installed (skipping package validation)"
    fi
}

validate_documentation() {
    print_header "Documentation"
    
    DOCS=(
        "README.md"
        "GENESIS_LAUNCH_GUIDE.md"
        "GENESIS_QUICK_START.md"
        "GENESIS_LAUNCH_CHECKLIST.md"
        "ARCHITECTURE.md"
    )
    
    for doc in "${DOCS[@]}"; do
        if [[ -f "$doc" ]]; then
            check_pass "$doc exists"
        else
            check_warn "$doc missing"
        fi
    done
}

validate_python_env() {
    print_header "Python Environment"
    
    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 --version | grep -oP '\d+\.\d+')
        PYTHON_MAJOR=$(echo "$PYTHON_VERSION" | cut -d. -f1)
        PYTHON_MINOR=$(echo "$PYTHON_VERSION" | cut -d. -f2)
        if [[ $PYTHON_MAJOR -gt 3 ]] || [[ $PYTHON_MAJOR -eq 3 && $PYTHON_MINOR -ge 11 ]]; then
            check_pass "Python version $PYTHON_VERSION >= 3.11"
        else
            check_warn "Python version $PYTHON_VERSION < 3.11 (recommended: 3.11+)"
        fi
        
        # Check key Python files
        PYTHON_FILES=(
            "sovereignmap_production_backend_v2.py"
            "tpm_cert_manager.py"
            "tpm_metrics_exporter.py"
        )
        
        for file in "${PYTHON_FILES[@]}"; do
            if [[ -f "$file" ]]; then
                if python3 -m py_compile "$file" 2> /dev/null; then
                    check_pass "$file syntax valid"
                else
                    check_fail "$file syntax error"
                fi
            else
                check_warn "$file not found"
            fi
        done
    else
        check_warn "Python3 not installed"
    fi
}

################################################################################
# Main Execution
################################################################################

main() {
    clear
    echo -e "${BLUE}"
    cat << "EOF"
    ╔═══════════════════════════════════════════════════════════════╗
    ║                                                               ║
    ║         🚀 Genesis Block Launch Validation 🚀                 ║
    ║                                                               ║
    ║         Sovereign Map Federated Learning v1.0.0              ║
    ║                                                               ║
    ╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}\n"
    
    # Run all validations
    validate_system_requirements
    validate_repository_files
    validate_docker_configs
    validate_monitoring_stack
    validate_network_ports
    validate_go_packages
    validate_python_env
    validate_documentation
    
    # Summary
    print_header "Validation Summary"
    
    TOTAL=$((PASSED + FAILED + WARNINGS))
    
    echo -e "${GREEN}Passed:${NC}   $PASSED / $TOTAL"
    echo -e "${RED}Failed:${NC}   $FAILED / $TOTAL"
    echo -e "${YELLOW}Warnings:${NC} $WARNINGS / $TOTAL"
    
    echo ""
    
    if [[ $FAILED -eq 0 ]]; then
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}✓ System is READY for Genesis Block Launch!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "Run: ${BLUE}./genesis-launch.sh${NC} to begin launch"
        echo ""
        exit 0
    else
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}✗ System NOT ready for launch${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "Please fix the ${RED}$FAILED${NC} failed check(s) above"
        echo ""
        exit 1
    fi
}

# Run main function
main "$@"
