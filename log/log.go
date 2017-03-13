package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields

type Level uint8

var DebugLevel = Level(logrus.DebugLevel)
var InfoLevel = Level(logrus.InfoLevel)
var WarnLevel = Level(logrus.WarnLevel)
var ErrorLevel = Level(logrus.ErrorLevel)

var logger = logrus.New()

func SetOutput(w io.Writer) {
	logger.Out = w
}

func WithField(key string, value interface{}) *logrus.Entry {
	return logger.WithField(key, value)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Debugln(format string) {
	logger.Debugln(format)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Infoln(format string) {
	logger.Infoln(format)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Warnln(format string) {
	logger.Warnln(format)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Errorln(format string) {
	logger.Errorln(format)
}

func Print(args ...interface{}) {
	logger.Print(args...)
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func Println(args ...interface{}) {
	logger.Println(args...)
}

func WithFields(fields Fields) *logrus.Entry {
	return logger.WithFields(logrus.Fields(fields))
}

func WithError(err error) *logrus.Entry {
	return logger.WithField("error", err)
}

func SetLevel(level Level) {
	logger.Level = (logrus.Level(level))
}

func Fatalln(format string) {
	logger.Fatalln(format)
}
