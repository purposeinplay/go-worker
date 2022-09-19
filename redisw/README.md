# Redis adapter for Go Worker

This package implements the github.com/purposeinplay/go-worker/ `Worker` interface using the github.com/gocraft/work package.

# Setup

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

redisWorker, err := redisw.New(
	WithPool(redisPool)
)

if err != nil {
    log.Fatal(err)
}

// Start the worker
go func() {
    err = redisWorker.Start()
}()
```