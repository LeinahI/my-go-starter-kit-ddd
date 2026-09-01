---
name: golang-master
description: Golang language expert specializing in concurrency, performance optimization, standard library, and testing. Use when writing Go code, debugging concurrency issues, or optimizing performance.
---

# Golang Language Expert

Expert assistant for Go language mastery including concurrency patterns, performance optimization, standard library usage, and testing strategies.

## Thinking Process

When activated, follow this structured thinking approach to solve Go-related problems:

### Step 1: Problem Classification

**Goal:** Understand what type of Go challenge this is.

**Key Questions to Ask:**
- Is this a concurrency problem? (goroutines, channels, race conditions)
- Is this a performance problem? (memory, CPU, allocations)
- Is this a design problem? (interfaces, struct composition, packages)
- Is this a debugging problem? (error tracing, unexpected behavior)
- Is this a testing problem? (unit tests, benchmarks, mocks)

**Decision Point:** Classify the problem to select appropriate patterns:
- Concurrency → Check for race conditions, proper synchronization
- Performance → Profile first, then optimize
- Design → Apply interface segregation, composition over inheritance
- Debugging → Trace error chain, check nil handling
- Testing → Table-driven tests, dependency injection

### Step 2: Context Analysis

**Goal:** Understand the existing codebase context and constraints.

**Key Questions to Ask:**
- What Go version is being used? (generics require 1.18+)
- What is the module structure? (standard layout vs flat)
- What dependencies exist? (standard library vs third-party)
- Are there existing patterns to follow? (error handling style, logging approach)

**Actions:**
1. Check `go.mod` for Go version and dependencies
2. Scan existing code for established patterns
3. Identify any project-specific conventions

### Step 3: Solution Design (Idiomatic Go)

**Goal:** Design a solution that follows Go best practices.

**Thinking Framework - Apply These Principles:**

1. **Simplicity First**
   - "Is there a simpler way to achieve this?"
   - Avoid over-engineering; Go favors explicitness

2. **Error Handling**
   - "What can fail here? How should failures propagate?"
   - Use `fmt.Errorf("context: %w", err)` for wrapping
   - Define custom error types for domain-specific errors

3. **Concurrency Safety**
   - "Is this data accessed from multiple goroutines?"
   - "Do I need a mutex, channel, or atomic operation?"
   - Prefer channels for coordination, mutexes for state protection

4. **Interface Design**
   - "What behavior is needed, not what data?"
   - Keep interfaces small (1-3 methods)
   - Accept interfaces, return concrete types

5. **Resource Management**
   - "What needs to be closed/cleaned up?"
   - Use defer for cleanup
   - Propagate context for cancellation

### Step 4: Implementation with Patterns

**Goal:** Apply well-known Go patterns appropriately.

**Pattern Selection Guide:**

| Problem Type | Recommended Pattern |
|-------------|---------------------|
| Process N items concurrently | Worker Pool |
| Limit concurrent operations | Semaphore |
| Timeout long operations | Context with Deadline |
| Coordinate multiple goroutines | WaitGroup |
| Fan-out/Fan-in processing | Channel pipelines |
| Graceful shutdown | Context cancellation + signal handling |

**Decision Point:** For each pattern, ask:
- "Why is this pattern appropriate for this problem?"
- "What are the edge cases I need to handle?"

### Step 5: Testing Strategy

**Goal:** Ensure code correctness and maintainability.

**Thinking Framework:**
- "What are the happy paths to test?"
- "What edge cases could break this?"
- "Is this code testable? If not, how should I refactor?"

**Testing Checklist:**
1. Table-driven tests for functions with multiple input scenarios
2. Race detection for concurrent code (`go test -race`)
3. Benchmarks for performance-critical code
4. Mock interfaces for external dependencies

### Step 6: Review and Refine

**Goal:** Ensure the solution is production-ready.

**Final Checks:**
- [ ] All errors are handled or explicitly ignored with comment
- [ ] Context is propagated through the call chain
- [ ] Resources are properly closed (defer, Close())
- [ ] No goroutine leaks (channels closed, contexts cancelled)
- [ ] Code is testable (dependencies injected)
- [ ] Documentation for exported symbols

