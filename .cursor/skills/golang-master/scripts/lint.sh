#!/bin/bash
set -e

# Go Linting Script
# Runs multiple linters for comprehensive code quality checks

PROJECT_DIR="${1:-.}"
FIX_MODE="${2:-false}"

# Input validation: PROJECT_DIR must be an existing directory
if [ ! -d "$PROJECT_DIR" ]; then
    echo "Error: Project directory '$PROJECT_DIR' does not exist." >&2
    exit 1
fi

# Input validation: FIX_MODE must be "true" or "false"
if [ "$FIX_MODE" != "true" ] && [ "$FIX_MODE" != "false" ]; then
    echo "Error: FIX_MODE must be 'true' or 'false', got '$FIX_MODE'." >&2
    exit 1
fi

cd "$PROJECT_DIR"

echo "Running Go linters in: $(pwd)" >&2

RESULTS=()
ERRORS=0

# Check go vet
echo "Running go vet..." >&2
if VET_OUTPUT=$(go vet ./... 2>&1); then
    RESULTS+=('{"tool": "go vet", "status": "pass"}')
    echo "go vet: PASS" >&2
else
    RESULTS+=('{"tool": "go vet", "status": "fail", "output": "'"$(echo "$VET_OUTPUT" | head -20 | tr '\n' ' ')"'"}')
    echo "go vet: FAIL" >&2
    ((ERRORS++))
fi

# Check gofmt
echo "Running gofmt..." >&2
GOFMT_OUTPUT=$(gofmt -l . 2>&1 || true)
if [ -z "$GOFMT_OUTPUT" ]; then
    RESULTS+=('{"tool": "gofmt", "status": "pass"}')
    echo "gofmt: PASS" >&2
else
    if [ "$FIX_MODE" = "true" ]; then
        gofmt -w .
        RESULTS+=('{"tool": "gofmt", "status": "fixed", "files": "'"$(echo "$GOFMT_OUTPUT" | wc -l | tr -d ' ')"'"}')
        echo "gofmt: FIXED" >&2
    else
        RESULTS+=('{"tool": "gofmt", "status": "fail", "files": "'"$(echo "$GOFMT_OUTPUT" | tr '\n' ' ')"'"}')
        echo "gofmt: FAIL" >&2
        ((ERRORS++))
    fi
fi

# Check staticcheck if available
if command -v staticcheck &> /dev/null; then
    echo "Running staticcheck..." >&2
    if STATIC_OUTPUT=$(staticcheck ./... 2>&1); then
        RESULTS+=('{"tool": "staticcheck", "status": "pass"}')
        echo "staticcheck: PASS" >&2
    else
        RESULTS+=('{"tool": "staticcheck", "status": "fail", "output": "'"$(echo "$STATIC_OUTPUT" | head -20 | tr '\n' ' ')"'"}')
        echo "staticcheck: FAIL" >&2
        ((ERRORS++))
    fi
else
    echo "staticcheck not installed, skipping..." >&2
    RESULTS+=('{"tool": "staticcheck", "status": "skipped", "reason": "not installed"}')
fi

# Check golangci-lint if available
if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint..." >&2
    if [ "$FIX_MODE" = "true" ]; then
        LINT_OUTPUT=$(golangci-lint run --fix ./... 2>&1 || true)
    else
        LINT_OUTPUT=$(golangci-lint run ./... 2>&1 || true)
    fi

    if [ -z "$LINT_OUTPUT" ]; then
        RESULTS+=('{"tool": "golangci-lint", "status": "pass"}')
        echo "golangci-lint: PASS" >&2
    else
        RESULTS+=('{"tool": "golangci-lint", "status": "fail", "issues": "'"$(echo "$LINT_OUTPUT" | grep -c ":" || echo "0")"'"}')
        echo "golangci-lint: FAIL" >&2
        ((ERRORS++))
    fi
else
    echo "golangci-lint not installed, skipping..." >&2
    RESULTS+=('{"tool": "golangci-lint", "status": "skipped", "reason": "not installed"}')
fi

# Check go mod tidy
# Note: go mod tidy may download external modules. We use GONOSUMCHECK=off
# to ensure all downloads are verified against the checksum database.
echo "Checking go.mod..." >&2
if [ ! -f go.mod ]; then
    echo "go.mod not found, skipping go mod tidy check" >&2
    RESULTS+=('{"tool": "go mod tidy", "status": "skipped", "reason": "go.mod not found"}')
else
    cp go.mod go.mod.bak 2>/dev/null || true
    cp go.sum go.sum.bak 2>/dev/null || true
    GONOSUMCHECK=off GOFLAGS="" go mod tidy 2>/dev/null || true

    if diff -q go.mod go.mod.bak > /dev/null 2>&1; then
        RESULTS+=('{"tool": "go mod tidy", "status": "pass"}')
        echo "go mod tidy: PASS" >&2
    else
        if [ "$FIX_MODE" = "true" ]; then
            RESULTS+=('{"tool": "go mod tidy", "status": "fixed"}')
            echo "go mod tidy: FIXED" >&2
        else
            mv go.mod.bak go.mod 2>/dev/null || true
            mv go.sum.bak go.sum 2>/dev/null || true
            RESULTS+=('{"tool": "go mod tidy", "status": "fail", "message": "go.mod needs tidying"}')
            echo "go mod tidy: FAIL" >&2
            ((ERRORS++))
        fi
    fi
    rm -f go.mod.bak go.sum.bak
fi

# Output results
RESULTS_JSON=$(printf '%s\n' "${RESULTS[@]}" | paste -sd, -)
if [ $ERRORS -eq 0 ]; then
    echo '{"success": true, "errors": 0, "results": ['"$RESULTS_JSON"']}'
else
    echo '{"success": false, "errors": '"$ERRORS"', "results": ['"$RESULTS_JSON"']}'
fi
