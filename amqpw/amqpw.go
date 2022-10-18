package amqpw

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/purposeinplay/go-commons/logs"
	"github.com/purposeinplay/go-commons/rand"
	"github.com/purposeinplay/go-worker"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// ErrInvalidConnection is returned when the Connection opt is not defined.
var ErrInvalidConnection = errors.New("invalid connection")

// Ensures Worker implements the Worker interface.
var _ worker.Worker = &Worker{}

// New creates a new AMQP adapter.
func New(opts ...Option) (*Worker, error) {
	logger, err := logs.NewLogger()
	if err != nil {
		return nil, err
	}

	options := &Options{
		name:           "defaultName",
		logger:         logger,
		maxConcurrency: 25,
	}

	for _, opt := range opts {
		opt.apply(options)
	}

	if options.connection == nil {
		return nil, ErrInvalidConnection
	}

	return &Worker{
		Connection:     options.connection,
		Logger:         options.logger,
		consumerName:   options.name,
		maxConcurrency: options.maxConcurrency,
	}, nil
}

// Worker implements the Worker interface.
type Worker struct {
	Connection     *amqp.Connection
	Channel        *amqp.Channel
	Logger         *zap.Logger
	consumerName   string
	exchange       string
	maxConcurrency int
}

// Start connects to the broker.
func (w *Worker) Start() error {
	// Ensure a Connection exists.
	if w.Connection == nil {
		return ErrInvalidConnection
	}

	// Start a new broker channel.
	c, err := w.Connection.Channel()
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("could not start a new broker channel: %w", err)
	}

	w.Channel = c

	// Exchange declaration
	if w.exchange != "" {
		err = c.ExchangeDeclare(
			w.exchange, // Name
			"direct",   // Type
			true,       // Durable
			false,      // Auto-deleted
			false,      // Internal
			false,      // No wait
			nil,        // Args
		)

		if err != nil {
			return fmt.Errorf("unable to declare exchange: %w", err)
		}
	}

	return nil
}

// Stop closes the connection to the broker.
func (w *Worker) Stop() error {
	w.Logger.Info("stopping AMQP worker")

	if w.Channel == nil {
		return nil
	}

	if err := w.Channel.Close(); err != nil {
		return err
	}

	return w.Connection.Close()
}

// Perform enqueues a new job.
func (w Worker) Perform(job worker.Job) error {
	w.Logger.Info("enqueuing job", zap.Any("job", job))

	err := w.Channel.Publish(
		w.exchange,  // exchange
		job.Handler, // routing key
		true,        // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(job.Args.String()),
		},
	)
	if err != nil {
		w.Logger.Error("error enqueuing job", zap.Any("job", job))

		return fmt.Errorf("error enqueuing job: %w", err)
	}

	return nil
}

// Register consumes a task, using the declared worker.Handler.
func (w *Worker) Register(name string, handler worker.Handler) error {
	w.Logger.Info("register job", zap.Any("job", name))

	_, err := w.Channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		amqp.Table{},
	)
	if err != nil {
		return fmt.Errorf("unable to create queue: %w", err)
	}

	msgs, err := w.Channel.Consume(
		name,
		fmt.Sprintf("%s_%s_%s", w.consumerName, name, rand.String(20)), // nolint: gomnd,revive
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("could not consume queue: %w", err)
	}

	// Process jobs with maxConcurrency workers
	sem := make(chan bool, w.maxConcurrency)

	go func() {
		for msg := range msgs {
			sem <- true

			w.Logger.Info(
				"received job",
				zap.Any("job", name),
				zap.Any("body", msg.Body),
			)

			args := worker.Args{}

			err := json.Unmarshal(msg.Body, &args)
			if err != nil {
				w.Logger.Info(
					"unable to retrieve job",
					zap.Any("job", name),
				)

				continue
			}

			if err := handler(args); err != nil {
				w.Logger.Info(
					"unable to process job",
					zap.Any("job", name),
				)

				continue
			}

			if err := msg.Ack(false); err != nil {
				w.Logger.Info(
					"unable to ack job",
					zap.Any("job", name),
				)
			}
		}

		for i := 0; i < cap(sem); i++ {
			sem <- true
		}
	}()

	return nil
}

// PerformIn performs a job delayed by the given duration.
func (w Worker) PerformIn(job worker.Job, t time.Duration) error {
	w.Logger.Info("enqueuing job", zap.Any("job", job))

	dur := int64(t / time.Millisecond)

	// Trick broker using x-dead-letter feature:
	// the message will be pushed in a temp queue with
	// the given duration as TTL.
	// When the TTL expires, the message is forwarded to the original queue.
	queue, err := w.Channel.QueueDeclare(
		fmt.Sprintf("%s_delayed_%d", job.Handler, dur),
		true, // Save on disk
		true, // Auto-deletion
		false,
		true,
		amqp.Table{
			"x-message-ttl":             dur,
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": job.Handler,
		},
	)
	if err != nil {
		w.Logger.Info(
			"error creating delayed temp queue for job %w",
			zap.Any("job", job.Handler),
		)

		return err
	}

	err = w.Channel.Publish(
		w.exchange, // exchange
		queue.Name, // publish to temp delayed queue
		true,       // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(job.Args.String()),
		},
	)

	if err != nil {
		w.Logger.Info("error enqueuing job %w", zap.Any("job", job.Handler))
		return err
	}

	return nil
}

// PerformAt performs a job at the given time.
func (w Worker) PerformAt(job worker.Job, t time.Time) error {
	return w.PerformIn(job, time.Until(t))
}
