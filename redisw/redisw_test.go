package redisw

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/purposeinplay/go-worker"
	"github.com/purposeinplay/go-worker/redisw/redisdocker"
	"github.com/stretchr/testify/require"
)

var port string

func newDependencies(t *testing.T) (*Worker, error) {
	t.Helper()

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

	conn := redisPool.Get()
	_, err := redis.String(conn.Do("PING"))
	require.NoError(t, err)

	redisWorker, _ := New(WithPool(redisPool))

	go func() {
		err = redisWorker.Start()
		require.NoError(t, err)
	}()

	t.Cleanup(func() {
		err := redisWorker.Stop()
		if err != nil {
			t.Fatal(err)
		}

		err = redisPool.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	return redisWorker, nil
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
	redisWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	redisWorker.Register("perform", func(job worker.Job) error {
		hit = true
		require.NotNil(t, job.ID)
		require.Equal(t, "perform", job.Handler)
		wg.Done()
		return nil
	})

	result, err := redisWorker.Perform(worker.Job{
		Handler: "perform",
	})
	require.NoError(t, err)

	wg.Wait()

	require.NotNil(t, result.ID)
	require.True(t, hit)
}

func Test_PerformAt(t *testing.T) {
	redisWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	redisWorker.Register("perform_at", func(args worker.Job) error {
		hit = true
		wg.Done()
		return nil
	})

	redisWorker.PerformAt(worker.Job{
		Handler: "perform_at",
	}, time.Now().Add(5*time.Nanosecond))

	wg.Wait()

	require.True(t, hit)
}

func Test_PerformIn(t *testing.T) {
	redisWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	redisWorker.Register("perform_in", func(worker.Job) error {
		hit = true
		wg.Done()
		return nil
	})

	redisWorker.PerformIn(worker.Job{
		Handler: "perform_in",
	}, 5*time.Nanosecond)

	wg.Wait()

	require.True(t, hit)
}
