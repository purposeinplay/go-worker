package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/purposeinplay/go-commons/logs"
	"github.com/purposeinplay/go-worker"
	"go.uber.org/zap"
)

var _ worker.Worker = (*Worker)(nil)

// Worker implements the Worker interface.
type Worker struct {
	server    *asynq.Server
	client    *asynq.Client
	inspector *asynq.Inspector
	mux       *asynq.ServeMux
	logger    *zap.Logger
}

const (
	// QueueCritical represents a critical queue.
	QueueCritical = "critical"
	// QueueDefault represents a default queue.
	QueueDefault = "default"
)

// New returns a new instance of Worker with a predefined config.
func New(opts ...Option) (*Worker, error) {
	logger, err := logs.NewLogger()
	if err != nil {
		return nil, err
	}

	options := &Options{
		redisClientCfg: &asynq.RedisClientOpt{
			Addr: ":6379",
		},
		cfg: asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					logger.
						Error(
							fmt.Errorf("process task failed %w", err).Error(),
							zap.String("type", task.Type()),
							zap.ByteString("payload", task.Payload()),
						)
				}),
			Logger: logger.Sugar(),
		},
		logger: logger,
	}

	for _, opt := range opts {
		opt.apply(options)
	}

	server := asynq.NewServer(
		options.redisClientCfg,
		options.cfg,
	)

	mux := asynq.NewServeMux()

	client := asynq.NewClient(options.redisClientCfg)

	inspector := asynq.NewInspector(options.redisClientCfg)

	return &Worker{
		client:    client,
		server:    server,
		inspector: inspector,
		mux:       mux,
		logger:    logger,
	}, nil
}

// Start the worker.
func (w Worker) Start() error {
	return w.server.Start(w.mux)
}

// Stop the worker.
func (w Worker) Stop() error {
	w.server.Stop()
	return nil
}

// Perform a job as soon as possible.
func (w Worker) Perform(job worker.Job) (*worker.JobInfo, error) {
	w.logger.Info("enqueuing job", zap.String("job", job.String()))

	payload, err := json.Marshal(job.Args)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(job.Handler, payload)

	enqueueOpts := []asynq.Option{}
	if job.Queue != "" {
		enqueueOpts = append(enqueueOpts, asynq.Queue(job.Queue))
	}

	enqueue, err := w.client.Enqueue(task, enqueueOpts...)
	if err != nil {
		w.logger.Error(
			"error enqueuing job",
			zap.Error(err),
			zap.String("job", job.String()),
		)

		return nil, errors.WithStack(err)
	}

	w.logger.Info(
		"enqueued job",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("queue", enqueue.Queue),
		zap.Int("max_retry", enqueue.MaxRetry),
	)

	return &worker.JobInfo{
		ID:           enqueue.ID,
		Retries:      enqueue.Retried,
		LastFailedAt: enqueue.LastFailedAt,
	}, nil
}

// PerformAt performs a job at a particular time.
func (w Worker) PerformAt(
	job worker.Job,
	jobTime time.Time,
) (
	*worker.JobInfo,
	error,
) {
	w.logger.Info(
		"enqueuing job",
		zap.String("job", job.String()),
		zap.Time("time", jobTime),
	)

	opts, err := json.Marshal(job.Args)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(job.Handler, opts)

	enqueueOpts := []asynq.Option{asynq.ProcessAt(jobTime)}
	if job.Queue != "" {
		enqueueOpts = append(enqueueOpts, asynq.Queue(job.Queue))
	}

	enqueue, err := w.client.Enqueue(task, enqueueOpts...)
	if err != nil {
		w.logger.Error(
			"error enqueuing job",
			zap.Error(err),
			zap.String("job", job.String()),
		)

		return nil, errors.WithStack(err)
	}

	w.logger.Info(
		"enqueued job",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("queue", enqueue.Queue),
		zap.Time("time", jobTime),
		zap.Int("max_retry", enqueue.MaxRetry),
	)

	return &worker.JobInfo{
		ID:           enqueue.ID,
		Retries:      enqueue.Retried,
		LastFailedAt: enqueue.LastFailedAt,
	}, nil
}

// PerformIn performs a job after waiting for a specified duration.
func (w Worker) PerformIn(
	job worker.Job,
	jobTime time.Duration,
) (*worker.JobInfo, error) {
	w.logger.Info(
		"enqueuing job",
		zap.String("job", job.String()),
		zap.Duration("duration", jobTime),
	)

	payload, err := json.Marshal(job.Args)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(job.Handler, payload)

	enqueueOpts := []asynq.Option{asynq.ProcessIn(jobTime)}
	if job.Queue != "" {
		enqueueOpts = append(enqueueOpts, asynq.Queue(job.Queue))
	}

	enqueue, err := w.client.Enqueue(task, enqueueOpts...)
	if err != nil {
		w.logger.Error(
			"error enqueuing job",
			zap.Error(err),
			zap.String("job", job.String()),
		)

		return nil, errors.WithStack(err)
	}

	w.logger.Info(
		"enqueued job",
		zap.String("type", task.Type()),
		zap.ByteString("payload", task.Payload()),
		zap.String("queue", enqueue.Queue),
		zap.String("job", job.String()),
		zap.Int("max_retry", enqueue.MaxRetry),
	)

	return &worker.JobInfo{
		ID:           enqueue.ID,
		Retries:      enqueue.Retried,
		LastFailedAt: enqueue.LastFailedAt,
	}, nil
}

// Register Handler with the worker.
func (w Worker) Register(name string, h worker.Handler) error {
	w.mux.HandleFunc(name, func(ctx context.Context, task *asynq.Task) error {
		var payload worker.Args
		err := json.Unmarshal(task.Payload(), &payload)
		if err != nil {
			return err
		}

		return h(worker.Job{
			ID:      task.ResultWriter().TaskID(),
			Handler: task.Type(),
			Args:    payload,
		})
	})

	return nil
}

// DeleteJob removes a job from the queue.
func (w Worker) DeleteJob(queue, jobID string) error {
	err := w.inspector.DeleteTask(queue, jobID)
	if err != nil {
		return err
	}

	return nil
}
