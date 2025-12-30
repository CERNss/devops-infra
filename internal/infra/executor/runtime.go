package executor

import "context"

type Runtime struct {
	Ctx      context.Context
	Trace    TraceSink
	LogDir   string
	NodeName string
	NodeAddr string
}

func NewRuntime(ctx context.Context, trace TraceSink) Runtime {
	return normalizeRuntime(Runtime{Ctx: ctx, Trace: trace})
}

func NormalizeRuntime(rt Runtime) Runtime {
	return normalizeRuntime(rt)
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

func WithNode(rt Runtime, name string, addr string) Runtime {
	rt.NodeName = name
	rt.NodeAddr = addr
	return rt
}

func WithLogDir(rt Runtime, dir string) Runtime {
	rt.LogDir = dir
	return rt
}
