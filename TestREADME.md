# Quick start (recommended)
make test-200-bft

# Or step-by-step:
make setup-200-test              # Generate data
docker-compose -f docker-compose.200nodes.yml up -d mongo backend aggregator
sleep 30
docker-compose -f docker-compose.200nodes.yml up -d node-agent  # Deploy 200 nodes
sleep 90
go test -v ./internal/consensus/... -run "Test200NodeBFT" -timeout 60m
make clean-200-test              # Cleanup
