# devops-infra

## 命令与操作

### 全局参数
- `--sudo`：默认开启，以 sudo 执行命令。
- `--dry-run`：仅打印将要执行的命令，不实际执行。
- `--verbose`：输出更详细的执行信息。

### install 命令
- `devops-infra install`：安装相关命令集合（目前仅实现 `base`）。
- `devops-infra install base`：安装基础环境（kernel/sysctl/cgroup、基础工具、docker、containerd）。
  - `--mirror`：切换系统软件源。
  - `--mirror-source`：指定系统镜像源（域名或别名，传入后自动启用 `--mirror`，支持 `国内-阿里云`、`教育网-清华`、`海外-xtom` 这样的分类写法）。
  - `--docker-install-mode=docker|nerdctl`：
    - `docker`：通过镜像脚本安装官方 Docker 并启动服务。
    - `nerdctl`：自动安装 nerdctl/runc/cni，创建 `/usr/bin/docker` 软链接。
  - `--docker-source`：指定 Docker CE 镜像源（域名或别名，支持 `国内-阿里云`、`教育网-清华`、`海外-docker`）。
  - `--docker-registry-mirror`：配置 Docker registry 镜像（可多次传入或逗号分隔，仅 docker 模式生效，支持 `国内-1ms`、`海外-dockerhub`）。
  - `--containerd-version`：指定 containerd 版本（默认 1.7.28）。
  - `--containerd-arch`：指定 containerd 架构（默认 amd64）。
  - `--containerd-checksum`：指定 containerd tarball 的 sha256 校验值（可选）。
  - `--skip-kernel`：跳过 kernel/sysctl 配置。
  - `--skip-tools`：跳过基础工具安装。

### 预留命令（尚未实现）
- `devops-infra install k8s`：预留的 Kubernetes 安装入口。
- `devops-infra install k3s`：预留的 K3s 安装入口。
- `devops-infra install k3d`：预留的 K3d 安装入口。
- `devops-infra status`：预留的状态检查入口。
- `devops-infra doctor`：预留的环境诊断入口。
- `devops-infra uninstall`：预留的卸载入口。

### 示例
- `devops-infra install base`
- `devops-infra install base --mirror --dry-run`
- `devops-infra install base --mirror-source=阿里云`
- `devops-infra install base --docker-install-mode=nerdctl`
- `devops-infra install base --docker-source=国内-阿里云 --docker-registry-mirror=国内-1ms,国内-dockerproxy`
- `devops-infra install base --containerd-version=1.7.28 --containerd-arch=arm64 --containerd-checksum=<sha256>`
- `devops-infra install base --skip-kernel --skip-tools`

### 镜像分类与别名
- 分类前缀：`国内`/`默认`/`大陆`/`cn`/`default`，`教育网`/`教育`/`校园`/`edu`，`海外`/`境外`/`abroad`/`overseas`。
- 常用系统镜像别名：`阿里云`/`aliyun`，`腾讯云`/`tencent`，`华为云`/`huawei`，`网易`/`163`，`天翼云`/`ctyun`，`清华`/`tuna`，`北大`/`pku`，`浙大`/`zju`，`南大`/`nju`，`交大`/`sjtu`，`中科大`/`ustc`，`中科院`/`iscas`，`火山`/`volc`。
- 常用 Docker CE 镜像别名：`阿里云`/`aliyun`，`腾讯云`/`tencent`，`华为云`/`huawei`，`网易`/`163`，`清华`/`tuna`，`北大`/`pku`，`浙大`/`zju`，`南大`/`nju`，`交大`/`sjtu`，`中科大`/`ustc`，`中科院`/`iscas`，`azure`，`docker`。
- 常用 Docker Registry 别名：`1ms`，`dockerproxy`，`daocloud`，`1panel`，`阿里云`/`aliyun`，`腾讯云`/`tencent`，`dockerhub`。

## 架构与流程
命令流程示例：`devops-infra install base --mirror --dry-run`
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
│ devops-infra install base / k8s / k3s / k3d │
│    devops-infra status / doctor / uninstall │
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
