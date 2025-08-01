.PHONY: help build test bench bench-quick bench-full clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build tasksh binary
	go build -o tasksh ./cmd/tasksh

test: ## Run all tests
	go test ./...

bench-quick: ## Run quick benchmarks (key performance metrics only)
	@echo "Running quick benchmarks..."
	@echo "Planning Model View:"
	@go test -bench=BenchmarkPlanningModelView -benchmem ./internal/planning -run=^$ 2>&1 | grep -E "BenchmarkPlanningModelView" || true
	@echo "Visual Width (Simple):"
	@go test -bench=BenchmarkVisualWidthPerf/Simple -benchmem ./internal/planning -run=^$ 2>&1 | grep -E "Simple" || true
	@echo "Mock View Render (Main):"
	@go test -bench=BenchmarkMockViewRender/MainView_80x24 -benchmem ./internal/preview -run=^$ 2>&1 | grep -E "MainView_80x24" || true

bench: ## Run standard benchmarks
	@echo "Running planning benchmarks..."
	go test -bench=. -benchmem ./internal/planning -run=^$ -benchtime=10s
	@echo ""
	@echo "Running preview benchmarks..."
	go test -bench=. -benchmem ./internal/preview -run=^$ -benchtime=10s
	@echo ""
	@echo "Running review benchmarks..."
	go test -bench=. -benchmem ./internal/review -run=^$ -benchtime=10s

bench-full: ## Run all benchmarks with full report
	./scripts/run_benchmarks.sh

clean: ## Clean build artifacts
	rm -f tasksh
	rm -f cpu.prof
	rm -rf bench_results/
	rm -f cmd/profile_planning/profile_planning
	rm -f cmd/bench_compare/bench_compare
	rm -f cmd/startup_bench/startup_bench