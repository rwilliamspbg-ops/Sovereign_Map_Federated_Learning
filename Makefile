# 200-Node BFT Test Targets

.PHONY: test-200-bft setup-200-test clean-200-test

setup-200-test:
	@echo "Setting up 200-node test environment..."
	mkdir -p test-results test-data
	go run scripts/generate-test-data.go
	docker network create sovereign-net || true

test-200-bft: setup-200-test
	@echo "🧪 Running 200-Node BFT Test (Duration: ~2 hours)"
	./scripts/run-200-bft-test.sh

clean-200-test:
	@echo "Cleaning up 200-node test environment..."
	docker-compose -f docker-compose.200nodes.yml down -v
	docker system prune -f
	rm -rf test-results/* test-data/*

benchmark-200:
	@echo "📊 Running 200-node performance benchmark..."
	go test -bench=Benchmark200Nodes -benchtime=10m ./internal/consensus/...

# Original targets preserved
test:
	go test -v ./...

build:
	go build -o bin/node-agent cmd/node-agent/main.go

deploy:
	docker-compose up -d

logs:
	docker-compose logs -f

clean:
	docker-compose down -v
	rm -rf bin/
