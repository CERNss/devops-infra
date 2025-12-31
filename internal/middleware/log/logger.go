package log

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

type noopLogger struct{}

func (noopLogger) Info(context.Context, string)  {}
func (noopLogger) Warn(context.Context, string)  {}
func (noopLogger) Error(context.Context, string) {}
func (noopLogger) Output(context.Context, string, string) {}

func NoopLogger() Logger {
	return noopLogger{}
}
