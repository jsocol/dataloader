package dataloader

import "time"

type Option func(*config)

func WithDelay(delay time.Duration) Option {
	return func(c *config) {
		c.delay = delay
	}
}

func WithMaxBatch(batchSize int) Option {
	return func(c *config) {
		c.maxBatch = batchSize
	}
}
