package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Nurlan270/cloud-storage-go/internal/core/config"
)

type Logger struct {
	*zap.Logger

	file *os.File
}

var (
	l    = &Logger{}
	once = sync.Once{}
)

func Init(conf config.Config) {
	once.Do(func() {
		env := conf.GetAppEnv()

		encoder := getEncoder(env)

		ws := getWriteSyncer(env, conf.GetAppName())

		enabler := getLevelEnabler(env)

		core := zapcore.NewCore(encoder, ws, enabler)

		zapLogger := zap.New(core, zap.AddCaller())

		l.Logger = zapLogger
	})
}

// Get returns global logger's instance.
// It can be insecure to call Get if logger
// was not yet initialized using Init.
func Get() *Logger {
	return l
}

func (l *Logger) Close() error {
	if err := l.file.Close(); err != nil {
		return err
	}

	return nil
}

func getEncoder(env string) zapcore.Encoder {
	encoderCfg := buildEncoderConfig(env)

	var encoder zapcore.Encoder

	switch env {
	case "local":
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	default:
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	return encoder
}

func getWriteSyncer(env, appName string) zapcore.WriteSyncer {
	var ws zapcore.WriteSyncer

	switch env {
	case "local":
		ws = zapcore.AddSync(os.Stdout)
	default:
		ws = zapcore.Lock(buildFileWriteSyncer(appName))
	}

	return ws
}

func getLevelEnabler(env string) zapcore.LevelEnabler {
	var enabler zapcore.LevelEnabler

	switch env {
	case "local":
		enabler = zapcore.DebugLevel
	default:
		enabler = zapcore.InfoLevel
	}

	return enabler
}

func buildEncoderConfig(env string) zapcore.EncoderConfig {
	var levelEncoder zapcore.LevelEncoder

	switch env {
	case "local":
		levelEncoder = zapcore.CapitalColorLevelEncoder
	default:
		levelEncoder = zapcore.CapitalLevelEncoder
	}

	return zapcore.EncoderConfig{
		NameKey:        "logger",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "message",
		TimeKey:        "timestamp",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeName:     zapcore.FullNameEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    levelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
}

func buildFileWriteSyncer(appName string) zapcore.WriteSyncer {
	logFilePath := filepath.Join(
		"logs/"+appName, fmt.Sprintf("%s.log", time.Now().Format("2006/01/02/15-04")),
	)

	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		panic(fmt.Sprintf("logger: failed to create log file: %s", err))
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("logger: failed to open log file: %s", err))
	}

	l.file = file

	return zapcore.AddSync(file)
}
