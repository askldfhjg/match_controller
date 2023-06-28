package db

import (
	"context"
)

var Default Service

type Service interface {
	Init(ctx context.Context, opts ...Option) error
	Close(ctx context.Context) error
	String() string
	GetQueueCount(ctx context.Context, gameId string, subType int64) (int, error)
	AddPoolVersion(ctx context.Context, gameId string, subType int64, version int64) error
	DelPoolVersion(ctx context.Context, gameId string, subType int64)
}

type MatchInfo struct {
	Id       string
	PlayerId string
	Score    int64
	GameId   string
	SubType  int64
}
