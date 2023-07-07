package redis

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const (
	allTickets     = "allTickets:%d:%s:%d"
	poolVersionKey = "poolVersionKey:%s:%d"
	taskFlag       = "taskFlag:%d:%s:%d"
	poolLock       = "poolLock:%s:%d"
)

func (m *redisBackend) GetQueueCounts(ctx context.Context, oldVersion int64, gameId string, subType int64, groupCount int) ([]int, error) {
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
	keys := []interface{}{fmt.Sprintf(allTickets, oldVersion, gameId, subType)}
	params := []interface{}{script, len(keys)}
	params = append(params, keys...)
	params = append(params, args...)
	return redis.Ints(redisConn.Do("EVAL", params...))
}

func (m *redisBackend) AddPoolVersion(ctx context.Context, gameId string, subType int64, version int64) (int64, int64, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer handleConnectionClose(&redisConn)

	script := `
	if(#ARGV < 1) then
		error("argv error")
	end
	local values = redis.call('HGETALL', KEYS[1])
	local result = {}
	if(#values <= 0) then
		result["poolVersionKey"] = "0"
		result["lastPoolVersionKey"] = "0"
	else
		for i = 1, #values, 2 do
			local key = values[i]
			local value = values[i + 1]
			result[key] = value
		end
		if(result["poolVersionKey"] == nil) then
			result["poolVersionKey"] = "0"
		end
		if(result["lastPoolVersionKey"] == nil) then
			result["lastPoolVersionKey"] = "0"
		end
	end
	redis.call('HSET', KEYS[1], 'poolVersionKey', ARGV[1], 'lastPoolVersionKey', result["poolVersionKey"])
	return {result["poolVersionKey"], result["lastPoolVersionKey"]}
	`

	args := []interface{}{version}
	keys := []interface{}{fmt.Sprintf(poolVersionKey, gameId, subType)}
	params := []interface{}{script, len(keys)}
	params = append(params, keys...)
	params = append(params, args...)
	rr, err := redis.Int64s(redisConn.Do("EVAL", params...))
	if err != nil {
		return 0, 0, err
	}
	return rr[0], rr[1], nil
}

// func (m *redisBackend) DelPoolVersion(ctx context.Context, gameId string, subType int64) {
// 	redisConn, err := m.redisPool.GetContext(ctx)
// 	if err != nil {
// 		return
// 	}
// 	defer handleConnectionClose(&redisConn)
// 	redisConn.Do("DEL", fmt.Sprintf(poolVersionKey, gameId, subType))
// }

func (m *redisBackend) InitPoolVersion(ctx context.Context, gameId string, subType int64, version int64) error {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer handleConnectionClose(&redisConn)
	_, err = redisConn.Do("HSETNX", fmt.Sprintf(poolVersionKey, gameId, subType), "poolVersionKey", version)
	return err
}

func (m *redisBackend) GetTaskFlag(ctx context.Context, gameId string, subType int64, version int64) (map[string]string, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)
	return redis.StringMap(redisConn.Do("HGETALL", fmt.Sprintf(taskFlag, version, gameId, subType)))
}

func (m *redisBackend) SetTaskFlag(ctx context.Context, gameId string, subType int64, version int64, info map[string]string) (int, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer handleConnectionClose(&redisConn)
	queryParams := make([]interface{}, len(info)*2+1)
	index := 0
	queryParams[index] = fmt.Sprintf(taskFlag, version, gameId, subType)
	for k, v := range info {
		index++
		queryParams[index] = k
		index++
		queryParams[index] = v
	}
	return redis.Int(redisConn.Do("HSET", queryParams...))
}

func (m *redisBackend) TryLockPool(ctx context.Context, gameId string, subType int64, version int64) (int64, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer handleConnectionClose(&redisConn)
	return redis.Int64(redisConn.Do("SET", fmt.Sprintf(poolLock, gameId, subType), version, "NX", "GET", "EX", 10))
}

func (m *redisBackend) ProcessLastTask(ctx context.Context, gameId string, subType int64, version int64, startPos string, endPos string) (map[string]string, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)

	zsetKey := fmt.Sprintf(allTickets, version, gameId, subType)
	reply, err := redis.StringMap(redisConn.Do("ZRANGEBYSCORE", zsetKey, startPos, endPos, "WITHSCORES"))
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply, nil
}

func (m *redisBackend) MoveTokens(ctx context.Context, version int64, retDetail map[string]string, gameId string, subType int64) (int, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer handleConnectionClose(&redisConn)
	if len(retDetail) <= 0 {
		return 0, err
	}
	zsetKey := fmt.Sprintf(allTickets, version, gameId, subType)

	queryParams := make([]interface{}, len(retDetail)*2+1)
	index := 0
	queryParams[index] = zsetKey

	for id, score := range retDetail {
		index++
		queryParams[index] = score
		index++
		queryParams[index] = id
	}
	return redis.Int(redisConn.Do("ZADD", queryParams...))

}
