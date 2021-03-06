package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLog = &sync.Pool{}
	fields    = map[string]zapcore.Field{}
	mu        sync.RWMutex
)

// zapLogger is a standard logger using zap
type zapLogger struct {
	*zap.SugaredLogger
	logLevel zapcore.Level
}

func (l *zapLogger) Warningf(s string, i ...interface{}) {
	l.Warnf(s, i...)
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
	logWriter = zapcore.AddSync(os.Stdout)

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
	zl := new(zapcore.InfoLevel)
	zl.initLogger()
	globalLog.Put(zl)
}

func New() *zapLogger {
	zl, ok := globalLog.Get().(*zapLogger)
	if !ok {
		zl = new(zapcore.InfoLevel)
		zl.initLogger()
		globalLog.Put(zl)
	}
	return zl
}

func Debug(args ...any) {
	zl := New()
	zl.Debug(args...)
}

func Debugf(template string, args ...any) {
	zl := New()
	zl.Debugf(template, args...)
}

func Fatal(args ...interface{}) {
	zl := New()
	zl.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	zl := New()
	zl.Fatalf(template, args...)
}

func Info(args ...interface{}) {
	zl := New()
	zl.Info(args...)
}

func Infof(template string, args ...interface{}) {
	zl := New()
	zl.Infof(template, args...)
}

func Print(args ...interface{}) {
	zl := New()
	zl.Info(args...)
}

func Println(args ...any) {
	zl := New()
	zl.Info(args...)
}

func Printf(template string, args ...interface{}) {
	zl := New()
	zl.Infof(template, args...)
}

func Warn(args ...interface{}) {
	zl := New()
	zl.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	zl := New()
	zl.Warnf(template, args...)
}

func Error(args ...interface{}) {
	zl := New()
	zl.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	zl := New()
	zl.Errorf(template, args...)
}
