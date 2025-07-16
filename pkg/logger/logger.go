package logger

import (
	"go.uber.org/zap"
)

const (
	EnvProd = "production"
	EnvDev  = "development"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
}

type zapLogger struct {
	logger *zap.Logger
}

func New(prod string) Logger {
	var logger *zap.Logger
	var err error

	switch prod {
	case EnvProd:
		logger, err = zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1))
	case EnvDev:
		logger, err = zap.NewDevelopment(zap.AddCaller(), zap.AddCallerSkip(1))
	default:
		logger, err = zap.NewDevelopment(zap.AddCaller(), zap.AddCallerSkip(1)) // Default to development if not specified
	}

	if err != nil {
		panic("cannot initialize zap logger: " + err.Error())
	}

	return &zapLogger{logger: logger}
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}
