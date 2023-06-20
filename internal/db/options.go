package db

type Options struct {
	IsSentinel bool
	//HealthCheckTimeout int
	Address         string
	Username        string
	Password        string
	PoolMaxIdle     int
	PoolMaxActive   int
	PoolIdleTimeout int
}

type Option func(options *Options)

func WithAddress(addr string) Option {
	return func(options *Options) {
		options.Address = addr
	}
}

// func WithHealthCheckTimeout(timeout int) Option {
// 	return func(options *Options) {
// 		options.HealthCheckTimeout = timeout
// 	}
// }

func WithPoolMaxIdle(cc int) Option {
	return func(options *Options) {
		options.PoolMaxIdle = cc
	}
}

func WithPoolMaxActive(cc int) Option {
	return func(options *Options) {
		options.PoolMaxActive = cc
	}
}

func WithPoolIdleTimeout(cc int) Option {
	return func(options *Options) {
		options.PoolIdleTimeout = cc
	}
}
