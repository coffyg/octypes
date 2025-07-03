#!/usr/bin/env bash

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
TEMP_DIR="/tmp/octypes_comparison"
BENCHMARK_FILE="$TEMP_DIR/benchmark_suite_test.go"

# Functions to run benchmarks and save results
run_benchmarks() {
  local version=$1
  local output_file=$2
  
  echo -e "${GREEN}Running benchmarks for $version...${NC}"
  cd "$CURRENT_DIR"
  
  # If we're testing a specific version, check it out first
  if [ "$version" != "current" ]; then
    git checkout $version
  fi
  
  # Copy benchmark file to current directory
  cp "$BENCHMARK_FILE" .
  
  # Run benchmarks with memory stats and save results
  go test -bench=. -benchmem ./... > "$output_file" 2>&1
  
  echo -e "${GREEN}Benchmarks for $version completed. Results saved to $output_file${NC}"
  echo
}

# Parse and compare benchmark results
compare_results() {
  local old_file=$1
  local new_file=$2
  local report_file=$3
  
  echo -e "${BLUE}Comparing benchmark results...${NC}" | tee -a "$report_file"
  echo | tee -a "$report_file"
  
  # Create header for the comparison table
  echo -e "| Benchmark | v0.0.4 (ops/s) | Current (ops/s) | Improvement % | v0.0.4 (B/op) | Current (B/op) | Memory Improvement % |" | tee -a "$report_file"
  echo -e "|-----------|---------------|----------------|--------------|--------------|---------------|---------------------|" | tee -a "$report_file"
  
  # Extract benchmark names from new results
  benchmarks=$(grep 'Benchmark' "$new_file" | awk '{print $1}' | sort)
  
  for bench in $benchmarks; do
    # Skip if it's not a real benchmark
    if [[ "$bench" == "Benchmark" ]]; then
      continue
    fi
    
    # Get old and new benchmark results
    old_result=$(grep "$bench " "$old_file" 2>/dev/null || echo "Not found")
    new_result=$(grep "$bench " "$new_file" 2>/dev/null || echo "Not found")
    
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
      
      # Use color coding based on performance
      if (( $(echo "$perf_improvement > 5" | bc -l) )); then
        perf_color="${GREEN}"
      elif (( $(echo "$perf_improvement < -5" | bc -l) )); then
        perf_color="${RED}"
      else
        perf_color="${NC}"
      fi
      
      if [[ "$mem_sign" == "+" && "$mem_improvement" != "N/A" ]]; then
        mem_color="${GREEN}"
      elif [[ "$mem_sign" == "-" && "$mem_improvement" != "N/A" ]]; then
        mem_color="${RED}"
      else
        mem_color="${NC}"
      fi
      
      # Add to report
      if [[ "$perf_improvement" != "N/A" && "$mem_improvement" != "N/A" ]]; then
        echo -e "| $readable_name | $old_ops | $new_ops | ${perf_color}${perf_sign}${perf_improvement}%${NC} | $old_memory | $new_memory | ${mem_color}${mem_sign}${mem_improvement}%${NC} |" | tee -a "$report_file"
      else
        echo -e "| $readable_name | $old_ops | $new_ops | N/A | $old_memory | $new_memory | N/A |" | tee -a "$report_file"
      fi
    else
      # Missing data for comparison
      echo -e "| $bench | Not found | Not found | N/A | N/A | N/A | N/A |" | tee -a "$report_file"
    fi
  done
  
  echo | tee -a "$report_file"
  echo -e "${GREEN}Benchmark comparison completed. Results saved to $report_file${NC}"
}

# Create output directories
mkdir -p "$TEMP_DIR/results"

# Files for storing results
OLD_RESULTS="$TEMP_DIR/results/v0.0.4_results.txt"
NEW_RESULTS="$TEMP_DIR/results/current_results.txt"
REPORT="$TEMP_DIR/results/benchmark_report.md"

# Initialize report file
echo "# OcTypes Benchmark Comparison Report" > "$REPORT"
echo "Comparing v0.0.4 vs latest optimized version" >> "$REPORT"
echo >> "$REPORT"
echo "## Benchmark Results" >> "$REPORT"
echo >> "$REPORT"

# Run benchmarks for v0.0.4
run_benchmarks "6796e4ad47623f6bde03c543a8f9698f04bc2103" "$OLD_RESULTS"

# Run benchmarks for current version
run_benchmarks "current" "$NEW_RESULTS"

# Compare results
compare_results "$OLD_RESULTS" "$NEW_RESULTS" "$REPORT"

# Return to the original branch
git checkout main

echo -e "${BLUE}=== Benchmark Testing Complete ===${NC}"
echo -e "${BLUE}See detailed report at: $REPORT${NC}"