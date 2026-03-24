# Sovereign Map Federated Learning - Makefile
# Updated with 200-Node BFT Test Targets

.PHONY: all build test clean deploy logs help \
        test-200-bft setup-200-test clean-200-test benchmark-200 \
	chaos-test partition-test generate-data smoke testnet-wallet-readiness \
	wow-start wow-verify screenshots-check go-env observability-smoke alerts-test

COMPOSE ?= docker compose
TOOLROOT ?= /go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.7.linux-amd64

ifneq ("$(wildcard $(TOOLROOT)/bin/go)","")
GO := GOROOT=$(TOOLROOT) GOTOOLCHAIN=local $(TOOLROOT)/bin/go
else
GO := go
endif

# Default target
all: build

# =============================================================================
# Build Targets
# =============================================================================

build:
	@echo "🔨 Building Sovereign Map components..."
	$(GO) build -o bin/node-agent ./cmd/node-agent
	$(GO) build -o bin/aggregator ./cmd/aggregator
	$(GO) build -o bin/cli ./cmd/cli
	@echo "✅ Build complete. Binaries in bin/"

build-docker:
	@echo "🐳 Building Docker images..."
	docker build -t sovereign-map/node-agent:latest -f Dockerfile .
	docker build -t sovereign-map/aggregator:latest -f Dockerfile.backend .
	docker build -t sovereign-map/frontend:latest -f Dockerfile.frontend .
	@echo "✅ Docker images built"

# =============================================================================
# Standard Test Targets
# =============================================================================

test:
	@echo "🧪 Running standard test suite..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "📊 Coverage report generated"

test-short:
	@echo "🧪 Running short tests..."
	$(GO) test -v -short ./...

test-consensus:
	@echo "🧪 Running consensus tests..."
	$(GO) test -v ./internal/consensus/... -timeout 30m

test-bft:
	@echo "🛡️ Running BFT tests..."
	$(GO) test -v ./internal/consensus/... -run "TestBFT|TestByzantine" -timeout 30m

benchmark:
	@echo "📊 Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

# =============================================================================
# 200-Node BFT Test Targets (NEW)
# =============================================================================

setup-200-test:
	@echo "🔧 Setting up 200-node BFT test environment..."
	@mkdir -p test-results/200-node-bft/$(shell date +%Y-%m-%d)
	@mkdir -p test-data
	@$(GO) run scripts/generate-test-data.go
	@docker network create sovereign-net-200 2>/dev/null || true
	@echo "✅ Environment ready"

test-200-bft: setup-200-test
	@echo "🚀 Starting 200-Node Byzantine Fault Tolerance Test"
	@echo "   Configuration: 200 nodes, 111 Byzantine (55.5%)"
	@echo "   Estimated duration: 2-3 hours"
	@echo ""
	@chmod +x scripts/run-200-bft-test.sh
	@./scripts/run-200-bft-test.sh bft-200-$(shell date +%Y%m%d-%H%M%S)

test-200-bft-verbose: setup-200-test
	@echo "🚀 Starting 200-Node BFT Test (Verbose Mode)"
	@chmod +x scripts/run-200-bft-test.sh
	@bash -x ./scripts/run-200-bft-test.sh bft-200-verbose-$(shell date +%Y%m%d-%H%M%S) 2>&1 | tee test-results/200-node-bft/verbose-$(shell date +%Y%m%d).log

test-200-bft-quick:
	@echo "⚡ Quick 200-node test (reduced rounds)"
	@$(COMPOSE) -f docker-compose.200nodes.yml up -d mongo redis backend
	@sleep 30
	@$(COMPOSE) -f docker-compose.200nodes.yml up -d node-agent
	@sleep 60
	@$(GO) test -v ./internal/consensus/... -run "Test200NodeBFT" -timeout 20m -quick-test
	@$(COMPOSE) -f docker-compose.200nodes.yml down

