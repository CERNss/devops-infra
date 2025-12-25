# devops-infra

## 命令与操作

### 全局参数
- `--sudo`：默认开启，以 sudo 执行命令。
- `--dry-run`：仅打印将要执行的命令，不实际执行。
- `--verbose`：输出更详细的执行信息。

### install 命令
- `infra install`：安装相关命令集合（目前仅实现 `base`）。
- `infra install base`：安装基础环境（kernel/sysctl/cgroup、基础工具、docker、containerd）。
  - `--mirror`：切换系统软件源。
  - `--docker-install-mode=docker|nerdctl`：
    - `docker`：通过镜像脚本安装官方 Docker 并启动服务。
    - `nerdctl`：自动安装 nerdctl/runc/cni，创建 `/usr/bin/docker` 软链接。
  - `--docker-registry-mirror`：配置 Docker registry 镜像（可多次传入或逗号分隔）。
  - `--containerd-version`：指定 containerd 版本（默认 1.7.28）。
  - `--containerd-arch`：指定 containerd 架构（默认 amd64）。
  - `--containerd-checksum`：指定 containerd tarball 的 sha256 校验值。
  - `--skip-kernel`：跳过 kernel/sysctl 配置。
  - `--skip-tools`：跳过基础工具安装。

### 预留命令（尚未实现）
- `infra install k8s`：预留的 Kubernetes 安装入口。
- `infra install k3s`：预留的 K3s 安装入口。
- `infra install k3d`：预留的 K3d 安装入口。
- `infra status`：预留的状态检查入口。
- `infra doctor`：预留的环境诊断入口。
- `infra uninstall`：预留的卸载入口。

### 示例
- `infra install base`
- `infra install base --mirror --dry-run`
- `infra install base --docker-install-mode=nerdctl`
- `infra install base --docker-registry-mirror=https://docker.1ms.run,https://dockerproxy.net`
- `infra install base --containerd-version=1.7.28 --containerd-checksum=<sha256>`
- `infra install base --skip-kernel --skip-tools`

## 架构与流程
命令流程示例：`infra install base --mirror --dry-run`
↓
cmd/install_base.go
↓
orchestration.InstallBase(ctx, options)
↓
DetectOS
NewLocalExecutor(execOpts)
NewOSDriver(osInfo, exec)
↓
base.New(...).Install()


┌──────────────────────────────────────┐
│            CLI Layer (cobra)         │
│ infra install base / k8s / k3s / k3d │
│  infra status / doctor / uninstall   │
└──────────────────▲───────────────────┘

┌──────────────────┴───────────────────┐
│          Orchestration Layer         │
│  - Install Flow                      │
│  - Dependency Check                  │
│  - Order & Idempotency               │
└──────────────────▲───────────────────┘

┌──────────────────┴───────────────────┐
│             Domain Layer             │
│  Base Layer        Platform Layer    │
│  docker            k8s               │
│  containerd        k3s               │
│  kernel            k3d               │
└──────────────────▲───────────────────┘

┌──────────────────┴───────────────────┐
│            OS Driver Layer           │
│  debian-family / rhel-family         │
│  apt / yum / systemd / sysctl        │
└──────────────────────────────────────┘


┌────────────────────────────┐
│        Platform Layer      │
│       k8s / k3s / k3d      │
│                            │
└──────────────▲─────────────┘
               │
┌──────────────┴─────────────┐
│        Base Layer          │
│    docker / containerd     │
│  kernel / sysctl / cgroup  │
│                            │
└──────────────▲─────────────┘
               │
┌──────────────┴─────────────┐
│       OS Driver Layer      │
│  ubuntu / debian / rhel    │
│                            │
└────────────────────────────┘
