package amqpw

import (
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

// Dialer is the dialer interface.
type Dialer interface {
	Dial(url string) (*amqp.Connection, error)
}

// Dial returns a new AMQP client conn.
func Dial(url string) (*amqp.Connection, error) {
	var rabbit *amqp.Connection

	operation := func() error {
		conn, err := amqp.Dial(url)
		rabbit = conn

		if err != nil {
			return fmt.Errorf(
				"could not establish rabbitmq connection: %w",
				err,
			)
		}

		return nil
	}

	err := backoff.Retry(
		operation,
		backoff.WithMaxRetries(
			backoff.NewExponentialBackOff(),
			5, // nolint: gomnd
		),
	)
	if err != nil {
		return nil, err
	}

	return rabbit, nil
}
