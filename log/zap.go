package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Levels = []string{"debug", "info", "warn", "error"}
)

func fileLevels(level string) (levels []string) {
	for i, l := range Levels {
		if l == level {
			return Levels[i:]
		}
	}

	panic(fmt.Sprintf("level %s not support", level))
}

func zapLevel(level string) zapcore.Level {
	levelMap := map[string]zapcore.Level{
		"debug": zap.DebugLevel,
		"info":  zap.InfoLevel,
		"warn":  zap.WarnLevel,
		"error": zap.ErrorLevel,
	}

	return levelMap[level]
}

type ZapLogger struct {
	zapLogger []*zap.SugaredLogger
}

func NewLogger(opt *Option) Log {
	opt.apply()
	if _, err := os.Stat(opt.DirPath); os.IsNotExist(err) {
		if err = os.Mkdir(opt.DirPath, os.ModePerm); err != nil {
			panic(err)
		}
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	dirOpen(opt.DirPath)
	logger := &ZapLogger{}
	for _, level := range fileLevels(opt.Level) {
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(config),
			zapcore.AddSync(NewRotateFile(
				fmt.Sprintf("%v%v.log", opt.DirPath, level),
				WithRotateTime(parseDuration(opt.RotateDuration)),
				WithBackTime(parseDuration(opt.BackTime)))),
			zapLevel(level))

		logger.zapLogger = append(logger.zapLogger, zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2)).Sugar())
	}
	return logger
}

func parseDuration(s string) (duration time.Duration) {
	if s == "" {
		return
	}

	hour := 1
	if strings.Contains(s, "d") {
		hour = 24
		s = strings.Replace(s, "d", "h", -1)
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}

	return duration * time.Duration(hour)
}

func (wl *ZapLogger) Debug(msg string) {
	for _, l := range wl.zapLogger {
		l.Debugw(msg)
	}
}

func (wl *ZapLogger) Debugf(format string, v ...interface{}) {
	for _, l := range wl.zapLogger {
		l.Debugf(format, v...)
	}
}

func (wl *ZapLogger) Infof(format string, v ...interface{}) {
	for _, l := range wl.zapLogger {
		l.Infof(format, v...)
	}
}

func (wl *ZapLogger) Info(msg string) {
	for _, l := range wl.zapLogger {
		l.Infow(msg)
	}
}

func (wl *ZapLogger) Warnf(format string, v ...interface{}) {
	for _, l := range wl.zapLogger {
		l.Warnf(format, v...)
	}
}

func (wl *ZapLogger) Warn(msg string) {
	for _, l := range wl.zapLogger {
		l.Warnw(msg)
	}
}

func (wl *ZapLogger) Errorf(format string, v ...interface{}) {
	for _, l := range wl.zapLogger {
		l.Errorf(format, v...)
	}
}

func (wl *ZapLogger) Error(msg string) {
	for _, l := range wl.zapLogger {
		l.Errorw(msg)
	}
}

func dirOpen(path string) {
	if fileExist(path) {
		return
	}

	if err := os.MkdirAll(path, 766); err != nil {
		panic(err)
	}
}
