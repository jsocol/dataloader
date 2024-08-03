package restdataloader

import "time"

type Option func(*config)

func WithDelay(delay time.Duration) Option {
	return func(c *config) {
		c.delay = delay
	}
}
