package asynq

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/purposeinplay/go-worker"
	"github.com/purposeinplay/go-worker/asynq/redisdocker"
	"github.com/stretchr/testify/require"
)

var port string

func newDependencies(t *testing.T) (*Worker, error) {
	t.Helper()

	addr := fmt.Sprintf("localhost:%s", port)
	redisOpt := &asynq.RedisClientOpt{
		Addr: addr,
	}

	asynqWorker, _ := New(WithRedisClientCfg(redisOpt))

	var err error

	go func() {
		err = asynqWorker.Start()
		require.NoError(t, err)
	}()

	t.Cleanup(func() {
		err := asynqWorker.Stop()
		if err != nil {
			t.Fatal(err)
		}
	})

	return asynqWorker, nil
}

func TestMain(m *testing.M) {
	container, err := redisdocker.NewContainer()
	port = container.Port()

	if err != nil {
		log.Fatal("starting redis client error:", err)
	}

	var ret int
	defer func() {
		err := container.Close()
		if err != nil {
			log.Fatal("shutting down redis client error:", err)
		}

		os.Exit(ret)
	}()

	ret = m.Run()
}

func Test_Perform(t *testing.T) {
	asynqWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	asynqWorker.Register("perform", func(worker.Args) error {
		hit = true
		wg.Done()
		return nil
	})

	asynqWorker.Perform(worker.Job{
		Handler: "perform",
	})

	wg.Wait()

	require.True(t, hit)
}

func Test_PerformAt(t *testing.T) {
	asynqWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	asynqWorker.Register("perform_at", func(args worker.Args) error {
		hit = true
		wg.Done()
		return nil
	})

	asynqWorker.PerformAt(worker.Job{
		Handler: "perform_at",
	}, time.Now().Add(5*time.Nanosecond))

	wg.Wait()

	require.True(t, hit)
}

func Test_PerformIn(t *testing.T) {
	asynqWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	asynqWorker.Register("perform_in", func(worker.Args) error {
		hit = true
		wg.Done()
		return nil
	})

	asynqWorker.PerformIn(worker.Job{
		Handler: "perform_in",
	}, 5*time.Nanosecond)

	wg.Wait()

	require.True(t, hit)
}
