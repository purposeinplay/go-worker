package amqpw

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/purposeinplay/go-commons/rand"
	"github.com/purposeinplay/go-worker"
	"github.com/purposeinplay/go-worker/amqpw/amqpdocker"
	"github.com/stretchr/testify/require"
)

var port string

func newDependencies(t *testing.T) (*Worker, error) {
	t.Helper()

	addr := fmt.Sprintf("amqp://test:test@localhost:%s", port)

	conn, err := Dial(addr)
	if err != nil {
		require.NoError(t, err)
	}

	amqpWorker, err := New(
		WithConnectionOption(conn),
		WithNameOption(rand.String(20)),
	)
	if err != nil {
		require.NoError(t, err)
	}

	err = amqpWorker.Start()
	if err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		err := amqpWorker.Stop()
		if err != nil {
			t.Fatal(err)
		}
	})

	return amqpWorker, nil
}

// Setup the adapter.
func TestMain(m *testing.M) {
	amqpContainer, err := amqpdocker.NewContainer(
		"test",
		"test",
		amqpdocker.WithPort(port),
	)
	if err != nil {
		log.Fatalf("new amqp container: %s", err)
	}

	var ret int
	defer func() {
		err = amqpContainer.Close()
		if err != nil {
			log.Println("err while tearing down amqp container:", err)
		}

		os.Exit(ret)
	}()

	ret = m.Run()
}

func Test_Perform(t *testing.T) {
	amqpWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}

	wg.Add(1)

	amqpWorker.Register("perform", func(job worker.Job) error {
		hit = true
		wg.Done()

		return nil
	})

	result, err := amqpWorker.Perform(worker.Job{
		Handler: "perform",
	})
	require.NoError(t, err)

	fmt.Printf("DDDDDDDDD %+v dddddddd", result)

	wg.Wait()

	require.True(t, hit)
}

func Test_PerformMultiple(t *testing.T) {
	amqpWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hitPerform1, hitPerform2 bool

	wg := &sync.WaitGroup{}
	wg.Add(2)

	amqpWorker.Register("perform1", func(job worker.Job) error {
		hitPerform1 = true
		wg.Done()
		return nil
	})

	amqpWorker.Register("perform2", func(job worker.Job) error {
		hitPerform2 = true
		wg.Done()
		return nil
	})

	amqpWorker.Perform(worker.Job{
		Handler: "perform1",
	})

	amqpWorker.Perform(worker.Job{
		Handler: "perform2",
	})

	wg.Wait()

	require.True(t, hitPerform1)
	require.True(t, hitPerform2)
}

func Test_PerformAt(t *testing.T) {
	amqpWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}

	wg.Add(1)

	amqpWorker.Register("perform_at", func(_ worker.Job) error {
		hit = true
		wg.Done()
		return nil
	})

	amqpWorker.PerformAt(worker.Job{
		Handler: "perform_at",
	}, time.Now().Add(5*time.Nanosecond))

	wg.Wait()

	require.True(t, hit)
}

func Test_PerformIn(t *testing.T) {
	amqpWorker, err := newDependencies(t)
	if err != nil {
		t.Fatal(err)
	}

	var hit bool

	wg := &sync.WaitGroup{}
	wg.Add(1)

	amqpWorker.Register("perform_in", func(job worker.Job) error {
		hit = true
		wg.Done()
		return nil
	})

	amqpWorker.PerformIn(worker.Job{
		Handler: "perform_in",
	}, 5*time.Nanosecond)

	wg.Wait()

	require.True(t, hit)
}
