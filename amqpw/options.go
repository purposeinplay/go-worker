package amqpw

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// Option describes how options should be implemented.
type Option interface {
	apply(*Options)
}

// Options are used to configure the AMQP worker adapter.
type Options struct {
	// Name is used to identify the app as a consumer.
	name string

	// Connection is the AMQP connection to use.
	connection *amqp.Connection

	// Exchange is used to customize the AMQP exchange name.
	exchange string

	// MaxConcurrency restricts the amount of workers in parallel.
	maxConcurrency int

	// Logger is a logger interface to write the worker logs.
	logger *zap.Logger
}

// WithNameOption configures the name option.
func WithNameOption(name string) Option {
	return nameOption(name)
}

type nameOption string

func (o nameOption) apply(opts *Options) {
	opts.name = string(o)
}

// WithConnectionOption configures the connection option.
func WithConnectionOption(conn *amqp.Connection) Option {
	return connectionOption{connection: conn}
}

type connectionOption struct {
	connection *amqp.Connection
}

func (o connectionOption) apply(opts *Options) {
	opts.connection = o.connection
}

// WithExchangeOption configures the exchange option.
func WithExchangeOption(exchange string) Option {
	return exchangeOption(exchange)
}

type exchangeOption string

func (o exchangeOption) apply(opts *Options) {
	opts.exchange = string(o)
}

// WithMaxConcurrency configures the maxConcurrency option.
func WithMaxConcurrency(v int) Option {
	return maxConcurrencyOption(v)
}

type maxConcurrencyOption int

func (o maxConcurrencyOption) apply(opts *Options) {
	opts.maxConcurrency = int(o)
}

// WithLogger configures the logger option.
func WithLogger(l *zap.Logger) Option {
	return loggerOption{Logger: l}
}

type loggerOption struct {
	Logger *zap.Logger
}

func (o loggerOption) apply(opts *Options) {
	opts.logger = o.Logger
}