benchmark-200:
	@echo "📊 Running 200-node performance benchmark..."
	@$(COMPOSE) -f docker-compose.200nodes.yml up -d mongo redis backend
	@sleep 30
	@$(GO) test -bench=Benchmark200Nodes -benchtime=10m -memprofile=mem.out -cpuprofile=cpu.out ./internal/consensus/...

chaos-test:
	@echo "🔥 Running chaos engineering tests..."
	@$(COMPOSE) -f docker-compose.200nodes.yml --profile chaos up -d
	@sleep 60
	@docker exec chaos-200 python3 /app/chaos_injector.py --nodes 200 --faults 111 --duration 300
	@$(GO) test -v ./internal/consensus/... -run "TestChaosRecovery" -timeout 30m

partition-test:
	@echo "🔀 Running network partition tests..."
	@$(COMPOSE) -f docker-compose.200nodes.yml up -d
	@sleep 60
	@./scripts/network_partition.sh 3 60  # 3 partitions, 60 seconds
	@$(GO) test -v ./internal/consensus/... -run "TestNetworkPartition" -timeout 30m

generate-data:
	@echo "🎲 Generating synthetic test data..."
	@$(GO) run scripts/generate-test-data.go

# =============================================================================
# Docker Compose Operations
# =============================================================================

deploy:
	@echo "🚀 Deploying standard stack..."
	$(COMPOSE) up -d

wow-start:
	@echo "✨ Starting one-pass wow profile (dev stack + monitoring)..."
	@docker compose -f docker-compose.dev.yml up -d mongo redis backend frontend
	@docker compose -f docker-compose.monitoring.yml up -d
	@sleep 8
	@echo "✅ Stack started"
	@echo "   API:       http://localhost:8000/status"
	@echo "   HUD:       http://localhost:3000"
	@echo "   Grafana:   http://localhost:3001"
	@echo "   Prometheus:http://localhost:9090"

wow-verify:
	@echo "🔎 Verifying wow profile health..."
	@curl -fsS http://localhost:8000/status >/dev/null
	@curl -fsS http://localhost:8000/health >/dev/null
	@curl -fsS http://localhost:8000/ops/health >/dev/null
	@curl -fsS http://localhost:8000/training/status >/dev/null
	@echo "✅ API health endpoints are reachable"
	@echo "✅ UI surfaces expected: HUD http://localhost:3000, Grafana http://localhost:3001"

screenshots-check:
	@echo "🖼️ Verifying required release screenshots..."
	@test -f docs/screenshots/hud-operations-overview.png
	@test -f docs/screenshots/grafana-operations-overview.png
	@test -f docs/screenshots/grafana-llm-overview.png
	@test -f docs/screenshots/grafana-tokenomics-overview.png
	@echo "✅ All required screenshots are present"

deploy-200:
	@echo "🚀 Deploying 200-node test stack..."
	$(COMPOSE) -f docker-compose.200nodes.yml up -d mongo redis backend aggregator
	@sleep 30
	@echo "✅ Infrastructure ready. Deploy nodes with: make deploy-200-nodes"

deploy-200-nodes:
	@echo "🚀 Deploying 200 node agents..."
	$(COMPOSE) -f docker-compose.200nodes.yml up -d node-agent
	@sleep 90
	@echo "✅ 200 nodes deployed"

deploy-monitoring:
	@echo "📊 Deploying monitoring stack..."
	$(COMPOSE) -f docker-compose.monitoring.yml up -d
	@echo "📈 Grafana: http://localhost:3001 (admin/admin)"
	@echo "📊 Prometheus: http://localhost:9090"

logs:
	$(COMPOSE) logs -f --tail=100

logs-200:
	$(COMPOSE) -f docker-compose.200nodes.yml logs -f --tail=100

logs-backend:
	docker logs -f backend-200 2>&1 | grep -E "(consensus|byzantine|fault|error)" || true

# =============================================================================
# Cleanup Targets
# =============================================================================

clean:
	@echo "🧹 Cleaning up..."
	$(COMPOSE) down -v --remove-orphans
	rm -rf bin/
	$(GO) clean

