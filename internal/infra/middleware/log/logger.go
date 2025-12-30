package log

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
}

type traceIDKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(), traceIDKey{}, traceID)
	}
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value(traceIDKey{}).(string)
	return traceID
}

func NewTraceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	}
	return hex.EncodeToString(buf)
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Info(ctx context.Context, msg string) {
	l.logger.Info(msg, traceField(ctx))
}

func (l *zapLogger) Warn(ctx context.Context, msg string) {
	l.logger.Warn(msg, traceField(ctx))
}

func (l *zapLogger) Error(ctx context.Context, msg string) {
	l.logger.Error(msg, traceField(ctx))
}

func traceField(ctx context.Context) zap.Field {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return zap.Skip()
	}
	return zap.String("trace_id", traceID)
}

type noopLogger struct{}

func (noopLogger) Info(context.Context, string)  {}
func (noopLogger) Warn(context.Context, string)  {}
func (noopLogger) Error(context.Context, string) {}

func NoopLogger() Logger {
	return noopLogger{}
}

func NewJSONLogger(path string) (Logger, error) {
	resolved, err := pathutil.ResolveUserPath(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "msg"
	encoderCfg.CallerKey = ""
	encoderCfg.StacktraceKey = ""
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	fileCore := zapcore.NewCore(encoder, zapcore.AddSync(f), zapcore.InfoLevel)
	stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.InfoLevel)
	logger := zap.New(zapcore.NewTee(fileCore, stderrCore))

	return &zapLogger{logger: logger}, nil
}

func NewStderrLogger() Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "msg"
	encoderCfg.CallerKey = ""
	encoderCfg.StacktraceKey = ""
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.InfoLevel)
	logger := zap.New(stderrCore)
	return &zapLogger{logger: logger}
}

func DefaultLogger(logDir string) Logger {
	path := constant.DefaultLogFile
	if logDir != "" {
		path = filepath.Join(logDir, "run.log")
	}
	logger, err := NewJSONLogger(path)
	if err != nil {
		return NewStderrLogger()
	}
	return logger
}
