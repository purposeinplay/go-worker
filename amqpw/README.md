# AMQP adapter for Go Worker

This package implements the github.com/purposeinplay/go-worker/ `Worker` interface using the `github.com/streadway/amqp` package.

# Setup

```go
import "github.com/purposeinplay/go-worker/amqpw"
import "github.com/streadway/amqp"

conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
if err != nil {
    log.Fatal(err)
}

amqpWorker, err := amqpw.New(
    WithConnectionOption(conn),
    WithNameOption("myapp"),
    WithMaxConcurrency(25),
)

if err != nil {
    log.Fatal(err)
}

// Start the worker
go func() {
    err = amqpWorker.Start()
}()
```