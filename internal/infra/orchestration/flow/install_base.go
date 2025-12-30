package flow

import (
	"context"
	"fmt"

	"devops-infra/internal/infra/executor"
	logmw "devops-infra/internal/infra/middleware/log"
	tracemw "devops-infra/internal/infra/middleware/trace"
	"devops-infra/internal/infra/install/base"
	"devops-infra/internal/infra/install/base/cni"
	"devops-infra/internal/infra/install/base/containerd"
	"devops-infra/internal/infra/install/base/docker"
	"devops-infra/internal/infra/install/base/kernel"
	"devops-infra/internal/infra/install/base/mirror"
	"devops-infra/internal/infra/install/base/tools"
	"devops-infra/internal/infra/orchestration/execfactory"
	"devops-infra/internal/infra/orchestration/topology"
	"devops-infra/internal/infra/os"
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
	SkipKernel            bool
	SkipTools             bool
}

func InstallBase(ctx context.Context, opts InstallBaseOptions) error {
	// 1. Detect OS
	osInfo, err := os.Detect()
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
		logger = logmw.DefaultLogger(opts.LogDir)
	}
	runtime := executor.NewRuntime(ctx, trace)
	if opts.OutputFactory != nil {
		runtime = executor.WithOutput(runtime, opts.OutputFactory)
	}
	if opts.LogDir != "" {
		runtime = executor.WithLogDir(runtime, opts.LogDir)
	}
	runtime = executor.WithLogger(runtime, logger)

	// 2. Create executor
	exec, err := factory.Build(node, runtime)
	if err != nil {
		return err
	}

	// 3. Create OS driver
	driver, err := os.NewDriver(osInfo, exec)
	if err != nil {
		return err
	}

	mode := opts.DockerInstallMode
	if mode == "" {
		mode = docker.InstallModeOfficial
	}

	// 4. Build base base
	components := []base.Component{}
	if !opts.SkipKernel {
		components = append(components, kernel.New(driver))
	}
	components = append(components, mirror.New(driver, mirror.Options{
		Enable: opts.EnableMirror,
		Source: opts.LinuxMirrorSource,
	}))
	if !opts.SkipTools {
		components = append(components, tools.New(driver))
	}
	components = append(components, cni.New(driver, cni.Options{
		Arch: opts.ContainerdArch,
	}))

	containerdInstaller := containerd.New(driver, containerd.Options{
		Version:          opts.ContainerdVersion,
		Arch:             opts.ContainerdArch,
		Checksum:         opts.ContainerdChecksum,
		EnsureCNIConfig:  true,
	})

	if mode == docker.InstallModeNerdctl {
		components = append(
			components,
			docker.New(driver, docker.Options{
				Mode:            mode,
				Source:          opts.DockerMirrorSource,
				RegistryMirrors: opts.DockerRegistryMirrors,
				EngineVersion:   opts.DockerEngineVersion,
			}),
			containerdInstaller,
		)
	} else {
		components = append(
			components,
			docker.New(driver, docker.Options{
				Mode:            mode,
				Source:          opts.DockerMirrorSource,
				RegistryMirrors: opts.DockerRegistryMirrors,
				EngineVersion:   opts.DockerEngineVersion,
			}),
			containerdInstaller,
		)
	}

	installer := base.New(components...).WithLogger(logger)

	// 5. Run
	logger.Info(ctx, "install-base: start")
	if err := installer.Install(ctx); err != nil {
		logger.Error(ctx, fmt.Sprintf("install-base: failed: %v", err))
		return err
	}
	logger.Info(ctx, "install-base: done")
	return nil
}
