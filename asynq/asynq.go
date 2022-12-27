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
	server *asynq.Server
	client *asynq.Client
	mux    *asynq.ServeMux
	logger *zap.Logger
}

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// New returns a new instance of Worker with a predefined config.
func New(opts ...Option) (*Worker, error) {
	logger, err := logs.NewLogger()
	if err != nil {
		return nil, err
	}

	options := &Options{
		redisClientCfg: &asynq.RedisClientOpt{
			Addr: fmt.Sprintf(":6379"),
		},
		cfg: asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
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

	return &Worker{
		client: client,
		server: server,
		mux:    mux,
		logger: logger,
	}, nil
}

func (w Worker) Start() error {
	return w.server.Start(w.mux)
}

func (w Worker) Stop() error {
	w.server.Stop()
	return nil
}

func (w Worker) Perform(job worker.Job) (*worker.JobInfo, error) {
	w.logger.Info("enqueuing job", zap.String("job", job.String()))
	payload, err := json.Marshal(job.Args)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(job.Handler, payload)
	enqueue, err := w.client.Enqueue(task)
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

func (w Worker) PerformAt(
	job worker.Job,
	t time.Time,
) (
	*worker.JobInfo,
	error,
) {
	w.logger.Info(
		"enqueuing job",
		zap.String("job", job.String()),
		zap.Time("time", t),
	)
	opts, err := json.Marshal(job.Args)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(job.Handler, opts)
	enqueue, err := w.client.Enqueue(task, asynq.ProcessAt(t))
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
		zap.Time("time", t),
		zap.Int("max_retry", enqueue.MaxRetry),
	)

	return &worker.JobInfo{
		ID:           enqueue.ID,
		Retries:      enqueue.Retried,
		LastFailedAt: enqueue.LastFailedAt,
	}, nil
}

func (w Worker) PerformIn(
	job worker.Job,
	t time.Duration,
) (*worker.JobInfo, error) {
	w.logger.Info(
		"enqueuing job",
		zap.String("job", job.String()),
		zap.Duration("duration", t),
	)
	payload, err := json.Marshal(job.Args)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(job.Handler, payload)
	enqueue, err := w.client.Enqueue(task, asynq.ProcessIn(t))
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
