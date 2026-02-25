package executor

import (
	"context"

	"devops-infra/internal/interceptor"
	logmw "devops-infra/internal/middleware/log"
	tracemw "devops-infra/internal/middleware/trace"
)

type Runtime struct {
	Ctx       context.Context
	Trace     tracemw.TraceSink
	Output    logmw.OutputSinkFactory
	Logger    logmw.Logger
	LogDir    string
	NodeName  string
	NodeAddr  string
	Component string
}

func NewRuntime(ctx context.Context, trace tracemw.TraceSink) Runtime {
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
		rt.Trace = tracemw.DefaultTraceSink()
	}
	if rt.Output == nil {
		rt.Output = logmw.CombinedOutputSinkFactory{}
	}
	if rt.Logger == nil {
		rt.Logger = interceptor.DefaultLogger(rt.LogDir)
	}
	return rt
}

func WithNode(rt Runtime, name string, addr string) Runtime {
	rt.NodeName = name
	rt.NodeAddr = addr
	rt.Ctx = logmw.WithFields(rt.Ctx, map[string]any{
		"node":      name,
		"node_addr": addr,
	})
	return rt
}

func WithComponent(rt Runtime, component string) Runtime {
	rt.Component = component
	rt.Ctx = logmw.WithFields(rt.Ctx, map[string]any{
		"component": component,
	})
	return rt
}

func WithLogDir(rt Runtime, dir string) Runtime {
	rt.LogDir = dir
	return rt
}

func WithOutput(rt Runtime, output logmw.OutputSinkFactory) Runtime {
	rt.Output = output
	return rt
}

func WithLogger(rt Runtime, logger logmw.Logger) Runtime {
	rt.Logger = logger
	return rt
}
