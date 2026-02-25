package log

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"time"
)

type Logger interface {
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
	Output(ctx context.Context, stream string, line string)
}

type traceIDKey struct{}
type fieldsKey struct{}

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

func WithFields(ctx context.Context, fields map[string]any) context.Context {
	if len(fields) == 0 {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	merged := map[string]any{}
	if existing, ok := ctx.Value(fieldsKey{}).(map[string]any); ok {
		maps.Copy(merged, existing)
	}
	maps.Copy(merged, fields)
	return context.WithValue(ctx, fieldsKey{}, merged)
}

func FieldsFromContext(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	fields, _ := ctx.Value(fieldsKey{}).(map[string]any)
	if len(fields) == 0 {
		return nil
	}
	copied := map[string]any{}
	maps.Copy(copied, fields)
	return copied
}

func NewTraceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	}
	return hex.EncodeToString(buf)
}

type noopLogger struct{}

func (noopLogger) Info(context.Context, string)           {}
func (noopLogger) Warn(context.Context, string)           {}
func (noopLogger) Error(context.Context, string)          {}
func (noopLogger) Output(context.Context, string, string) {}

func NoopLogger() Logger {
	return noopLogger{}
}
