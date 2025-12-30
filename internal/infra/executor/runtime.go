package executor

import "context"

type Runtime struct {
	Ctx   context.Context
	Trace TraceSink
}

func NewRuntime(ctx context.Context, trace TraceSink) Runtime {
	return normalizeRuntime(Runtime{Ctx: ctx, Trace: trace})
}

func DefaultRuntime() Runtime {
	return normalizeRuntime(Runtime{})
}

func normalizeRuntime(rt Runtime) Runtime {
	if rt.Ctx == nil {
		rt.Ctx = context.Background()
	}
	if rt.Trace == nil {
		rt.Trace = NoopTraceSink()
	}
	return rt
}
