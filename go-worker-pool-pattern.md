# Go Worker Pool Pattern — A Detailed, Beginner-Friendly Guide

> Goal: understand what a worker pool is, how goroutines and channels cooperate, why channels are closed, how `WaitGroup` works, how results are returned, and how deadlocks happen.

This guide builds the pattern step by step using simple Go examples and a real-world pizza-shop analogy.

---

## 1. What problem does a worker pool solve?

Imagine a pizza shop receives 100 orders at once.

- The **orders** are jobs.
- The **chefs** are workers.
- The **order counter** is the jobs channel.
- The **completed-pizza counter** is the results channel.
- The cashier is the `main` goroutine.

You could create a new goroutine for every order. But for many workloads, it is useful to limit how many jobs are being processed at once. A worker pool creates a fixed number of workers and lets them share the work.

For example, with 100 jobs and 3 workers:

- There are 100 jobs total.
- There are only 3 workers.
- Each job is handled by one worker.
- A worker can process multiple jobs over time.
- A job is **not** automatically sent to every worker.

A worker pool is useful for tasks such as resizing images, processing uploaded files, sending emails, generating reports, calling external APIs with a concurrency limit, and processing messages from a queue.

A worker pool is not automatically faster for every task. It is most useful when you want bounded concurrency or when many independent jobs can make progress concurrently.

---

## 2. The basic vocabulary

### Job

A **job** is one unit of work.

In our examples, a job is an integer. The work is to calculate its square:

```go
job := 5
result := job * job // 25
```

In a real application, a job might be an image filename to resize, an email to send, a user ID whose report should be generated, or an order ID to process.

### Worker

A **worker** is a goroutine that repeatedly receives and processes jobs.

A worker is not a special Go type. It is just a function running in a goroutine, often with a loop.

### Jobs channel

The `jobs` channel carries work from the producer to workers:

```go
jobs := make(chan int)
```

### Results channel

The `results` channel carries completed work back to the consumer:

```go
results := make(chan int)
```

### WaitGroup

A `sync.WaitGroup` lets one goroutine wait until a group of goroutines has finished:

```go
var wg sync.WaitGroup
wg.Add(3) // Three workers are expected to call Done.
```

A WaitGroup does not send values and does not protect shared maps or other state. It tracks completion.

---

## 3. Goroutines: concurrency versus parallelism

Starting a worker with:

```go
go worker(1, jobs, results, &wg)
```

starts the function in a separate goroutine. The caller does not wait for the worker function to return.

With three workers:

```go
go worker(1, jobs, results, &wg)
go worker(2, jobs, results, &wg)
go worker(3, jobs, results, &wg)
```

the program has three worker goroutines plus the `main` goroutine.

- **Concurrency** means multiple tasks are in progress and can make progress independently.
- **Parallelism** means multiple tasks are literally executing at the same instant, usually on different CPU cores.

Go schedules goroutines. Their execution order is not something the program should rely on. Multiple goroutines may execute in parallel when resources allow.

When a goroutine blocks on a channel receive, it does not continuously burn CPU waiting. It becomes blocked until the operation can proceed.

When `main()` returns, the Go process ends. Other goroutines do not keep the process alive automatically. Use synchronization when work must finish before the program exits.

---

## 4. Channels: sending and receiving

The two most important operations are:

```go
ch <- value // Send
value := <-ch // Receive
```

A channel is a communication mechanism between goroutines.

### Unbuffered channel

```go
jobs := make(chan int)
```

This channel has capacity zero. It does not hold a queue of values. A send must synchronize with a receiver:

```go
jobs <- 100
```

The send cannot complete until a receiver accepts the value. Likewise, a receive blocks until a sender provides a value. Think of it as a direct handoff.

### Buffered channel

```go
results := make(chan int, 5)
```

This channel can temporarily hold up to five values.

- If there is room, a send can complete without a receiver waiting at that exact moment.
- If the buffer is full, a send blocks until a receiver makes room.
- If the buffer is empty, a receive blocks until a value arrives.

A buffer is not infinite storage and does not eliminate the need to coordinate goroutines. It only changes when sends and receives block.

### Important distinction

A successful send on an unbuffered channel means the value has been handed to a receiver. It does **not** guarantee that the receiver has finished all the work after receiving it.

A buffered send can finish when the value is stored, even though nobody has received it yet.

---

## 5. Step 1 — One worker processes one job

Start with the smallest example:

