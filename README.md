# devops-infra

## 命令与操作

### 全局参数
- `--sudo`：默认开启，以 sudo 执行命令。
- `--dry-run`：仅打印将要执行的命令，不实际执行。
- `--verbose`：输出更详细的执行信息。

### install 命令
- `infra install`：安装相关命令集合（目前仅实现 `base`）。
- `infra install base`：安装基础环境（kernel/sysctl/cgroup、基础工具、docker、containerd）。
  - `--mirror`：切换系统软件源，并配合镜像脚本处理 Docker 源。
  - `--docker-install-mode=mirror|nerdctl`：
    - `mirror`：通过镜像脚本安装 Docker 并启动服务。
    - `nerdctl`：要求已安装 nerdctl，创建 `/usr/bin/docker` 软链接。

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
