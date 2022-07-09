package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLog = &sync.Pool{
		New: func() any {
			return new(zapcore.InfoLevel)
		},
	}
	fields = map[string]zapcore.Field{}
	mu     sync.RWMutex
)

// zapLogger is a standard logger using zap
type zapLogger struct {
	logLevel zapcore.Level
	*zap.SugaredLogger
}

// new creates Zap Logger constructor
func new(level zapcore.Level) *zapLogger {
	zl := &zapLogger{logLevel: level}
	zl.initLogger()
	return zl
}

// initLogger Init logger
func (l *zapLogger) initLogger() {
	// set log level
	logLevel := l.logLevel
	var logWriter zapcore.WriteSyncer
	var encoderCfg zapcore.EncoderConfig
	logWriter = zapcore.AddSync(os.Stderr)

	var encoder zapcore.Encoder

	if l.logLevel == zapcore.DebugLevel {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	encoderCfg.LevelKey = "lvl"
	encoderCfg.CallerKey = "caller"
	encoderCfg.TimeKey = "time"
	encoderCfg.NameKey = "name"
	encoderCfg.MessageKey = "msg"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	if l.logLevel == zapcore.DebugLevel {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	mu.RLock()
	zfields := make([]zapcore.Field, len(fields))
	i := 0
	for _, v := range fields {
		zfields[i] = v
		i++
	}
	mu.RUnlock()

	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zfields...),
	)

	l.SugaredLogger = logger.Sugar()
}

// WithFields methods to satisfy Logger interface
func WithFields(args ...zapcore.Field) {
	mu.Lock()
	for _, a := range args {
		fields[a.Key] = a
	}
	mu.Unlock()
}

func Debug(args ...any) {
	zl := globalLog.Get().(*zapLogger)
	zl.Debug(args...)
}

func Debugf(template string, args ...any) {
	zl := globalLog.Get().(*zapLogger)
	zl.Debugf(template, args...)
}

func Fatal(args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Fatalf(template, args...)
}

func Info(args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Info(args...)
}

func Infof(template string, args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Infof(template, args...)
}

func Print(args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Info(args...)
}

func Println(args ...any) {
	zl := globalLog.Get().(*zapLogger)
	zl.Info(args...)
}

func Printf(template string, args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Infof(template, args...)
}

func Warn(args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Warnf(template, args...)
}

func Error(args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	zl := globalLog.Get().(*zapLogger)
	zl.Errorf(template, args...)
}