```go
package main

import "fmt"

func worker(id int, jobs chan int) {
    job := <-jobs
    fmt.Println("worker", id, "got job", job)
}

func main() {
    jobs := make(chan int)
    go worker(1, jobs)
    jobs <- 100
}
```

The worker waits at `job := <-jobs`. When `main` sends `100`, the receiver gets it and can print it.

However, this is not a robust example of waiting for the worker to finish. The send/receive handoff does not guarantee that the worker has completed its `fmt.Println` before `main` returns. The process may terminate before the print occurs.

A temporary `time.Sleep` can make an experiment appear to work, but sleep is not proper synchronization. A WaitGroup is a better way to wait for completion.

---

## 6. Step 2 — Keep one worker alive for multiple jobs

A worker can loop:

```go
func worker(id int, jobs chan int) {
    for {
        job := <-jobs
        fmt.Println("worker", id, "got job", job)
    }
}
```

The `for {}` loop repeats forever:

1. Wait for a job.
2. Receive the job.
3. Process it.
4. Return to the top of the loop.

The goroutine remains alive because the function has not returned. When no job is available, it blocks on the receive.

But this loop has no normal shutdown condition. A common alternative is `for job := range jobs`, explained later.

---

## 7. Step 3 — Multiple workers share one jobs channel

A worker pool starts several worker goroutines, all receiving from the same jobs channel:

```go
go worker(1, jobs)
go worker(2, jobs)
go worker(3, jobs)
```

Imagine sending:

```go
jobs <- 100
jobs <- 200
jobs <- 300
jobs <- 400
jobs <- 500
```

A possible assignment is:

| Job | Example worker |
|---:|---|
| 100 | Worker 2 |
| 200 | Worker 1 |
| 300 | Worker 3 |
| 400 | Worker 2 |
| 500 | Worker 1 |

This is only an example. The actual assignment and order are not guaranteed.

**Each job is received by one worker, not broadcast to every worker.** One worker can handle several jobs over time. The workers share the work.

The channel does not create the workers. The `go worker(...)` statements create them; the channel is their shared communication point.

---

## 8. Step 4 — Why close the jobs channel?

A worker loop such as:

```go
for {
    job := <-jobs
    // process job
}
```

does not have a built-in way to know that the producer has permanently finished sending jobs. It may wait forever for another job.

Use this pattern instead:

```go
func worker(id int, jobs chan int) {
    for job := range jobs {
        fmt.Println("worker", id, "got job", job)
    }

    fmt.Println("worker", id, "finished")
}
```

The `range` loop keeps receiving values until the channel is closed and drained.

The producer closes the jobs channel after sending all jobs:

```go
jobs <- 100
jobs <- 200
jobs <- 300

close(jobs)
```

Closing means: **no more values will be sent on this channel.**

It does not free memory immediately, discard buffered values, or kill receivers immediately. Already-buffered values remain available to receive. Once the channel is closed and empty, `range` ends.

### Who should close a channel?

Usually, the sender/producer closes the channel, because it knows when it has finished sending.

Never send to a closed channel; that causes a panic. Avoid closing a channel twice. A receiver generally should not close a channel that other goroutines may still be sending to.

If you do not close `jobs` and workers use `for job := range jobs`, they may stay blocked waiting for future jobs after all current jobs have been processed. If `main` returns, the entire process terminates anyway, potentially before workers finish.

---

## 9. Step 5 — Wait for workers with `sync.WaitGroup`

A WaitGroup is useful when `main` must not exit until the workers have finished.

```go
var wg sync.WaitGroup
wg.Add(3)

go worker(1, jobs, &wg)
go worker(2, jobs, &wg)
go worker(3, jobs, &wg)
```

Each worker signals completion:

```go
func worker(id int, jobs chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Println("worker", id, "got job", job)
    }
}
```

Then the producer closes jobs and waits:

```go
close(jobs)
wg.Wait()

fmt.Println("all workers finished")
```

What each method means:

- `wg.Add(3)`: the counter increases by three.
- `wg.Done()`: decreases the counter by one.
- `wg.Wait()`: blocks until the counter reaches zero.

`defer wg.Done()` runs when the worker function returns, including when it returns because the jobs loop ends.

### Correct ordering

For a finite batch of jobs:

1. Start workers.
2. Submit all jobs.
3. Close `jobs`.
4. Wait for workers.

