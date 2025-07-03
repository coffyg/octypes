# OcTypes Benchmark Guide

This directory contains comprehensive benchmarks for the OcTypes library, comparing the original implementation (v0.0.4) with the current optimized version.

## Available Benchmarks

The benchmark suite tests:

1. **Serialization Performance**
   - JSON marshaling/unmarshaling for all types
   - Binary serialization/deserialization 
   - NULL value handling

2. **Memory Layout Impact** 
   - Memory usage for struct arrays
   - Field alignment optimizations

3. **String Operations**
   - String interning in maps
   - Large string allocation efficiency

4. **Batch Operations**
   - Performance with multiple objects
   - Memory allocations under load

## Running the Benchmarks

To run all benchmarks and generate a comparison report:

```bash
./run_benchmarks.sh
```

This script:
1. Runs benchmarks for v0.0.4
2. Runs benchmarks for the current optimized version
3. Generates a detailed comparison report in `benchmark_results/benchmark_report.md`

## Understanding the Results

The benchmark report includes:

- Operations per second (higher is better)
- Memory allocation per operation (lower is better)
- Percentage improvements in both performance and memory usage

## Individual Benchmark Execution

You can also run specific benchmarks using:

```bash
cd benchmark
go test -bench=BenchmarkNullStringMarshalJSON -benchmem ./...
```

## Expected Improvements

The optimized version should show:
- 40-70% faster JSON processing
- 6-9x faster binary serialization
- Reduced memory allocations
- Significant improvements for map operations with string interning