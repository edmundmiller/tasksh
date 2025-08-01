#!/bin/bash

# Script to run all benchmarks and generate a summary report

echo "Running Tasksh Benchmarks"
echo "========================"
echo ""

# Create output directory
mkdir -p bench_results

# Run benchmarks for each package
packages=(
    "internal/planning"
    "internal/preview" 
    "internal/review"
    "internal/taskwarrior"
)

for pkg in "${packages[@]}"; do
    echo "Benchmarking $pkg..."
    output_file="bench_results/$(echo $pkg | tr '/' '_').txt"
    go test -bench=. -benchmem "./$pkg" -run=^$ > "$output_file" 2>&1
    
    if [ $? -eq 0 ]; then
        echo "✓ Completed $pkg"
        # Show summary of key benchmarks
        echo "  Key results:"
        grep -E "^Benchmark.*-[0-9]+\s+" "$output_file" | head -5 | sed 's/^/    /'
    else
        echo "✗ Failed $pkg"
    fi
    echo ""
done

echo "Generating summary report..."

# Create summary report
cat > bench_results/SUMMARY.md << EOF
# Tasksh Benchmark Summary

Generated on: $(date)

## Key Performance Metrics

### Planning UI Performance
EOF

# Extract key metrics
if [ -f "bench_results/internal_planning.txt" ]; then
    echo "" >> bench_results/SUMMARY.md
    echo '```' >> bench_results/SUMMARY.md
    grep -E "^BenchmarkPlanningModelView|^BenchmarkVisualWidth.*Simple|^BenchmarkCachedVisualWidth" bench_results/internal_planning.txt >> bench_results/SUMMARY.md
    echo '```' >> bench_results/SUMMARY.md
fi

cat >> bench_results/SUMMARY.md << EOF

### Preview/Mock Performance
EOF

if [ -f "bench_results/internal_preview.txt" ]; then
    echo "" >> bench_results/SUMMARY.md
    echo '```' >> bench_results/SUMMARY.md
    grep -E "^BenchmarkMockViewRender.*MainView_80x24|^BenchmarkGeneratePreview" bench_results/internal_preview.txt >> bench_results/SUMMARY.md
    echo '```' >> bench_results/SUMMARY.md
fi

cat >> bench_results/SUMMARY.md << EOF

### Review UI Performance
EOF

if [ -f "bench_results/internal_review.txt" ]; then
    echo "" >> bench_results/SUMMARY.md
    echo '```' >> bench_results/SUMMARY.md
    grep -E "^BenchmarkReviewModelView|^BenchmarkRenderTask.*Simple" bench_results/internal_review.txt >> bench_results/SUMMARY.md
    echo '```' >> bench_results/SUMMARY.md
fi

echo ""
echo "Complete benchmark results saved in bench_results/"
echo "Summary report: bench_results/SUMMARY.md"