# InMemory adapter for Go Worker

This package implements the github.com/purposeinplay/go-worker/ `Worker` interface using go routines.

## Setup

```go
    import "github.com/purposeinplay/go-worker/inmem"

	memWorker, err := inmem.New()
	if err != nil {
		t.Fatal(err)
	}

    // Start the worker
    go func() {
    err = amqpWorker.Start()
    }()
```