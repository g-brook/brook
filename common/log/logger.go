/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"os"
	"strings"

	"github.com/brook/common/configs"
	"github.com/brook/common/hash"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerSetting struct {
	level   string
	logPath string
	outs    *hash.Set[string]
}

var sugar *zap.SugaredLogger

func defaultSetting() *LoggerSetting {
	return &LoggerSetting{
		level:   "info",
		logPath: "./logs/current_brook.log",
		outs:    hash.NewSet[string]("stdout", "file"),
	}
}

func NewLogger(config *configs.LoggerConfig) {
	setting := &LoggerSetting{
		outs: hash.NewSet[string](),
	}
	if config != nil {
		setting.level = config.LoggLevel
		setting.logPath = config.LogPath
		if config.Outs != "" {
			infos := strings.Split(config.Outs, ",")
			for _, v := range infos {
				setting.outs.Add(v)
			}
		}
	}
	initLogger(setting)
}

// InitLogger initializes the logging system with the specified log level
// It sets up both console and file logging with proper formatting and rotation
func initLogger(setting *LoggerSetting) {
	setting = newSetting(setting)
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

	// File output with rotation
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/current_brook.log",
		MaxSize:    100, // MB
		MaxAge:     30,  // days
		MaxBackups: 10,
		Compress:   true,
	})
	level := parseLevel(setting.level)
	var cores []zapcore.Core
	if setting.outs.Contains("stdout") {
		core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, core)
	}
	if setting.outs.Contains("file") {
		core := zapcore.NewCore(encoder, fileWriter, level)
		cores = append(cores, core)
	}
	// Multiple outputs: console + file
	core := zapcore.NewTee(
		cores...,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)) // zap.NewProduction()
	zap.ReplaceGlobals(logger)
	sugar = logger.Sugar()
}

func newSetting(setting *LoggerSetting) *LoggerSetting {
	def := defaultSetting()
	if setting == nil {
		setting = def
	}
	if setting.level == "" {
		setting.level = def.level
	}
	if setting.outs == nil || setting.outs.Len() == 0 {
		setting.outs = def.outs
	}
	return setting
}

// parseLevel converts a string level to zapcore.Level
// It maps common log level names to their corresponding zapcore values
// If the input is not recognized, it returns InfoLevel as default
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
// @Description: info log.
// @param msg The log message to print.
func Info(msg string, args ...any) {
	sugar.Infof(msg, args...)
}

// Debug
//
// @Description: debug log
// @param msg The log message
func Debug(msg string, args ...any) {
	sugar.Debugf(msg, args...)
}

// Warn
//
// @Description: warning log
// @param msg The log message
func Warn(msg string, args ...any) {
	sugar.Warnf(msg, args...)
}

// Error
//
// @Description: error log
// @param msg The log message
func Error(msg string, args ...any) {
	sugar.Errorf(msg, args...)
}

// Fatal
//
// @Description: fatal error log
// @param msg The log message
func Fatal(msg string, args ...any) {
	sugar.Fatalf(msg, args...)
}

// GnetLogger is a logger implementation for gnet library
// It forwards logs to our internal logging system
// This allows us to maintain consistent logging across different components
type GnetLogger struct {
}

// Debugf logs debug level messages.
// It forwards the log message to the internal debug logger.
func (g GnetLogger) Debugf(format string, args ...any) {
	Debug(format, args...)
}

// Infof logs info level messages.
// It forwards the log message to the internal info logger.
func (g GnetLogger) Infof(format string, args ...any) {
	Info(format, args...)
}

// Warnf logs warning level messages.
// It forwards the log message to the internal warn logger.
func (g GnetLogger) Warnf(format string, args ...any) {
	Warn(format, args...)
}

// Errorf logs error level messages.
// It forwards the log message to the internal error logger.
func (g GnetLogger) Errorf(format string, args ...any) {
	Error(format, args...)
}

// Fatalf logs fatal level messages and then calls os.Exit(1).
// It forwards the log message to the internal fatal logger.
func (g GnetLogger) Fatalf(format string, args ...any) {
	Fatal(format, args...)
}
