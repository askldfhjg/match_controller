package redis

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const (
	allTickets     = "allTickets:%s:%d"
	poolVersionKey = "poolVersionKey:%s:%d"
)

func (m *redisBackend) GetQueueCounts(ctx context.Context, gameId string, subType int64, groupCount int) ([]int, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)

	script := `local count = redis.call('ZCARD', KEYS[1])
	if(not count) then
		return {}
	end
	if(not ARGV[1]) then
		error("groupCount error")
	end
	local groupCount = tonumber(ARGV[1])
	if(groupCount <= 0) then
		error("groupCount <= 0")
	end
	local segmentSize = math.ceil(count / groupCount)
	local segmentBoundaries = {}

	for i = 1, segmentSize+1 do
		local startRank = ((i-1) * groupCount)
		if(startRank > count) then
			startRank = -1
		end
		local segmentMembers = redis.call("ZRANGE", KEYS[1], startRank, startRank, "WITHSCORES")
		local score = tonumber(segmentMembers[2])
		table.insert(segmentBoundaries, score)
	end
	return segmentBoundaries
	`

	args := []interface{}{groupCount}
	keys := []interface{}{fmt.Sprintf(allTickets, gameId, subType)}
	params := []interface{}{script, len(keys)}
	params = append(params, keys...)
	params = append(params, args...)
	return redis.Ints(redisConn.Do("EVAL", params...))
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
