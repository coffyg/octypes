#!/usr/bin/env bash

# This script runs the benchmark tests for both the old version (v0.0.4)
# and the current optimized version, then compares the results

set -e

# Color codes for pretty output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== OcTypes Benchmark Comparison ===${NC}"
echo -e "${BLUE}Comparing v0.0.4 vs latest optimized version${NC}"
echo

# Get the current directory
CURRENT_DIR=$(pwd)
RESULTS_DIR="$CURRENT_DIR/benchmark_results"
mkdir -p "$RESULTS_DIR"

# Save current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Files for storing results
OLD_RESULTS="$RESULTS_DIR/v0.0.4_results.txt"
NEW_RESULTS="$RESULTS_DIR/current_results.txt"
REPORT="$RESULTS_DIR/benchmark_report.md"

# Initialize report file
echo "# OcTypes Benchmark Comparison Report" > "$REPORT"
echo "Comparing v0.0.4 vs latest optimized version" >> "$REPORT"
echo >> "$REPORT"
echo "## Benchmark Results" >> "$REPORT"
echo >> "$REPORT"

echo -e "${GREEN}Running benchmarks for v0.0.4...${NC}"
# Checkout v0.0.4
git checkout 6796e4ad47623f6bde03c543a8f9698f04bc2103
# Run benchmarks and save results
cd benchmark
go test -bench=. -benchmem ./... > "$OLD_RESULTS" 2>&1
cd ..

echo -e "${GREEN}Running benchmarks for current optimized version...${NC}"
# Checkout the current branch (with optimizations)
git checkout main
# Run benchmarks and save results
cd benchmark
go test -bench=. -benchmem ./... > "$NEW_RESULTS" 2>&1
cd ..

# Compare results
echo -e "${BLUE}Comparing benchmark results...${NC}" | tee -a "$REPORT"
echo | tee -a "$REPORT"

# Create header for the comparison table
echo -e "| Benchmark | v0.0.4 (ops/s) | Current (ops/s) | Improvement % | v0.0.4 (B/op) | Current (B/op) | Memory Improvement % |" | tee -a "$REPORT"
echo -e "|-----------|---------------|----------------|--------------|--------------|---------------|---------------------|" | tee -a "$REPORT"

# Extract benchmark names from new results
benchmarks=$(grep 'Benchmark' "$NEW_RESULTS" | awk '{print $1}' | sort)

for bench in $benchmarks; do
  # Skip if it's not a real benchmark
  if [[ "$bench" == "Benchmark" ]]; then
    continue
  fi
  
  # Get old and new benchmark results
  old_result=$(grep "$bench " "$OLD_RESULTS" 2>/dev/null || echo "Not found")
  new_result=$(grep "$bench " "$NEW_RESULTS" 2>/dev/null || echo "Not found")
  
  # If both results exist, compare them
  if [[ "$old_result" != "Not found" && "$new_result" != "Not found" ]]; then
    # Extract ops/s, time/op, and memory allocated
    old_ops=$(echo "$old_result" | awk '{print $3}')
    new_ops=$(echo "$new_result" | awk '{print $3}')
    
    old_memory=$(echo "$old_result" | awk '{print $5}')
    new_memory=$(echo "$new_result" | awk '{print $5}')
    
    # Calculate improvement percentage
    if [[ "$old_ops" =~ ^[0-9]+(\.[0-9]+)?$ && "$new_ops" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
      perf_improvement=$(echo "scale=2; (($new_ops - $old_ops) / $old_ops) * 100" | bc)
      if (( $(echo "$perf_improvement >= 0" | bc -l) )); then
        perf_sign="+"
      else
        perf_sign=""
      fi
    else
      perf_improvement="N/A"
      perf_sign=""
    fi
    
    # Calculate memory improvement percentage
    if [[ "$old_memory" =~ ^[0-9]+$ && "$new_memory" =~ ^[0-9]+$ ]]; then
      mem_improvement=$(echo "scale=2; (($old_memory - $new_memory) / $old_memory) * 100" | bc)
      if (( $(echo "$mem_improvement >= 0" | bc -l) )); then
        mem_sign="+"
      else
        mem_sign="-"
        # Convert negative to positive for display
        mem_improvement=$(echo "scale=2; -1 * $mem_improvement" | bc)
      fi
    else
      mem_improvement="N/A"
      mem_sign=""
    fi
    
    # Format the benchmark name to be more readable
    readable_name=$(echo "$bench" | sed 's/Benchmark//g')
    
    # Add to report
    if [[ "$perf_improvement" != "N/A" && "$mem_improvement" != "N/A" ]]; then
      echo -e "| $readable_name | $old_ops | $new_ops | ${perf_sign}${perf_improvement}% | $old_memory | $new_memory | ${mem_sign}${mem_improvement}% |" | tee -a "$REPORT"
    else
      echo -e "| $readable_name | $old_ops | $new_ops | N/A | $old_memory | $new_memory | N/A |" | tee -a "$REPORT"
    fi
  else
    # Missing data for comparison
    echo -e "| $bench | Not found | Not found | N/A | N/A | N/A | N/A |" | tee -a "$REPORT"
  fi
done

# Return to the original branch
git checkout "$CURRENT_BRANCH"

echo
echo -e "${GREEN}Benchmark comparison completed. Results saved to $REPORT${NC}"
echo -e "${BLUE}=== Benchmark Testing Complete ===${NC}"