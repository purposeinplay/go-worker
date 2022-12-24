# Redis adapter for Asynq

This package implements the github.com/purposeinplay/go-worker/ `Worker` interface using the github.com/hibiken/asynq/ package.

# Setup

```go
import "github.com/purposeinplay/go-worker/asynq"

redisOpt := &asynq.RedisClientOpt{
  Addr: fmt.Sprintf("localhost:%s", port),
}

asynqWorker, _ := New(WithRedisClientCfg(redisOpt))

if err != nil {
    log.Fatal(err)
}

// Start the worker
go func() {
    err = asynqWorker.Start()
}()
```