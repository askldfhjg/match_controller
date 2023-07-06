package db

import (
	"context"
)

var Default Service

type Service interface {
	Init(ctx context.Context, opts ...Option) error
	Close(ctx context.Context) error
	String() string
	GetQueueCounts(ctx context.Context, version int64, gameId string, subType int64, groupCount int) ([]int, error)
	AddPoolVersion(ctx context.Context, gameId string, subType int64, version int64) (int64, error)
	DelPoolVersion(ctx context.Context, gameId string, subType int64)
	InitPoolVersion(ctx context.Context, gameId string, subType int64, version int64) error
}

type MatchInfo struct {
	Id       string
	PlayerId string
	Score    int64
	GameId   string
	SubType  int64
}
