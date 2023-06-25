package controller

var DefaultManager Manager

type Manager interface {
	Start() error
	Stop() error
	AddTask(gameId string) error
}
