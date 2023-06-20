package db

import (
	"context"
	match_frontend "match_frontend/proto"
)

var Default Service

type Service interface {
	Init(ctx context.Context, opts ...Option) error
	Close(ctx context.Context) error
	String() string
	AddToken(ctx context.Context, info *match_frontend.MatchInfo) error
	RemoveToken(ctx context.Context, playerId string, gameId string, subType int64) error
	GetToken(ctx context.Context, playerId string) (*match_frontend.MatchInfo, error)
	GetQueueCount(ctx context.Context, gameId string, subType int64) (int, error)
}

type MatchInfo struct {
	Id       string
	PlayerId string
	Score    int64
	GameId   string
	SubType  int64
}
