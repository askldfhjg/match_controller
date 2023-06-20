package redis

import (
	"context"
	"fmt"
	match_frontend "match_frontend/proto"

	"github.com/gomodule/redigo/redis"
	"github.com/micro/micro/v3/service/logger"
	"google.golang.org/protobuf/proto"
)

func (m *redisBackend) AddToken(ctx context.Context, info *match_frontend.MatchInfo) error {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer handleConnectionClose(&redisConn)

	playerId := info.GetPlayerId()

	value, err := proto.Marshal(info)
	if err != nil {
		return err
	}

	if value == nil {
		return fmt.Errorf("failed to marshal the ticket proto, id: %s: proto: Marshal called with nil", playerId)
	}
	key := fmt.Sprintf(ticketKey, playerId)
	result, err := redis.String(redisConn.Do("SET", key, value, "NX", "EX", 10))
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("%s have in add", playerId)
	}
	zsetKey := fmt.Sprintf(allTickets, info.GameId, info.SubType)
	_, err = redisConn.Do("ZADD", zsetKey, info.Score, playerId)
	if err != nil {
		_, errs := redisConn.Do("DEL", key)
		if errs != nil {
			logger.Error("ZADD and DEL %s have err %s", playerId, errs.Error())
		}
		return err
	}
	return nil
}

func (m *redisBackend) RemoveToken(ctx context.Context, playerId string, gameId string, subType int64) error {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer handleConnectionClose(&redisConn)
	key := fmt.Sprintf(ticketKey, playerId)
	_, err = redisConn.Do("DEL", key)
	if err != nil {
		return err
	}
	zsetKey := fmt.Sprintf(allTickets, gameId, subType)
	_, err = redisConn.Do("ZREM", zsetKey, playerId)
	return err
}

func (m *redisBackend) GetToken(ctx context.Context, playerId string) (*match_frontend.MatchInfo, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)
	key := fmt.Sprintf(ticketKey, playerId)

	bb, err := redis.Bytes(redisConn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	t := &match_frontend.MatchInfo{}
	err = proto.Unmarshal(bb, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (m *redisBackend) GetQueueCount(ctx context.Context, gameId string, subType int64) (int, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer handleConnectionClose(&redisConn)
	zsetKey := fmt.Sprintf(allTickets, gameId, subType)
	return redis.Int(redisConn.Do("ZCARD", zsetKey))
}
