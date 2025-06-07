# Greenlight API Profiling Guide

This document explains how to use the profiling features in the Greenlight API for performance analysis and optimization.

## What You Need

- `curl` - For capturing profiles (should be installed by default with your distro)
- `jq` - For JSON formatting (optional)
- `go` - For profile analysis(it comes with pprof, in net/http/pprof package)


## Quick Start

1. **Start the API with profiling enabled:**
   ```bash
   make run/api/profiling
   ```

2. **Capture a CPU profile:**
   ```bash
   make profile/cpu
   ```

3. **View current metrics:**
   ```bash
   make profile/metrics
   ```

## Configuration

Profiling is controlled by command-line flags:

- `--profiling-enabled`: Enable/disable profiling server (default: false)
- `--profiling-port`: Port for profiling server (default: 5000)

Example:
```bash
go run ./cmd/api -db-dsn=$GREENLIGHT_DB_DSN -profiling-enabled=true -profiling-port=6000
```

## Available Profiles

### CPU Profile
Captures CPU usage over time to identify performance bottlenecks.

```bash
# Via Makefile
make profile/cpu

# Direct script usage
./scripts/profile.sh cpu 60  # 60-second duration

# Manual curl
curl "http://localhost:5000/debug/pprof/profile?seconds=30" -o cpu.prof
go tool pprof cpu.prof
```

### Memory Profiles

**Heap Profile** - Current memory allocations:
```bash
make profile/heap
./scripts/profile.sh heap
```

**Allocation Profile** - All memory allocations:
```bash
./scripts/profile.sh allocs
```

### Concurrency Profiles

**Goroutine Profile** - Current goroutines:
```bash
make profile/goroutines
./scripts/profile.sh goroutine
```

**Block Profile** - Blocking operations:
```bash
./scripts/profile.sh block
```

**Mutex Profile** - Mutex contention:
```bash
./scripts/profile.sh mutex
```

### Execution Trace
Detailed execution timeline:
```bash
make profile/trace
./scripts/profile.sh trace 10  # 10-second trace
```

### All Profiles
Capture everything at once:
```bash
make profile/all
./scripts/profile.sh all
```

## Endpoints

*Note: start with the first endpoint, as manually typing the endpoints will prompt a download popup, but should you not want this, click on the links provided to view in the browser instead*

When profiling is enabled, these endpoints are available on the profiling port:

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Interactive profile index |
| `/debug/pprof/profile` | CPU profile |
| `/debug/pprof/heap` | Heap memory profile |
| `/debug/pprof/goroutine` | Goroutine profile |
| `/debug/pprof/allocs` | Allocation profile |
| `/debug/pprof/block` | Block profile |
| `/debug/pprof/mutex` | Mutex profile |
| `/debug/pprof/trace` | Execution trace |
| `/debug/vars` | Runtime metrics (expvar) |
| `/health` | Profiling server health check |

## Analysis Tools

### Command Line Analysis
```bash
# Analyze a profile file
go tool pprof profile.prof

# Interactive web interface
# NOTE: you need a package called [graphviz](https://www.graphviz.org/)
go tool pprof -http=:8080 profile.prof

# Analyze trace file
go tool trace trace.prof
```

### Web Interface
```bash
# Direct web analysis
go tool pprof -http=:8080 http://localhost:5000/debug/pprof/profile?seconds=30
```

### Common pprof Commands
```bash
# Inside pprof interactive mode:
(pprof) top10          # Show top 10 functions
(pprof) list funcname  # Show function source
(pprof) web            # Generate web visualization
(pprof) svg            # Generate SVG graph
(pprof) help           # Show all commands

go tool pprof -help    # Show all commands without being in interactive mode
```


## Runtime Metrics

View real-time application metrics:
```bash
curl http://localhost:5000/debug/vars | jq .
```

Available metrics:
- `goroutines`: Current number of goroutines
- `database`: Database connection pool stats
- `total_requests_received`: Total HTTP requests
- `total_responses_sent`: Total HTTP responses
- `total_processing_time_µs`: Cumulative processing time
- `total_responses_sent_by_status`: Response counts by HTTP status
- `version`: Application version

## Performance Monitoring

### Slow Request Detection
The API automatically logs requests taking longer than 500ms with details:
- Method and path
- Duration in milliseconds
- HTTP status code
- Response size

### Memory Usage Monitoring
Monitor heap usage with periodic heap profiles:
```bash
watch -n 30 "curl -s http://localhost:5000/debug/pprof/heap -o heap_$(date +%s).prof"
```

### Goroutine Leak Detection
Monitor goroutine count over time:
```bash
watch -n 10 "curl -s http://localhost:5000/debug/vars | jq .goroutines"
```

## Best Practices

### When to Profile
- During load testing
- When investigating performance issues
- Before and after optimizations
- During production monitoring (with caution)

### Profile Duration
- **CPU profiles**: 30-60 seconds during load
- **Memory profiles**: Instantaneous snapshots
- **Traces**: 5-15 seconds (generates large files)

### Profile Storage
Profiles are saved to `./profiles/` directory with timestamps:
```
profiles/
├── cpu_20240101_143022.prof
├── heap_20240101_143045.prof
└── trace_20240101_143100.prof
```

## Troubleshooting

### Profiling Server Not Starting
1. Check if profiling is enabled: `--profiling-enabled=true`
2. Verify port is not in use: `lsof -i :5000   # not a must it be port 5000`
3. Check logs for startup errors

### Empty Profiles
1. Ensure application is under load during profiling
2. Check profile duration (too short may miss activity)
3. Verify endpoints are accessible

### High Memory Usage
1. Take heap profiles before and after operations
2. Look for memory leaks with `go tool pprof -alloc_space`
3. Monitor goroutine count for leaks

### Performance Regression
1. Compare profiles before and after changes
2. Use `go tool pprof -base old.prof new.prof` for diffs
3. Focus on hot paths in CPU profiles

## Example Workflow

```bash
# 1. Start API with profiling
make run/api/profiling

# 2. Generate some load (in another terminal)
for i in {1..1000}; do
  curl http://localhost:4000/v1/healthcheck
done

# 3. Capture profiles during load
make profile/all

# 4. Analyze results
go tool pprof -http=:8080 profiles/cpu_*.prof
```