Do not call `wg.Wait()` before closing `jobs` if the workers are waiting for the channel to close to exit their `range` loops. Main could wait forever for workers that are themselves waiting for main to close the channel.

A WaitGroup tracks completion; it does not close channels or deliver jobs.

---

## 10. Step 6 — Have workers return results

A worker pool often needs to return values, not just print that a job was processed.

For our example, each worker squares a number:

```go
func worker(
    id int,
    jobs chan int,
    results chan int,
    wg *sync.WaitGroup,
) {
    defer wg.Done()

    for job := range jobs {
        fmt.Println("worker", id, "got job", job)

        result := job * job
        results <- result
    }
}
```

There are now two flows:

```text
jobs:     main ──────────> workers
results:  main <────────── workers
```

For job `5`, the worker computes `25` and sends it:

```go
results <- result
```

Main receives it with:

```go
result := <-results
```

The `results` channel is how results are communicated back to main.

### Unbuffered results channel

```go
results := make(chan int)
```

A worker sending a result blocks until a receiver is ready. That can be perfectly fine when main receives results while workers run.

It can deadlock if main waits for workers to finish before receiving results and workers are blocked trying to send those results.

### Buffered results channel

```go
results := make(chan int, 5)
```

Workers can send up to five results without a receiver immediately waiting. Once the buffer is full, further sends block until main receives a value.

The number `5` here is a capacity, not a special relationship to the number of workers. If you have 3 workers and 100 jobs, you can have 100 results. A buffer of 3 does not imply there will be only 3 results.

---

## 11. A common deadlock — waiting before receiving results

Consider this worker:

```go
func worker(
    id int,
    jobs chan int,
    results chan int,
    wg *sync.WaitGroup,
) {
    defer wg.Done()

    for job := range jobs {
        result := job * job
        results <- result
    }
}
```

And this main logic:

```go
jobs := make(chan int)
results := make(chan int)

var wg sync.WaitGroup
wg.Add(3)

go worker(1, jobs, results, &wg)
go worker(2, jobs, results, &wg)
go worker(3, jobs, results, &wg)

// Send jobs...
close(jobs)

wg.Wait() // Problem!
for result := range results {
    fmt.Println(result)
}
```

This can deadlock because `results` is unbuffered.

The chain of events is:

1. Main sends the jobs and closes `jobs`.
2. Workers process jobs and attempt `results <- result`.
3. Since main has not started receiving, the sends block.
4. Workers cannot return, so their deferred `wg.Done()` calls have not happened.
5. Main is stuck in `wg.Wait()`, waiting for workers to finish.
6. Main cannot reach the results loop to receive values.

In short:

```text
main waits for workers
workers wait for main to receive results
main will not receive until workers finish
```

Everyone is waiting for someone else. This is a deadlock.

---

## 12. Why a buffer can appear to fix it — and why that is limited

With five jobs:

```go
results := make(chan int, 5)
```

and:

```go
wg.Wait()

for result := range results {
    fmt.Println(result)
}
```

this might work if all five results fit into the buffer. The workers can send all five results, return, and decrement the WaitGroup. Main then begins receiving.

But suppose you have 100 jobs and a results buffer of five. Once the buffer is full, another worker trying to send a result blocks. If main is still waiting in `wg.Wait()`, the deadlock returns.

Therefore, **a sufficiently large buffer can hide this ordering problem for a fixed workload, but it is not the most robust general design**.

---

## 13. The complete worker-pool pattern with results

This is the pattern we built. It uses:
- three worker goroutines,
- an unbuffered jobs channel,
- a buffered results channel,
- a WaitGroup,
- an extra coordinator goroutine that waits for workers and closes results.

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(
    id int,
    jobs chan int,
    results chan int,
    wg *sync.WaitGroup,
) {
    defer wg.Done()

    for job := range jobs {
        fmt.Println("worker", id, "got job", job)

        result := job * job
        time.Sleep(time.Second) // Simulate work.

        results <- result
    }
}

