package redis

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func (m *redisBackend) GetQueueCount(ctx context.Context, gameId string, subType int64) (int, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer handleConnectionClose(&redisConn)
	zsetKey := fmt.Sprintf(allTickets, gameId, subType)
	return redis.Int(redisConn.Do("ZCARD", zsetKey))
}

func (m *redisBackend) AddPoolVersion(ctx context.Context, gameId string, subType int64, version int64) error {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer handleConnectionClose(&redisConn)
	_, err = redisConn.Do("SET", fmt.Sprintf(poolVersionKey, gameId, subType), version)
	return err
}

func (m *redisBackend) DelPoolVersion(ctx context.Context, gameId string, subType int64) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return
	}
	defer handleConnectionClose(&redisConn)
	redisConn.Do("DEL", fmt.Sprintf(poolVersionKey, gameId, subType))
}
