package logger

type Logger interface {
	Info(args ...interface{})
	Infof(msg string, args ...interface{})
	Debugf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
	Panicf(msg string, args ...interface{})
}
