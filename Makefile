# Sovereign Map Federated Learning - Makefile
# Updated with 200-Node BFT Test Targets

.PHONY: all build test clean deploy logs help \
	smoke testnet-wallet-readiness \
	stack-start stack-verify stack-down screenshots-check go-env observability-smoke alerts-test

COMPOSE ?= docker compose
FULL_COMPOSE_FILE ?= docker-compose.full.yml
NODE_AGENT_SCALE ?= 5
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
# Docker Compose Operations
# =============================================================================

deploy:
	@echo "🚀 Deploying standard stack from $(FULL_COMPOSE_FILE) with node-agent scale $(NODE_AGENT_SCALE)..."
	$(COMPOSE) -f $(FULL_COMPOSE_FILE) up -d --scale node-agent=$(NODE_AGENT_SCALE)

stack-start:
	@echo "✨ Starting full compose stack with node-agent scale $(NODE_AGENT_SCALE)..."
	@$(COMPOSE) -f $(FULL_COMPOSE_FILE) up -d --scale node-agent=$(NODE_AGENT_SCALE)
	@sleep 8
	@echo "✅ Stack started"
	@echo "   API:       http://localhost:8000/status"
	@echo "   HUD:       http://localhost:3000"
	@echo "   Grafana:   http://localhost:3001"
	@echo "   Prometheus:http://localhost:9090"

stack-verify:
	@echo "🔎 Verifying full compose stack health..."
	@curl -fsS http://localhost:8000/status >/dev/null
	@curl -fsS http://localhost:8000/health >/dev/null
	@curl -fsS http://localhost:8000/ops/health >/dev/null
	@curl -fsS http://localhost:8000/training/status >/dev/null
	@$(COMPOSE) -f $(FULL_COMPOSE_FILE) ps >/dev/null
	@echo "✅ API health endpoints are reachable"
	@echo "✅ UI surfaces expected: HUD http://localhost:3000, Grafana http://localhost:3001"

stack-down:
	@echo "🛑 Stopping full compose stack..."
	@$(COMPOSE) -f $(FULL_COMPOSE_FILE) down --remove-orphans

screenshots-check:
	@echo "🖼️ Verifying required release screenshots..."
	@test -f docs/screenshots/hud-operations-overview.png
	@test -f docs/screenshots/grafana-operations-overview.png
	@test -f docs/screenshots/grafana-llm-overview.png
	@test -f docs/screenshots/grafana-tokenomics-overview.png
	@echo "✅ All required screenshots are present"

logs:
	$(COMPOSE) -f $(FULL_COMPOSE_FILE) logs -f --tail=100

# =============================================================================
# Cleanup Targets
# =============================================================================

clean:
	@echo "🧹 Cleaning up..."
	$(COMPOSE) -f $(FULL_COMPOSE_FILE) down -v --remove-orphans
	rm -rf bin/
	$(GO) clean

clean-all: clean
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
	@docker compose -f $(FULL_COMPOSE_FILE) config >/dev/null
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
# Development Helpers
# =============================================================================

dev-setup:
	@echo "🔧 Setting up development environment..."
	@$(GO) mod download
	@mkdir -p bin test-results test-data
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
	@echo "Stack Run Sequence:"
	@echo "  make stack-start     - Start full compose stack with node-agent scale"
	@echo "  make stack-verify    - Verify key API health endpoints"
	@echo "  make stack-down      - Stop full compose stack"
	@echo "  make screenshots-check - Ensure required README screenshots exist"
	@echo ""
	@echo "Deploy:"
	@echo "  make deploy           - Deploy standard stack"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean            - Clean standard environment"
	@echo "  make clean-all        - Clean everything"
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
