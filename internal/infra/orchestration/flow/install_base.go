package flow

import (
	"context"
	"fmt"
	"io"
	"os"

	"devops-infra/internal/infra/executor"
	"devops-infra/internal/infra/install/base"
	"devops-infra/internal/infra/install/base/cni"
	"devops-infra/internal/infra/install/base/containerd"
	"devops-infra/internal/infra/install/base/docker"
	"devops-infra/internal/infra/install/base/kernel"
	"devops-infra/internal/infra/install/base/mirror"
	"devops-infra/internal/infra/install/base/tools"
	"devops-infra/internal/infra/install/base/validate"
	"devops-infra/internal/infra/orchestration/execfactory"
	"devops-infra/internal/infra/orchestration/topology"
	osinfra "devops-infra/internal/infra/os"
	"devops-infra/internal/interceptor"
	logmw "devops-infra/internal/middleware/log"
	tracemw "devops-infra/internal/middleware/trace"
)

type InstallBaseOptions struct {
	ExecOpts              executor.Options
	Topology              *topology.Topology
	ExecutorFactory       execfactory.ExecutorFactory
	TraceSink             tracemw.TraceSink
	OutputFactory         logmw.OutputSinkFactory
	Logger                logmw.Logger
	LogDir                string
	EnableMirror          bool
	LinuxMirrorSource     string
	DockerInstallMode     docker.InstallMode
	DockerMirrorSource    string
	DockerRegistryMirrors []string
	DockerEngineVersion   string
	ContainerdVersion     string
	ContainerdArch        string
	ContainerdChecksum    string
	CNISubnet             string
	CNIRouteDst           string
	SkipKernel            bool
	SkipTools             bool
}

type failureAggregator interface {
	Close() error
	HasFailures() bool
	Summary() tracemw.FailureSummary
}

func finalizeFailureAggregator(aggregator failureAggregator, out io.Writer) {
	if aggregator == nil {
		return
	}
	_ = aggregator.Close()
	if !aggregator.HasFailures() {
		return
	}
	if out == nil {
		out = io.Discard
	}
	fmt.Fprint(out, tracemw.FormatFailureSummary(aggregator.Summary()))
}

func InstallBase(ctx context.Context, opts InstallBaseOptions) error {
	// 1. Detect OS
	osInfo, err := osinfra.Detect()
	if err != nil {
		return err
	}

	topo := opts.Topology
	if topo == nil {
		defaultTopology := topology.NewSingleNodeTopology(opts.ExecOpts)
		topo = &defaultTopology
	}
	if len(topo.Nodes) != 1 {
		return fmt.Errorf("only single-node topology is supported for now")
	}

	node := topo.Nodes[0]
	if node.ExecOpts == nil {
		node.ExecOpts = &opts.ExecOpts
	}

	factory := opts.ExecutorFactory
	if factory == nil {
		factory = execfactory.DefaultExecutorFactory{}
	}

	trace := opts.TraceSink
	if trace == nil {
		trace = tracemw.DefaultTraceSink()
	}
	logger := opts.Logger
	if logger == nil {
		logger = interceptor.DefaultLogger(opts.LogDir)
	}
	aggregator, aggErr := tracemw.NewFailureAggregator(opts.LogDir, "install-base")
	if aggErr != nil {
		logger.Warn(ctx, fmt.Sprintf("failed to initialize failure aggregator: %v", aggErr))
	} else {
		defer finalizeFailureAggregator(aggregator, os.Stdout)
		trace = tracemw.NewMultiTraceSink(trace, aggregator)
	}

	runtime := executor.NewRuntime(ctx, trace)
	if opts.OutputFactory != nil {
		runtime = executor.WithOutput(runtime, opts.OutputFactory)
	}
	if opts.LogDir != "" {
		runtime = executor.WithLogDir(runtime, opts.LogDir)
	}
	runtime = executor.WithLogger(runtime, logger)

	mode := opts.DockerInstallMode
	if mode == "" {
		mode = docker.InstallModeOfficial
	}

	newDriver := func(component string) (osinfra.Driver, error) {
		componentRuntime := executor.WithComponent(runtime, component)
		exec, buildErr := factory.Build(node, componentRuntime)
		if buildErr != nil {
			return nil, buildErr
		}
		return osinfra.NewDriver(osInfo, exec)
	}

	// 2. Build base components
	components := []base.Component{}
	if !opts.SkipKernel {
		driver, driverErr := newDriver("kernel")
		if driverErr != nil {
			return driverErr
		}
		components = append(components, kernel.New(driver))
	}

	driverMirror, err := newDriver("system-mirror")
	if err != nil {
		return err
	}
	components = append(components, mirror.New(driverMirror, mirror.Options{
		Enable: opts.EnableMirror,
		Source: opts.LinuxMirrorSource,
	}))

	if !opts.SkipTools {
		driver, driverErr := newDriver("common-tools")
		if driverErr != nil {
			return driverErr
		}
		components = append(components, tools.New(driver))
	}

	driverCNI, err := newDriver("cni-plugins")
	if err != nil {
		return err
	}
	components = append(components, cni.New(driverCNI, cni.Options{
		Arch: opts.ContainerdArch,
	}))

	driverContainerd, err := newDriver("containerd")
	if err != nil {
		return err
	}
	containerdInstaller := containerd.New(driverContainerd, containerd.Options{
		Version:         opts.ContainerdVersion,
		Arch:            opts.ContainerdArch,
		Checksum:        opts.ContainerdChecksum,
		EnsureCNIConfig: true,
		CNISubnet:       opts.CNISubnet,
		CNIRouteDst:     opts.CNIRouteDst,
	})

	driverDocker, err := newDriver("docker")
	if err != nil {
		return err
	}
	dockerInstaller := docker.New(driverDocker, docker.Options{
		Mode:            mode,
		Source:          opts.DockerMirrorSource,
		RegistryMirrors: opts.DockerRegistryMirrors,
		EngineVersion:   opts.DockerEngineVersion,
	})

	if mode == docker.InstallModeNerdctl {
		components = append(
			components,
			containerdInstaller,
			dockerInstaller,
		)
	} else {
		components = append(
			components,
			dockerInstaller,
			containerdInstaller,
		)
	}

	driverValidate, err := newDriver("base-postflight")
	if err != nil {
		return err
	}
	components = append(components, validate.New(driverValidate, validate.Options{Mode: mode}))

	installer := base.New(components...).WithLogger(logger)

	// 3. Run
	logger.Info(ctx, "install-base: start")
	if err := installer.Install(ctx); err != nil {
		logger.Error(ctx, fmt.Sprintf("install-base: failed: %v", err))
		return err
	}
	logger.Info(ctx, "install-base: done")
	return nil
}
