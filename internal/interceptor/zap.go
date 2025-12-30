package interceptor

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"devops-infra/internal/constant"
	logmw "devops-infra/internal/middleware/log"
	pathutil "devops-infra/internal/utils/path"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

func (l *zapLogger) Output(ctx context.Context, stream string, line string) {
	l.logger.Info("cmd_output", traceField(ctx), zap.String("stream", stream), zap.String("output", line))
}

func traceField(ctx context.Context) zap.Field {
	traceID := logmw.TraceIDFromContext(ctx)
	if traceID == "" {
		return zap.Skip()
	}
	return zap.String("trace_id", traceID)
}

func NewJSONLogger(path string) (logmw.Logger, error) {
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

func NewStderrLogger() logmw.Logger {
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

func DefaultLogger(logDir string) logmw.Logger {
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
