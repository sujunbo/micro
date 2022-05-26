package log

type Log interface {
	Debug(msg string)
	Debugf(format string, a ...interface{})

	Info(msg string)
	Infof(format string, a ...interface{})

	Warn(msg string)
	Warnf(format string, a ...interface{})

	Error(msg string)
	Errorf(format string, a ...interface{})
}

type Fields map[string]interface{}

var (
	logger Log
)

func GetInstance() Log {
	return logger
}

func InitLogger(l *Option) {
	logger = NewLogger(l)
}

func Debug(msg string) {
	logger.Debug(msg)
}

func Debugf(format string, a ...interface{}) {
	logger.Debugf(format, a...)
}

func Info(msg string) {
	logger.Info(msg)
}

func Infof(format string, a ...interface{}) {
	logger.Infof(format, a...)
}

func Warn(msg string) {
	logger.Warn(msg)
}

func Warnf(format string, a ...interface{}) {
	logger.Warnf(format, a...)
}

func Error(msg string) {
	logger.Error(msg)
}

func Errorf(format string, a ...interface{}) {
	logger.Errorf(format, a...)
}
