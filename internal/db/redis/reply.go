package redis

type RedisReply struct {
	Reply interface{}
	Err   error
}

func reply(reply interface{}, err error) *RedisReply {
	return &RedisReply{reply, err}
}
