package redisw

import (
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

// Option describes how options should be implemented.
type Option interface {
	apply(*Options)
}

// Options are used to configure the adapter config.
type Options struct {
	pool           *redis.Pool
	name           string
	maxConcurrency uint
	logger         *zap.Logger
}

// WithName configures the name option.
func WithName(name string) Option {
	return nameOption(name)
}

type nameOption string

func (o nameOption) apply(opts *Options) {
	opts.name = string(o)
}

// WithMaxConcurrency configures the maxConcurrency option.
func WithMaxConcurrency(v int) Option {
	return maxConcurrencyOption(v)
}

type maxConcurrencyOption int

func (o maxConcurrencyOption) apply(opts *Options) {
	opts.maxConcurrency = uint(o)
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

// WithPool configures the Pool option.
func WithPool(p *redis.Pool) Option {
	return poolOption{Pool: p}
}

type poolOption struct {
	Pool *redis.Pool
}

func (o poolOption) apply(opts *Options) {
	opts.pool = o.Pool
}
