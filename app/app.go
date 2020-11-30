package app

type App interface {
	Name() string
	Init(...Option)
	Run() error
}

func NewApp(opt ...Option) App {
	return newApp(opt...)
}
