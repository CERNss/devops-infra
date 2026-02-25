package interceptor

import (
	"context"
	"os"
	"path/filepath"
	"sort"
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
	l.logger.Info(msg, contextFields(ctx)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string) {
	l.logger.Warn(msg, contextFields(ctx)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string) {
	l.logger.Error(msg, contextFields(ctx)...)
}

func (l *zapLogger) Output(ctx context.Context, stream string, line string) {
	fields := contextFields(ctx)
	fields = append(fields, zap.String("stream", stream), zap.String("output", line))
	l.logger.Info("cmd_output", fields...)
}

func contextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 4)
	traceID := logmw.TraceIDFromContext(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	contextValues := logmw.FieldsFromContext(ctx)
	if len(contextValues) == 0 {
		return fields
	}

	keys := make([]string, 0, len(contextValues))
	for key := range contextValues {
		if key == "trace_id" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := contextValues[key]
		switch typed := value.(type) {
		case string:
			fields = append(fields, zap.String(key, typed))
		case bool:
			fields = append(fields, zap.Bool(key, typed))
		case int:
			fields = append(fields, zap.Int(key, typed))
		case int32:
			fields = append(fields, zap.Int32(key, typed))
		case int64:
			fields = append(fields, zap.Int64(key, typed))
		case uint:
			fields = append(fields, zap.Uint(key, typed))
		case uint32:
			fields = append(fields, zap.Uint32(key, typed))
		case uint64:
			fields = append(fields, zap.Uint64(key, typed))
		case float64:
			fields = append(fields, zap.Float64(key, typed))
		case time.Duration:
			fields = append(fields, zap.Duration(key, typed))
		default:
			fields = append(fields, zap.Any(key, value))
		}
	}
	return fields
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
	logger := zap.New(fileCore)

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
		return logmw.NoopLogger()
	}
	return logger
}