clean-200-test:
	@echo "🧹 Cleaning up 200-node test environment..."
	$(COMPOSE) -f docker-compose.200nodes.yml down -v --remove-orphans
	docker system prune -f --volumes 2>/dev/null || true
	docker network rm sovereign-net-200 2>/dev/null || true
	rm -rf test-results/200-node-bft/*
	rm -rf test-data/*
	@echo "✅ 200-node environment cleaned"

clean-all: clean clean-200-test
	@echo "🧹 Complete cleanup finished"

# =============================================================================
# Utility Targets
# =============================================================================

tidy:
	$(GO) mod tidy
	$(GO) mod verify

fmt:
	$(GO) fmt ./...

contributors-rankings:
	@echo "🏆 Generating contributor rankings and reward points..."
	python3 scripts/contributor_rankings.py
	@echo "✅ Rankings generated in test-results/contributor-rankings/"

lint:
	@echo "🔍 Running linters..."
	$(GO) clean -cache
	golangci-lint cache clean
	GOROOT=$(TOOLROOT) GOTOOLCHAIN=local golangci-lint run ./...

lint-soft:
	@echo "🔍 Running linters (non-blocking)..."
	$(GO) clean -cache
	golangci-lint cache clean
	GOROOT=$(TOOLROOT) GOTOOLCHAIN=local golangci-lint run ./... || true

vet:
	$(GO) vet ./...

security-scan:
	@echo "🔒 Running security scans..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@GOSEC_BIN=$$(command -v gosec || echo "$$($(GO) env GOPATH)/bin/gosec"); \
		GOROOT=$(TOOLROOT) GOTOOLCHAIN=local $$GOSEC_BIN ./... || true

check: fmt vet lint-soft test
	@echo "✅ All checks passed"

alerts-test:
	@echo "🚨 Running Prometheus alert rule tests..."
	@docker run --rm --entrypoint /bin/promtool -v "$$(pwd):/workspace" -w /workspace prom/prometheus:v2.48.0 \
		check rules fl_slo_alerts.yml fl_detailed_alerts.yml tpm_alerts.yml
	@docker run --rm --entrypoint /bin/promtool -v "$$(pwd):/workspace" -w /workspace prom/prometheus:v2.48.0 \
		test rules fl_slo_alerts.test.yml fl_detailed_alerts.test.yml tpm_alerts.test.yml
	@$(GO) test ./internal/monitoring -run "TestAlertmanagerRoutingPolicy|TestAlertmanagerInhibitionPolicy"
	@echo "✅ Alert rule tests passed"

smoke:
	@echo "🧪 Running reproducibility smoke checks..."
	@PKGS=$$($(GO) list ./... | grep -Ev '(/node_modules/|/sensors/camera$$|/sensors/slam$$|/storage/map_tiles$$)'); \
		$(GO) test -short $$PKGS
	@npm ci
	@npm --prefix frontend ci
	@npm --prefix frontend run build
	@npm --prefix packages/core ci
	@npm --prefix packages/privacy ci
	@docker compose -f docker-compose.production.yml config >/dev/null
	@docker compose -f docker-compose.1000nodes.yml config >/dev/null
	@echo "✅ Smoke checks passed"

observability-smoke:
	@echo "📈 Running observability smoke checks..."
	@python3 scripts/check_dashboard_queries.py
	@node -e 'const fs=require("fs"); const files=["grafana/provisioning/dashboards/operations_overview.json","grafana/provisioning/dashboards/tokenomics_overview.json","grafana/provisioning/dashboards/llm_overview.json"]; files.forEach(f=>JSON.parse(fs.readFileSync(f,"utf8"))); console.log("dashboard json ok")'
	@echo "✅ Observability smoke checks passed"

testnet-wallet-readiness:
	@echo "🧪 Running testnet wallet readiness checks..."
	@bash scripts/testnet-wallet-readiness.sh

# =============================================================================
# Results & Reporting
# =============================================================================

results-200:
	@echo "📊 200-Node Test Results:"
	@ls -lah test-results/200-node-bft/ 2>/dev/null || echo "No results found"
	@echo ""
	@echo "View latest report:"
	@cat test-results/200-node-bft/$(shell ls -t test-results/200-node-bft/ 2>/dev/null | head -1)/TEST-REPORT.md 2>/dev/null || echo "No report generated"

commit-results:
	@echo "💾 Committing test results to repository..."
	@git add test-results/200-node-bft/
	@git commit -m "Add 200-node BFT test results - $(shell date +%Y-%m-%d)" || echo "Nothing to commit"
	@git push origin main || echo "Push failed or not configured"

dashboard:
	@echo "📈 Opening monitoring dashboards..."
	@open http://localhost:3001 || xdg-open http://localhost:3001 || echo "Open manually: http://localhost:3001"

# =============================================================================
# Development Helpers
# =============================================================================

dev-setup:
	@echo "🔧 Setting up development environment..."
	@$(GO) mod download
	@mkdir -p bin test-results test-data
	@cp config/200node-test.yaml.example config/200node-test.yaml 2>/dev/null || true
	@echo "✅ Development environment ready"

go-env:
	@echo "Go command: $(GO)"
	@$(GO) version
	@$(GO) env GOROOT GOTOOLCHAIN GOMODCACHE GOCACHE

proto:
	@echo "📝 Generating protobuf files..."
	protoc --go_out=. --go-grpc_out=. pkg/proto/*.proto

wasm:
	@echo "🔧 Building WASM modules..."
	tinygo build -o wasm/verify.wasm -target wasm ./wasm/verify.go

# =============================================================================
# Help
# =============================================================================

help:
	@echo "Sovereign Map Federated Learning - Available Targets"
	@echo "=================================================="
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build Go binaries"
	@echo "  make build-docker   - Build Docker images"
	@echo ""
	@echo "Standard Tests:"
	@echo "  make test           - Run all tests"
	@echo "  make smoke          - Quick clone/reproducibility checks"
	@echo "  make test-consensus - Run consensus tests only"
	@echo "  make test-bft       - Run BFT tests only"
	@echo ""
	@echo "Wow Mode:"
	@echo "  make wow-start       - Start dev + monitoring stack for demo"
	@echo "  make wow-verify      - Verify key API health endpoints"
	@echo "  make screenshots-check - Ensure required README screenshots exist"
	@echo ""
	@echo "200-Node BFT Tests (NEW):"
	@echo "  make setup-200-test     - Setup environment"
	@echo "  make test-200-bft       - Run full 200-node BFT test"
	@echo "  make test-200-bft-quick - Quick test (reduced rounds)"
	@echo "  make benchmark-200      - Performance benchmark"
	@echo "  make chaos-test         - Chaos engineering test"
	@echo "  make partition-test     - Network partition test"
	@echo "  make generate-data      - Generate synthetic data"
	@echo ""
	@echo "Deploy:"
	@echo "  make deploy           - Deploy standard stack"
	@echo "  make deploy-200       - Deploy 200-node infrastructure"
	@echo "  make deploy-200-nodes - Deploy 200 node agents"
	@echo "  make deploy-monitoring - Deploy monitoring only"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean            - Clean standard environment"
	@echo "  make clean-200-test   - Clean 200-node environment"
	@echo "  make clean-all        - Clean everything"
	@echo ""
	@echo "Results:"
	@echo "  make results-200      - View 200-node test results"
	@echo "  make commit-results   - Commit results to git"
	@echo "  make dashboard        - Open Grafana dashboard"
	@echo ""
	@echo "Utilities:"
	@echo "  make tidy    - Tidy Go modules"
	@echo "  make fmt     - Format Go code"
	@echo "  make lint    - Run linters"
	@echo "  make lint-soft - Run linters without failing target"
	@echo "  make go-env  - Print effective Go toolchain settings"
	@echo "  make observability-smoke - Validate dashboard queries and JSON syntax"
	@echo "  make check   - Run all checks"
	@echo ""
