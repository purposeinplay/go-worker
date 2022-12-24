package redisw

import (
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/purposeinplay/go-commons/logs"
	"github.com/purposeinplay/go-worker"
	"go.uber.org/zap"
)

var _ worker.Worker = (*Worker)(nil)

// Worker implements the Worker interface.
type Worker struct {
	enqueuer *work.Enqueuer
	pool     *work.WorkerPool
	logger   *zap.Logger
}

// New returns a new instance of Worker with a predefined config.
func New(opts ...Option) (*Worker, error) {
	logger, err := logs.NewLogger()
	if err != nil {
		return nil, err
	}

	options := &Options{
		maxConcurrency: 25,
		name:           "redisworker",
		logger:         logger,
		pool: &redis.Pool{
			MaxActive: 5,
			MaxIdle:   5,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", ":6379")
			},
		},
	}

	for _, opt := range opts {
		opt.apply(options)
	}

	enqueuer := work.NewEnqueuer(options.name, options.pool)

	pool := work.NewWorkerPool(
		struct{}{},
		options.maxConcurrency,
		options.name,
		options.pool,
	)

	adapter := &Worker{
		enqueuer: enqueuer,
		pool:     pool,
		logger:   logger,
	}

	return adapter, nil
}

// Start starts the adapter event loop.
func (a *Worker) Start() error {
	a.logger.Info("starting worker")
	a.pool.Start()

	return nil
}

// Stop stops the adapter event loop.
func (a *Worker) Stop() error {
	a.logger.Info("stopping worker")
	a.pool.Stop()

	return nil
}

// Register binds a new job, with a name and a handler.
func (a *Worker) Register(name string, h worker.Handler) error {
	a.pool.Job(name, func(job *work.Job) error {
		return h(job.Args)
	})

	return nil
}

// RegisterWithOptions binds a new job, with a name, options and a handler.
func (a *Worker) RegisterWithOptions(
	name string,
	opts work.JobOptions,
	h worker.Handler,
) error {
	a.pool.JobWithOptions(name, opts, func(job *work.Job) error {
		return h(job.Args)
	})

	return nil
}

// Perform sends a new job to the queue, now.
func (a *Worker) Perform(job worker.Job) error {
	a.logger.Info("enqueuing job", zap.String("job", job.String()))

	_, err := a.enqueuer.Enqueue(job.Handler, job.Args)
	if err != nil {
		a.logger.Error(
			"error enqueuing job",
			zap.String("job", job.String()),
		)

		return errors.WithStack(err)
	}

	return nil
}

// PerformIn sends a new job to the queue, with a given delay.
func (a *Worker) PerformIn(job worker.Job, t time.Duration) error {
	a.logger.Info("enqueuing job", zap.String("job", job.String()))

	d := int64(t / time.Second)

	_, err := a.enqueuer.EnqueueIn(job.Handler, d, job.Args)
	if err != nil {
		a.logger.Error(
			"error enqueuing job",
			zap.String("job", job.String()),
		)

		return errors.WithStack(err)
	}

	return nil
}

// PerformAt sends a new job to the queue, with a given start time.
func (a *Worker) PerformAt(job worker.Job, t time.Time) error {
	return a.PerformIn(job, time.Until(t))
}
