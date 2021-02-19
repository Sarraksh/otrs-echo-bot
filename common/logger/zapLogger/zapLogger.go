package zapLogger

import (
	"fmt"
	"github.com/Sarraksh/otrs-echo-bot/common/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
)

type ZapLogger struct {
	Logger *zap.Logger // Initialized logger.
	Module string      // Name of module that uses logger.
}

// Return logger with debug level 10 MB file size and 5 log files preservation.
func NewDefault(logFilePath string) ZapLogger {
	zapLogger := NewZapSimpleLoggerWithRotation("debug", logFilePath, 10, 5)
	return ZapLogger{
		Logger: zapLogger,
	}
}

// Return simple logger with rotation.
// Take logging level, full path to log file, max size of log file in MB and number of backup files.
// Have no time limit for store log files
func NewZapSimpleLoggerWithRotation(logLevelStr string, logFilePath string, maxSize, maxBackups int) *zap.Logger {
	var logLevel zapcore.Level
	var isUnmarshalFail bool = false
	err := logLevel.UnmarshalText([]byte(logLevelStr))
	if err != nil {
		log.Printf("can't unmarshall log level '%v'. use 'error' log level instead", logLevelStr)
		isUnmarshalFail = true
		logLevel = zapcore.ErrorLevel
	}

	var cfg zap.Config
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006.01.02 15:04:05")
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
	})

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		writer,
		logLevel,
	)
	logger := zap.New(core)
	if isUnmarshalFail {
		logger.Error(fmt.Sprintf("can't unmarshall log level '%v'. use 'error' log level instead", logLevelStr))
	}

	return logger
}

// Add formatted module name to error string.
func formatString(module, message string) string {
	return fmt.Sprintf("[%25s] - '%s'", module, message)
}

// Set caller module name.
func (zl ZapLogger) SetModuleName(name string) logger.Logger {
	zl.Module = name
	return zl
}

func (zl ZapLogger) Error(message string) {
	formattedMessage := formatString(zl.Module, message)
	zl.Logger.Error(formattedMessage)
}

func (zl ZapLogger) Info(message string) {
	formattedMessage := formatString(zl.Module, message)
	zl.Logger.Info(formattedMessage)
}

func (zl ZapLogger) Debug(message string) {
	formattedMessage := formatString(zl.Module, message)
	zl.Logger.Debug(formattedMessage)
}
