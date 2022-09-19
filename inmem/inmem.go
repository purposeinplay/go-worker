package inmem

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/purposeinplay/go-commons/logs"
	"github.com/purposeinplay/go-worker"
	"go.uber.org/zap"
)

var (
	errNoHandler             = errors.New("name or handler cannot be empty/nil")
	errHandlerExists         = errors.New("handler already exists")
	errWorkerStopped         = errors.New("worker has not started")
	errJobHandlerNameEmpty   = errors.New("no job handler name given")
	errJobHandlerDoesntExist = errors.New("job handler doesn't exist")
)

var _ worker.Worker = (*Worker)(nil)

// Worker is a simple implementation of the Worker interface.
type Worker struct {
	logger   *zap.Logger
	handlers map[string]worker.Handler
	mu       *sync.Mutex
	wg       sync.WaitGroup
	started  atomic.Bool
}

// New creates a basic implementation of the Worker interface
// that is backed using just the standard library and goroutines.
func New() (*Worker, error) {
	logger, err := logs.NewLogger()

	defer func() {
		_ = logger.Sync()
	}()

	if err != nil {
		return nil, fmt.Errorf("could not create logger %w", err)
	}

	return &Worker{
		logger:   logger,
		handlers: map[string]worker.Handler{},
		mu:       &sync.Mutex{},
		started:  atomic.Bool{},
	}, nil
}

// Register Handler with the worker.
func (w *Worker) Register(name string, h worker.Handler) error {
	if name == "" || h == nil {
		return errNoHandler
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.handlers[name]; ok {
		return errHandlerExists
	}

	w.handlers[name] = h

	return nil
}

// Start the worker.
func (w *Worker) Start() error {
	w.logger.Info("starting worker")

	w.mu.Lock()
	defer w.mu.Unlock()

	w.started.Store(true)

	return nil
}

// Stop the worker.
func (w *Worker) Stop() error {
	// prevent job submission when stopping.
	w.mu.Lock()
	defer w.mu.Unlock()

	w.logger.Info("stopping worker")

	w.started.Store(false)

	w.wg.Wait()
	w.logger.Info("worker jobs stopped completely")

	return nil
}

// Perform a job as soon as possible.
func (w *Worker) Perform(job worker.Job) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.started.Load() {
		return errWorkerStopped
	}

	w.logger.Debug("performing job %s", zap.String("job", job.String()))

	if job.Handler == "" {
		return errJobHandlerNameEmpty
	}

	if h, ok := w.handlers[job.Handler]; ok {
		w.wg.Add(1)

		go func() {
			defer w.wg.Done()

			err := safeProcess(func() error {
				return h(job.Args)
			})
			if err != nil {
				w.logger.Error("safeProcess err", zap.Error(err))
			}

			w.logger.Debug("completed job %s", zap.String("job", job.String()))
		}()

		return nil
	}

	return errJobHandlerDoesntExist
}

// safeProcess run the function safely knowing that if it panics
// the panic will be caught and returned as an error.
func safeProcess(fn func() error) (err error) {
	defer func() {
		if ex := recover(); ex != nil {
			if e, ok := ex.(error); ok {
				err = e
				return
			}

			err = errors.New(fmt.Sprint(ex)) // nolint: goerr113
		}
	}()

	return fn()
}

// PerformAt performs a job at a particular time.
func (w *Worker) PerformAt(job worker.Job, t time.Time) error {
	return w.PerformIn(job, time.Until(t))
}

// PerformIn performs a job after waiting for a specified duration.
func (w *Worker) PerformIn(job worker.Job, dur time.Duration) error {
	w.wg.Add(1) // waiting job also should be counted

	if !w.started.Load() {
		return errWorkerStopped
	}

	go func() {
		defer w.wg.Done()

		waiting := 100 * time.Millisecond

		ticker := time.NewTicker(waiting)

	loop:
		for { // nolint: gosimple
			select {
			case <-ticker.C:
				if w.started.Load() {
					break loop
				}

				dur -= waiting // nolint: revive
				if dur < 0 {
					if !w.started.Load() {
						break loop
					}
				}
			}
		}

		<-time.After(dur)

		if w.started.Load() {
			_ = w.Perform(job)
		}
	}()

	return nil
}
