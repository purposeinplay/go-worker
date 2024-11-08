package asynq

import (
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Option describes how options should be implemented.
type Option interface {
	apply(*Options)
}

// Options are used to configure the adapter config.
type Options struct {
	redisClientCfg *asynq.RedisClientOpt
	cfg            asynq.Config
	logger         *zap.Logger
}

// WithRedisClientCfg configures the RedisClientCfg option.
func WithRedisClientCfg(r *asynq.RedisClientOpt) Option {
	return redisClientCfgOption{RedisClientCfg: r}
}

type redisClientCfgOption struct {
	RedisClientCfg *asynq.RedisClientOpt
}

func (o redisClientCfgOption) apply(opts *Options) {
	opts.redisClientCfg = o.RedisClientCfg
}

// WithConfig configures the Asynq config option.
func WithConfig(r asynq.Config) Option {
	return cfgOption{Cfg: r}
}

type cfgOption struct {
	Cfg asynq.Config
}

func (o cfgOption) apply(opts *Options) {
	opts.cfg = o.Cfg
}

func WithQueues(queues map[string]int) Option {
	return optionFunc(func(opts *Options) {
		opts.cfg.Queues = queues
	})
}

type optionFunc func(*Options)

func (f optionFunc) apply(opts *Options) {
	f(opts)
}
