package center

var DefaultManager Manager

type Manager interface {
	Start() error
	Stop() error
}
