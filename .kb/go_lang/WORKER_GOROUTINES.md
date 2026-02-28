# Worker Goroutines in Go

> Quick reference on how worker goroutines operate and how to structure a worker pool.

---

## Concept
- A **goroutine** is a lightweight thread managed by the Go runtime.
- A **worker goroutine** loops over a queue (channel) of jobs and processes them one by one.
- Workers typically share common resources (channels, wait groups) and exit automatically when the queue is closed.

---

## Anatomy of a Worker Pool
```go
jobs := make(chan Job)
results := make(chan Result)
var wg sync.WaitGroup

for i := 0; i < workerCount; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        for job := range jobs {
            results <- process(id, job)
        }
    }(i)
}

// enqueue work
for _, job := range allJobs {
    jobs <- job
}
close(jobs)

wg.Wait()
close(results)
```

### Key ideas
1. **Channel ownership**: the goroutine that closes the `jobs` channel owns it.
2. **Backpressure**: buffered channels control how many jobs can be queued without blocking.
3. **Fairness**: Go runtime schedules goroutines cooperatively—long blocking operations stall the worker unless moved into another goroutine.
4. **Cancellation**: use `context.Context` or a `done` channel to stop workers early.

---

## Debugging Tips
- Use `runtime.NumGoroutine()` to ensure goroutines exit.
- Add structured logs (worker id, job id, duration) to spot stragglers.
- Run tests with `-race` to catch data races between workers.

---

## Common Pitfalls
| Pitfall | Fix |
| --- | --- |
| Blocking send on `results` when no reader | Launch a collector goroutine or buffer the channel. |
| Forgetting to close `jobs` | Workers never exit; WaitGroup blocks forever. |
| Shared state without locks | Use mutexes or per-worker state to prevent races. |
| Spawning unbounded goroutines | Keep worker count fixed; reuse goroutines through the loop. |

---

## When to Use
- Parallelizing independent tasks (e.g., HTTP fetches, test executions).
- Rate-limiting access to scarce resources (database connections, browsers).
- Fan-out/fan-in pipelines (split work, then merge results).

Worker goroutines are the backbone of kexas's parallel test runner: each worker owns a browser instance, pulls tests from a channel, and reports results back through a synchronized collector.
