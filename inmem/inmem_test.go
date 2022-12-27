package inmem

import (
	"sync"
	"testing"
	"time"

	"github.com/purposeinplay/go-worker"
	"github.com/stretchr/testify/require"
)

func Test_InMem_RegisterEmpty(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Register("", func(job worker.Job) error {
		return nil
	})
	require.Error(t, err)
}

func Test_InMem_RegisterNil(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Register("sample", nil)
	require.Error(t, err)
}

func Test_InMem_RegisterEmptyNil(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Register("", nil)
	require.Error(t, err)
}

func Test_InMem_RegisterExisting(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Register("sample", func(job worker.Job) error {
		return nil
	})
	require.NoError(t, err)

	err = memWorker.Register("sample", func(job worker.Job) error {
		return nil
	})
	require.Error(t, err)
}

func Test_InMem_StartStop(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Start()
	require.NoError(t, err)

	err = memWorker.Stop()
	require.NoError(t, err)
}

func Test_InMem_Perform(t *testing.T) {
	t.Parallel()

	var hit bool

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Start())

	err = memWorker.Register("x", func(job worker.Job) error {
		require.NotNil(t, job.ID)
		require.Equal(t, "x", job.Handler)
		hit = true
		return nil
	})
	require.NoError(t, err)

	result, err := memWorker.Perform(worker.Job{
		Handler: "x",
	})
	require.NotNil(t, result.ID)

	// the worker should guarantee the job is finished before the worker stopped
	require.NoError(t, memWorker.Stop())
	require.True(t, hit)
}

func Test_InMem_PerformBroken(t *testing.T) {
	t.Parallel()

	var hit bool

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Start())

	err = memWorker.Register("x", func(job worker.Job) error {
		hit = true

		println([]string{}[0])

		return nil
	})

	memWorker.Perform(worker.Job{
		Handler: "x",
	})

	require.NoError(t, memWorker.Stop())
	require.True(t, hit)
}

func Test_InMem_PerformWithEmptyJob(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Start())

	defer memWorker.Stop()

	_, err = memWorker.Perform(worker.Job{})
	require.Error(t, err)
}

func Test_InMem_PerformWithUnknownJob(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Start())

	defer memWorker.Stop()

	_, err = memWorker.Perform(worker.Job{Handler: "unknown"})
	require.Error(t, err)
}

func Test_InMem_PerformBeforeStart(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Register("sample", func(job worker.Job) error {
		return nil
	}))

	_, err = memWorker.Perform(worker.Job{Handler: "sample"})
	require.Error(t, err)
}

func Test_InMem_PerformAfterStop(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Register("sample", func(job worker.Job) error {
		return nil
	}))
	require.NoError(t, memWorker.Start())
	require.NoError(t, memWorker.Stop())

	_, err = memWorker.Perform(worker.Job{Handler: "sample"})
	require.Error(t, err)
}

func Test_InMem_PerformAt(t *testing.T) {
	t.Parallel()

	var hit bool

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Start())

	wg := &sync.WaitGroup{}
	wg.Add(1)

	memWorker.Register("x", func(job worker.Job) error {
		hit = true
		wg.Done()
		return nil
	})

	memWorker.PerformAt(worker.Job{
		Handler: "x",
	}, time.Now().Add(5*time.Millisecond))

	// how long does the handler take for assignment? hmm,
	time.Sleep(100 * time.Millisecond)
	wg.Wait()
	require.True(t, hit)

	require.NoError(t, memWorker.Stop())
}

func Test_InMem_PerformIn(t *testing.T) {
	t.Parallel()

	var hit bool

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Start())

	wg := &sync.WaitGroup{}
	wg.Add(1)

	memWorker.Register("x", func(job worker.Job) error {
		hit = true
		wg.Done()
		return nil
	})

	memWorker.PerformIn(worker.Job{
		Handler: "x",
	}, 5*time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	wg.Wait()
	require.True(t, hit)

	require.NoError(t, memWorker.Stop())
}

func Test_InMem_PerformInBeforeStart(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, memWorker.Register("sample", func(job worker.Job) error {
		return nil
	}))

	_, err = memWorker.PerformIn(
		worker.Job{Handler: "sample"}, 5*time.Millisecond,
	)
	require.Error(t, err)
}

func Test_InMem_PerformInAfterStop(t *testing.T) {
	t.Parallel()

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Register("sample", func(job worker.Job) error {
		return nil
	})
	require.NoError(t, err)

	err = memWorker.Start()
	require.NoError(t, err)

	err = memWorker.Stop()
	require.NoError(t, err)

	_, err = memWorker.PerformIn(worker.Job{Handler: "sample"}, 5*time.Millisecond)
	require.Error(t, err)
}

// Stop blocks any pending jobs from being executed.
func Test_InMem_PerformInFollowedByStop(t *testing.T) {
	t.Parallel()

	var hit bool

	memWorker, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = memWorker.Start()
	require.NoError(t, err)

	memWorker.Register("sample", func(job worker.Job) error {
		hit = true
		return nil
	})

	_, err = memWorker.PerformIn(worker.Job{
		Handler: "sample",
	}, 300*time.Millisecond)

	require.NoError(t, err)

	err = memWorker.Stop()
	require.NoError(t, err)

	require.False(t, hit)
}
