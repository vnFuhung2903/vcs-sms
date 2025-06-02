package logger

import (
	"os"

	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func Load(env *env.LoggerEvn) error {
	level, err := zapcore.ParseLevel(env.Level)
	if err != nil {
		return err
	}

	if err := os.MkdirAll("./logs", 0755); err != nil {
		return err
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   env.FilePath,
		MaxSize:    env.MaxSize,
		MaxAge:     env.MaxAge,
		MaxBackups: env.MaxBackups,
		Compress:   true,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(lumberjackLogger),
		level,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	core := zapcore.NewTee(fileCore, consoleCore)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func Sync() error {
	return Log.Sync()
}

func WithContext(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

func Error(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	Log.Error(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Fatal(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	Log.Fatal(msg, fields...)
}
