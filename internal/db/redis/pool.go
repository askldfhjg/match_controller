package redis

import (
	"context"
	"fmt"
	"match_controller/internal/db"
	"os"
	ossignal "os/signal"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/micro/micro/v3/service/logger"
	"github.com/pkg/errors"
)

const (
	allTickets     = "allTickets:%s:%d"
	ticketKey      = "ticket:%s"
	poolVersionKey = "poolVersionKey:%s:%d"
)

func New(opts ...db.Option) (db.Service, error) {
	srv := &redisBackend{}
	err := srv.Init(context.Background(), opts...)
	if nil != err {
		log.Fatalf("storage create failed, err:%s", err.Error())
	}

	go func() {
		notify := make(chan os.Signal, 1)
		ossignal.Notify(notify,
			syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
		)
		<-notify
		srv.Close(context.Background())
		log.Info("redis close")
	}()
	return srv, nil
}

type redisBackend struct {
	options   db.Options
	redisPool *redis.Pool
}

func (m *redisBackend) Init(ctx context.Context, opts ...db.Option) error {
	for _, o := range opts {
		o(&m.options)
	}

	if s := os.Getenv("MICRO_REDIS_ADDRESS"); s != "" {
		m.options.Address = s
		log.Infof("MICRO_REDIS_ADDRESS:%s", m.options.Address)
	}

	if s := os.Getenv("MICRO_REDIS_USERNAME"); s != "" {
		m.options.Username = s
		log.Infof("MICRO_REDIS_USERNAME:%s", m.options.Username)
	}

	if s := os.Getenv("MICRO_REDIS_PASSWORD"); s != "" {
		m.options.Password = s
		log.Infof("MICRO_REDIS_PASSWORD:%s", m.options.Password)
	}

	if s := os.Getenv("MICRO_REDIS_IS_SENTINEL"); s != "" {
		m.options.IsSentinel = true
		log.Info("MICRO_REDIS_IS_SENTINEL:true")
	}

	m.redisPool = m.getRedisPool()
	m.ping(ctx)
	return nil
}

// Close the connection to the database.
func (rb *redisBackend) Close(ctx context.Context) error {
	return rb.redisPool.Close()
}

func (m *redisBackend) String() string {
	return "redis"
}

// GetRedisPool configures a new pool to connect to redis given the config.
func (m *redisBackend) getRedisPool() *redis.Pool {
	var dialFunc func(context.Context) (redis.Conn, error)
	maxIdle := m.options.PoolMaxIdle
	maxActive := m.options.PoolMaxActive
	idleTimeout := time.Duration(m.options.PoolIdleTimeout) * time.Second

	if m.options.IsSentinel {
		sentinelPool := m.getSentinelPool()
		dialFunc = func(ctx context.Context) (redis.Conn, error) {
			if ctx != nil && ctx.Err() != nil {
				return nil, ctx.Err()
			}

			sentinelConn, err := sentinelPool.GetContext(ctx)
			if err != nil {
				log.Errorf("failed to connect to redis sentinel. %s", err.Error())
				return nil, err
			}

			masterInfo, err := redis.Strings(sentinelConn.Do("SENTINEL", "GET-MASTER-ADDR-BY-NAME", os.Getenv("redis.sentinelMaster")))
			if err != nil {
				log.Errorf("failed to get current master from redis sentinel. %s", err.Error())
				return nil, err
			}

			masterURL := m.redisURLFromAddr(fmt.Sprintf("%s:%s", masterInfo[0], masterInfo[1]))
			return redis.DialURL(masterURL, redis.DialConnectTimeout(idleTimeout), redis.DialReadTimeout(idleTimeout))
		}
	} else {
		masterURL := m.redisURL()
		dialFunc = func(ctx context.Context) (redis.Conn, error) {
			if ctx != nil && ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return redis.DialURL(masterURL, redis.DialConnectTimeout(idleTimeout), redis.DialReadTimeout(idleTimeout))
		}
	}

	return &redis.Pool{
		MaxIdle:      maxIdle,
		MaxActive:    maxActive,
		IdleTimeout:  idleTimeout,
		Wait:         true,
		TestOnBorrow: testOnBorrow,
		DialContext:  dialFunc,
	}
}

func (m *redisBackend) getSentinelPool() *redis.Pool {
	maxIdle := m.options.PoolMaxIdle
	maxActive := m.options.PoolMaxActive
	idleTimeout := time.Duration(m.options.PoolIdleTimeout) * time.Second

	sentinelURL := m.redisURL()
	return &redis.Pool{
		MaxIdle:      maxIdle,
		MaxActive:    maxActive,
		IdleTimeout:  idleTimeout,
		Wait:         true,
		TestOnBorrow: testOnBorrow,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			if ctx != nil && ctx.Err() != nil {
				return nil, ctx.Err()
			}
			log.Infof("sentinelAddr %s Attempting to connect to Redis Sentinel", sentinelURL)
			return redis.DialURL(sentinelURL, redis.DialConnectTimeout(idleTimeout), redis.DialReadTimeout(idleTimeout))
		},
	}
}

func testOnBorrow(c redis.Conn, lastUsed time.Time) error {
	// Assume the connection is valid if it was used in 30 sec.
	if time.Since(lastUsed) < 15*time.Second {
		return nil
	}

	_, err := c.Do("PING")
	return err
}

func (m *redisBackend) redisURL() string {
	if len(m.options.Password) > 0 {
		return fmt.Sprintf("redis://%s:%s@%s", m.options.Username, m.options.Password, m.options.Address)
	} else {
		return fmt.Sprintf("redis://%s", m.options.Address)
	}
}

func (m *redisBackend) redisURLFromAddr(addr string) string {
	if len(m.options.Password) > 0 {
		return fmt.Sprintf("redis://%s:%s@%s", m.options.Username, m.options.Password, addr)
	} else {
		return "redis://" + addr
	}

}

func handleConnectionClose(conn *redis.Conn) {
	err := (*conn).Close()
	if err != nil {
		log.Errorf("failed to close redis client connection. %s", err.Error())
	}
}

func (m *redisBackend) ping(ctx context.Context) error {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "ping, failed to connect to redis")
	}
	defer handleConnectionClose(&redisConn)
	_, err = redisConn.Do("PING")
	return err
}