func main() {
    jobs := make(chan int)
    results := make(chan int, 5)

    var wg sync.WaitGroup
    wg.Add(3)

    go worker(1, jobs, results, &wg)
    go worker(2, jobs, results, &wg)
    go worker(3, jobs, results, &wg)

    // Producer sends all jobs.
    jobs <- 100
    jobs <- 200
    jobs <- 300
    jobs <- 400
    jobs <- 500

    // No more jobs will be sent.
    close(jobs)

    // Coordinator waits for every worker and then closes results.
    go func() {
        wg.Wait()
        close(results)
    }()

    // Main receives results while workers are processing.
    fmt.Println("Results from jobs:")

    for result := range results {
        fmt.Println(result)
    }

    fmt.Println("All workers finished")
}
```

The results can arrive in any order. For example, the squares are `10000`, `40000`, `90000`, `160000`, and `250000`, but the order in which they are printed is not guaranteed.

### Why is the extra goroutine necessary in this design?

Main has a job to do: receive results. It should not stop receiving while waiting for the workers, because workers may need main to receive results before they can finish.

This goroutine runs independently:

```go
go func() {
    wg.Wait()
    close(results)
}()
```

It waits for all workers while main continues receiving results. Once the workers have finished, it closes `results`, allowing main's `range` loop to end after all results have been consumed.

This is a **coordinator** or **closer goroutine**.

### Why not `defer close(results)` in main?

If main does this:

```go
defer close(results)

for result := range results {
    fmt.Println(result)
}
```

the deferred call happens only when `main` is returning. But `main` cannot return while the `range` loop is waiting for the channel to close. That creates a circular dependency.

Also, closing results before all workers have finished sending can cause a panic. The closer goroutine avoids that by waiting first.

### Why does `range results` end?

It ends when:
1. `results` is closed, and
2. all buffered results have been consumed.

Closing the channel tells the receiver that no more values will arrive. It does not discard values already buffered.

---

## 14. The complete lifecycle in one diagram

```text
main creates jobs and results channels
                |
                v
       Starts 3 workers
                |
                v
       Sends five jobs
                |
                v
          close(jobs)
                |
                v
      Workers finish jobs
       and send results
                |
       +--------+---------+
       |                  |
       v                  v
Main receives         Coordinator waits
and prints results    with wg.Wait()
                          |
                          v
                   All workers finish
                          |
                          v
                   close(results)
                          |
                          v
               Main's range loop exits
                          |
                          v
                   Program finishes
```

---

## 15. Real-world example: image-processing service

Imagine an application where users upload photos and the server needs to create thumbnails.

- **Job:** the path or ID of an uploaded image.
- **Worker:** a goroutine that reads the image and creates a thumbnail.
- **Jobs channel:** distributes image tasks.
- **Results channel:** returns each thumbnail path or processing status.
- **WaitGroup:** tracks whether all workers for the batch have finished.
- **Close `jobs`:** signals that the current batch contains no more images.
- **Close `results`:** signals that no more thumbnail results will be produced.

If 50 images are uploaded and you have 4 workers, each image is processed by one worker. Each worker can handle several images. You do not need 50 goroutines just to process those 50 images concurrently; a pool of four workers limits the number of jobs being processed at once.

The same pattern applies to sending emails, generating reports, resizing images, or making independent API calls.

---

## 16. Common mistakes and the rule to remember

| Mistake | Why it is a problem |
|---|---|
| Expecting every worker to process every job | Workers share the channel; each job is received by one worker. |
| Forgetting to close `jobs` for a `range` worker | Workers can wait forever for more jobs. |
| Calling `wg.Wait()` before closing `jobs` | Workers may never exit their receive loops. |
| Calling `wg.Wait()` before receiving unbuffered results | Workers may block sending results, while main waits for them to finish. |
| Assuming a buffer is unlimited | Sends block when the buffer fills. |
| Closing `results` before workers finish | A worker may send to a closed channel and panic. |
| Using `defer close(results)` in main before a `range` loop | The loop may wait forever before the deferred close can run. |
| Closing a channel to free memory | Closing signals that no more sends will occur; it is not a memory-freeing operation. |
| Using `time.Sleep` to coordinate goroutines | Sleep does not guarantee that other goroutines have completed. Use synchronization. |

### Quick mental model

- `go worker(...)` — start a worker.
- `jobs <- job` — submit work.
- `<-jobs` / `for job := range jobs` — receive work.
- `close(jobs)` — no more jobs are coming.
- `results <- result` — publish completed work.
- `<-results` / `for result := range results` — collect results.
- `wg.Done()` — signal that a worker has returned.
- `wg.Wait()` — wait for all tracked workers to finish.
- `close(results)` — no more results will be sent; receivers can finish once the channel is drained.

**The core lesson:** a worker pool is not just about starting multiple goroutines. It is about coordinating the lifecycle of work: distributing jobs, processing them concurrently, collecting results, and signalling when everything is finished.