## Usage

### Run Linters

```bash
bash .cursor/skills/golang-master/scripts/lint.sh [project-dir] [fix-mode]
```

**Arguments:**
- `project-dir` - Project directory (default: current directory)
- `fix-mode` - Set to `true` to auto-fix issues (default: false)

**Examples:**
```bash
bash .cursor/skills/golang-master/scripts/lint.sh
bash .cursor/skills/golang-master/scripts/lint.sh ./my-project true
```

**Runs:** go vet, gofmt, staticcheck, golangci-lint, go mod tidy

### Run Benchmarks

```bash
bash .cursor/skills/golang-master/scripts/benchmark.sh [project-dir] [pattern] [profile-type]
```

**Arguments:**
- `project-dir` - Project directory (default: current directory)
- `pattern` - Benchmark pattern to match (default: .)
- `profile-type` - Profiling: none, cpu, mem, all (default: none)

**Examples:**
```bash
bash .cursor/skills/golang-master/scripts/benchmark.sh
bash .cursor/skills/golang-master/scripts/benchmark.sh . BenchmarkProcess cpu
bash .cursor/skills/golang-master/scripts/benchmark.sh ./myproject . all
```

## Documentation Resources

**Official Documentation:**
- Go Docs: `https://go.dev/doc/`
- Effective Go: `https://go.dev/doc/effective_go`
- Go by Example: `https://gobyexample.com/`
- uber-go/guide: `https://github.com/uber-go/guide`

## Concurrency Patterns

### Worker Pool

```go
func workerPool(jobs <-chan Job, numWorkers int) <-chan Result {
    results := make(chan Result, len(jobs))
    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}
```

### Context with Timeout

```go
func fetchWithTimeout(url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

### Semaphore Pattern

```go
type Semaphore struct {
    sem chan struct{}
}

func NewSemaphore(n int) *Semaphore {
    return &Semaphore{sem: make(chan struct{}, n)}
}

func (s *Semaphore) Acquire() { s.sem <- struct{}{} }
func (s *Semaphore) Release() { <-s.sem }
```

## Error Handling

### Wrapping Errors

```go
func processUser(id int) error {
    user, err := fetchUser(id)
    if err != nil {
        return fmt.Errorf("failed to fetch user %d: %w", id, err)
    }

    if err := validateUser(user); err != nil {
        return fmt.Errorf("validation failed for user %d: %w", id, err)
    }

    return nil
}
```

### Custom Error Types

```go
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// Usage
if errors.Is(err, &NotFoundError{}) {
    // Handle not found
}
```

### Error Checking

```go
// Use errors.Is for sentinel errors
if errors.Is(err, sql.ErrNoRows) {
    return nil, &NotFoundError{Resource: "user", ID: id}
}

// Use errors.As for error types
var netErr *net.OpError
if errors.As(err, &netErr) {
    // Handle network error
}
```

## Performance Profiling

### CPU Profiling

```bash
# Run benchmark with CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Interactive commands
(pprof) top10
(pprof) web
(pprof) list FunctionName
```

### Memory Profiling

```bash
# Run benchmark with memory profile
go test -memprofile=mem.prof -bench=.
go tool pprof -alloc_space mem.prof

# Check for allocations
go build -gcflags='-m' ./...
```

### Benchmark Example

```go
func BenchmarkProcess(b *testing.B) {
    data := setupTestData()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        process(data)
    }
}

func BenchmarkProcessParallel(b *testing.B) {
    data := setupTestData()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            process(data)
        }
    })
}
```

## Testing Patterns

### Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

## Present Results to User

When providing Go solutions:
- Use idiomatic Go patterns
- Always handle errors explicitly
- Include context propagation
- Provide test examples
- Note Go version requirements (generics: 1.18+)

## Troubleshooting

**"Data race detected"**
- Use `go test -race` to find races
- Protect shared state with mutex
- Consider channels for communication

**"Goroutine leak"**
- Ensure channels are closed
- Use context for cancellation
- Check for blocking operations

**"High memory usage"**
- Profile with pprof
- Check for slice capacity issues
- Look for map accumulation
