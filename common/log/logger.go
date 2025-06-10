package log

import (
	"common/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var sugar *zap.SugaredLogger

func InitFunc(config configs.LoggerConfig) {
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// 文件输出，带滚动
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/current_brook.log",
		MaxSize:    100, // MB
		MaxAge:     30,  // days
		MaxBackups: 10,
		Compress:   true,
	})
	level := parseLevel(config.LoggLevel)
	// 多输出：控制台 + 文件
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(encoder, fileWriter, level),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)) // zap.NewProduction()
	zap.ReplaceGlobals(logger)
	sugar = logger.Sugar()
}
func parseLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Info
//
//	@Description: info日志.
//	@param msg 打印的日志信息.
func Info(msg string, args ...any) {
	sugar.Infof(msg, args...)
}

// Debug
//
//	@Description: debug
//	@param msg 打印的日志信息
func Debug(msg string, args ...any) {
	sugar.Debugf(msg, args...)
}

// Warn
//
//	@Description:
//	@param msg 打印的日志信息
func Warn(msg string, args ...any) {
	sugar.Warnf(msg, args...)
}

// Error
//
//	@Description:
//	@param msg 打印的日志信息
func Error(msg string, args ...any) {
	sugar.Errorf(msg, args...)
}

// Fatal
//
//	@Description:
//	@param msg 打印的日志信息
func Fatal(msg string, args ...any) {
	sugar.Fatalf(msg, args...)
}

type GnetLogger struct {
}

func (g GnetLogger) Debugf(format string, args ...any) {
	Debug(format, args...)
}

func (g GnetLogger) Infof(format string, args ...any) {
	Info(format, args...)
}

func (g GnetLogger) Warnf(format string, args ...any) {
	Warn(format, args...)
}

func (g GnetLogger) Errorf(format string, args ...any) {
	Error(format, args...)
}

func (g GnetLogger) Fatalf(format string, args ...any) {
	Fatal(format, args...)
}
