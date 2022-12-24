# Go Worker ![GitHub release](https://img.shields.io/github/v/tag/purposeinplay/go-worker)

[![lint-test-redisw](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_redisw.yml/badge.svg)](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_redisw.yml)
[![lint-test-asynq](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_asynq.yml/badge.svg)](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_asynq.yml)
[![lint-test-inmem](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_inmem.yml/badge.svg)](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_inmem.yml)
[![lint-test-amqpw](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_amqpw.yml/badge.svg)](https://github.com/purposeinplay/go-worker/actions/workflows/lint-test_amqpw.yml)

[![Go Report Card](https://goreportcard.com/badge/github.com/purposeinplay/go-worker)](https://goreportcard.com/report/github.com/purposeinplay/go-worker)

Go Worker is a Go library for performing asynchronous background jobs backed by:

- *Go Routines*: Great for simple application that don't require persistant jobs queues. It uses go routines to implement
- *Redis*: It implements `github.com/gocraft/work` package using Redis as store.
- *RabbitMQ*: A Worker implementation to use with AMQP-compatible brokers (such as RabbitMQ).

## Worker Interface
To add additional implementations, the worker interface needs to be satisfied.

```go
type Worker interface {
    // Start the worker
    Start() error
    // Stop the worker
    Stop() error
    // Perform a job as soon as possible
    Perform(Job) error
    // PerformAt performs a job at a particular time
    PerformAt(Job, time.Time) error
    // PerformIn performs a job after waiting for a specified amount of time
    PerformIn(Job, time.Duration) error
    // Register a Handler
    Register(string, Handler) error
}
```

## Setup
To be able to use background tasks, you’ll need to setup a worker adapter, register job handlers and trigger jobs.

### Step 1: Setup a Worker Adapter

```go
import "github.com/purposeinplay/go-worker/redisw"
import "github.com/gomodule/redigo/redis"

// Make a redis pool
redisPool := &redis.Pool{
    MaxActive: 25,
    MaxIdle:   25,
    Wait:      true,
    Dial: func() (redis.Conn, error) {
        addr := fmt.Sprintf("localhost:%s", port)

        conn, err := redis.Dial("tcp", addr)
        if err != nil {
        return nil, err
        }

        return conn, nil
    },
}

redisWorker, _ := redisw.New(
	WithPool(redisPool)
)

// Start the worker
go func() {
    err = redisWorker.Start()
}()
```

This step is optional if you want to use the goroutines-based worker.

### Step 2: Register a Work Handler
To enqueue jobs, you will need to register a Handler that will be used to run & process the jobs. Each handler has to implement the following interface.

```go
// Handler function that will be run by the worker and given
// a slice of arguments
type Handler func(worker.Args) error
```

```go
import "github.com/purposeinplay/go-worker"

redisWorker.Register("send_email", func(args worker.Args) error {
// do work to send an email
return nil
})
```

### Step 3: Enqueue a Job
Worker handlers are all-set 🙌! Now you’ll need to send jobs to the queue. It is recommended only to use basic types when enqueueing a job. For example, use a model's ID, not the whole model itself.

Please note that if the workers are busy, the process will wait for one to become available.

You can choose to trigger jobs right now or wait for a given time or duration.

#### worker.PerformIn
The `PerformIn` method enqueues the job, so the worker should try and run the job after the duration has passed.

```go
  // Send the send_email job to the queue, and process it in 5 seconds.
	
  w.PerformIn(worker.Job{
    Queue: "default",
    Handler: "send_email",
    Args: worker.Args{
      "user_id": 1,
    },
  }, 5 * time.Second)
```

#### worker.PerformAt

The `PerformAt` method enqueues the job, so the worker should try and run the job at (or near) the time specified.

```go
w.PerformAt(worker.Job{
    Queue: "default",
    Handler: "send_email",
    Args: worker.Args{
      "user_id": 1,
    },
  }, time.Now().Add(5 * time.Second))
```

### Start and Stop Workers
Based on your requirements, starting and stopping workers on demand might be necessary. You can use `w.Stop()` and `w.Start()` for such scenarios.