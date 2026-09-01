#!/bin/bash
set -e

# Go Benchmark Script
# Runs benchmarks with optional profiling

PROJECT_DIR="${1:-.}"
BENCH_PATTERN="${2:-.}"
PROFILE_TYPE="${3:-none}"

# Input validation: sanitize BENCH_PATTERN to only allow safe benchmark patterns
# Valid Go benchmark patterns contain only alphanumeric chars, underscores, dots, slashes, carets, dollars, and stars
if [[ ! "$BENCH_PATTERN" =~ ^[a-zA-Z0-9_./*^$]+$ ]]; then
    echo "Error: Invalid benchmark pattern. Only alphanumeric characters, underscores, dots, slashes, carets, dollars, and stars are allowed." >&2
    exit 1
fi

# Input validation: PROJECT_DIR must be an existing directory
if [ ! -d "$PROJECT_DIR" ]; then
    echo "Error: Project directory '$PROJECT_DIR' does not exist." >&2
    exit 1
fi

cd "$PROJECT_DIR"

echo "Running Go benchmarks in: $(pwd)" >&2
echo "Pattern: $BENCH_PATTERN" >&2
echo "Profile: $PROFILE_TYPE" >&2

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_DIR="./bench_results"
mkdir -p "$OUTPUT_DIR"

cleanup() {
    echo "Benchmark completed" >&2
}
trap cleanup EXIT

show_usage() {
    cat << 'EOF'
Usage: benchmark.sh [project-dir] [pattern] [profile-type]

Arguments:
  project-dir   Project directory (default: current directory)
  pattern       Benchmark pattern to match (default: .)
  profile-type  Profiling type: none, cpu, mem, all (default: none)

Examples:
  benchmark.sh
  benchmark.sh . BenchmarkProcess
  benchmark.sh . . cpu
  benchmark.sh ./myproject BenchmarkHash all

Output:
  ./bench_results/benchmark_TIMESTAMP.txt     # Benchmark results
  ./bench_results/cpu_TIMESTAMP.prof          # CPU profile (if requested)
  ./bench_results/mem_TIMESTAMP.prof          # Memory profile (if requested)

Profile Analysis:
  go tool pprof ./bench_results/cpu_*.prof
  go tool pprof -alloc_space ./bench_results/mem_*.prof
EOF
}

# Build benchmark command using arrays to prevent flag injection
BENCH_CMD=(go test "-bench=$BENCH_PATTERN" -benchmem "-run=^$")

case "$PROFILE_TYPE" in
    cpu)
        BENCH_CMD+=("-cpuprofile=$OUTPUT_DIR/cpu_$TIMESTAMP.prof")
        ;;
    mem)
        BENCH_CMD+=("-memprofile=$OUTPUT_DIR/mem_$TIMESTAMP.prof")
        ;;
    all)
        BENCH_CMD+=("-cpuprofile=$OUTPUT_DIR/cpu_$TIMESTAMP.prof" "-memprofile=$OUTPUT_DIR/mem_$TIMESTAMP.prof")
        ;;
    none)
        ;;
    *)
        echo "Unknown profile type: $PROFILE_TYPE" >&2
        show_usage
        exit 1
        ;;
esac

BENCH_CMD+=("./...")

echo "Executing: ${BENCH_CMD[*]}" >&2
RESULT_FILE="$OUTPUT_DIR/benchmark_$TIMESTAMP.txt"

if "${BENCH_CMD[@]}" 2>&1 | tee "$RESULT_FILE"; then
    # Extract benchmark summary
    BENCH_COUNT=$(grep -c "^Benchmark" "$RESULT_FILE" 2>/dev/null || echo "0")

    # Build output JSON
    OUTPUT='{"success": true, "benchmarks": '"$BENCH_COUNT"', "results_file": "'"$RESULT_FILE"'"'

    if [ "$PROFILE_TYPE" = "cpu" ] || [ "$PROFILE_TYPE" = "all" ]; then
        OUTPUT="$OUTPUT"', "cpu_profile": "'"$OUTPUT_DIR/cpu_$TIMESTAMP.prof"'"'
    fi

    if [ "$PROFILE_TYPE" = "mem" ] || [ "$PROFILE_TYPE" = "all" ]; then
        OUTPUT="$OUTPUT"', "mem_profile": "'"$OUTPUT_DIR/mem_$TIMESTAMP.prof"'"'
    fi

    OUTPUT="$OUTPUT"'}'
    echo "$OUTPUT"
else
    echo '{"success": false, "error": "Benchmark execution failed"}'
    exit 1
fi
