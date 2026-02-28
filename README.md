# devops-infra

## 编译

- 本机编译（当前平台）：
  - `go build -o ./build/bin/devops-infra .`
- Linux amd64 编译（发布版）：
  - `mkdir -p build/bin`
  - `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./build/bin/devops-infra-linux-amd64 .`

## 命令与操作

### 全局参数
- `--sudo`：默认开启，以 sudo 执行命令。
- `--dry-run`：仅打印将要执行的命令，不实际执行。
- `--verbose`：输出更详细的执行信息。
- `--timeout`：全局超时时间（默认 5m，传 0 关闭）。

### install 命令
- `devops-infra install`：安装相关命令集合（目前仅实现 `base`）。
- `devops-infra install base`：安装基础环境（kernel/sysctl/cgroup、基础工具、docker、containerd）。
  - `--mirror`：切换系统软件源。
  - `--mirror-source`：指定系统镜像源（域名或别名，传入后自动启用 `--mirror`，支持 `国内-阿里云`、`教育网-清华`、`海外-xtom` 这样的分类写法）。
  - `--docker-install-mode=docker|nerdctl`：
    - `docker`：通过镜像脚本安装官方 Docker 并启动服务。
    - `nerdctl`：自动安装 nerdctl/runc/cni，创建 `/usr/bin/docker` 软链接；当 `/etc/cni/net.d` 为空时自动生成 `99-nerdctl-bridge.conflist` 最小 CNI 配置。
  - `--docker-source`：指定 Docker CE 镜像源（域名或别名，仅 docker 模式生效，支持 `国内-阿里云`、`教育网-清华`、`海外-docker`）。
  - `--docker-registry-mirror`：配置 Docker registry 镜像（可多次传入或逗号分隔，仅 docker 模式生效，支持 `国内-1ms`、`海外-dockerhub`）。
  - `--docker-version`：指定 Docker Engine 版本（默认 27.5.1，仅 docker 模式生效）。
  - `--containerd-version`：指定 containerd 版本（默认 2.1.0）。
  - `--containerd-arch`：指定 containerd 架构（默认 amd64）。
  - `--containerd-checksum`：指定 containerd tarball 的 sha256 校验值（可选）。
  - `--cni-subnet`：nerdctl 模式下的 CNI 子网（默认 10.88.0.0/16）。
  - `--cni-route-dst`：nerdctl 模式下的 CNI 路由目标（默认 0.0.0.0/0）。
  - `--log-dir`：日志目录（默认 logs）。
  - `--enable-log`：是否启用运行日志与命令输出日志（默认 true）。
  - `--trace-dir`：trace 目录（默认 trace/trace.jsonl）。
  - `--enable-trace`：是否启用 trace 事件（默认 true）。
  - `--skip-kernel`：跳过 kernel/sysctl 配置。
  - `--skip-tools`：跳过基础工具安装。

### 预留命令（尚未实现）
- `devops-infra install k3s`：预留的 K3s 安装入口。
- `devops-infra install k3d`：预留的 K3d 安装入口。
- `devops-infra status`：预留的状态检查入口。
- `devops-infra doctor`：预留的环境诊断入口。
- `devops-infra uninstall`：预留的卸载入口。

### install k8s 命令
- `devops-infra install k8s`：安装 Kubernetes（kubeadm）。
  - `--kubernetes-version`：kubeadm init 版本（默认 1.28.15）。
  - `--cri-socket`：CRI socket 路径（默认 containerd）。
  - `--control-plane-endpoint`：控制面入口。
  - `--apiserver-advertise-address`：API Server 宣告地址。
  - `--pod-network-cidr`：Pod 网段（默认 10.244.0.0/16）。
  - `--service-cidr`：Service 网段（默认 10.96.0.0/12）。
  - `--service-dns-domain`：Service DNS 域（默认 cluster.local）。
  - `--image-repository`：镜像仓库（默认 registry.k8s.io）。
  - `--token`/`--token-ttl`：引导 token 与有效期。
  - `--upload-certs`/`--certificate-key`：控制面证书上传配置。
  - `--ignore-preflight-errors`：忽略指定预检。
  - `--feature-gates`：功能开关。
  - `--patches`：kubeadm patches 目录。
  - `--config`：kubeadm 配置文件（使用该参数将忽略其他 init 参数）。
  - `--disable-selinux`：关闭 SELinux（仅 RHEL）。
  - `--disable-firewall`：关闭 firewalld/ufw。
  - `--skip-init`：跳过 kubeadm init。
  - `--setup-kubeconfig`：配置 root 的 kubeconfig（默认 true）。
  - `--cni`：CNI 插件（flannel|calico|none，默认 flannel）。
  - `--skip-cni`：跳过 CNI 安装。

### 示例
- `devops-infra install base`
- `devops-infra install base --mirror --dry-run`
- `devops-infra install base --mirror-source=aliyun`
- `devops-infra install base --docker-install-mode=nerdctl`
- `devops-infra install base --docker-source=国内-阿里云 --docker-registry-mirror=国内-1ms,国内-dockerproxy`
- `devops-infra install base --docker-version=27.5.1`
- `devops-infra install base --containerd-version=2.1.0 --containerd-arch=arm64 --containerd-checksum=<sha256>`
- `devops-infra install base --cni-subnet=10.88.0.0/16 --cni-route-dst=0.0.0.0/0`
- `devops-infra install base --log-dir=/var/log/devops-infra --enable-log=false --enable-trace=false`
- `devops-infra install base --skip-kernel --skip-tools`
- `devops-infra install k8s --kubernetes-version=1.28.15 --pod-network-cidr=10.244.0.0/16`

### 镜像分类与别名
- 分类前缀：`国内`/`默认`/`大陆`/`cn`/`default`，`教育网`/`教育`/`校园`/`edu`，`海外`/`境外`/`abroad`/`overseas`。
- 常用系统镜像别名：`阿里云`/`aliyun`，`腾讯云`/`tencent`，`华为云`/`huawei`，`网易`/`163`，`天翼云`/`ctyun`，`清华`/`tuna`，`北大`/`pku`，`浙大`/`zju`，`南大`/`nju`，`交大`/`sjtu`，`中科大`/`ustc`，`中科院`/`iscas`，`火山`/`volc`。
- 常用 Docker CE 镜像别名：`阿里云`/`aliyun`，`腾讯云`/`tencent`，`华为云`/`huawei`，`网易`/`163`，`清华`/`tuna`，`北大`/`pku`，`浙大`/`zju`，`南大`/`nju`，`交大`/`sjtu`，`中科大`/`ustc`，`中科院`/`iscas`，`azure`，`docker`。
- 常用 Docker Registry 别名：`1ms`，`dockerproxy`，`daocloud`，`1panel`，`阿里云`/`aliyun`，`腾讯云`/`tencent`，`dockerhub`。

## 日志与错误收集

### 结构化日志字段（Zap）
- `trace_id`：命令级关联 ID。
- `command`：执行命令（原始字符串）。
- `component`：当前安装组件（如 `docker`、`containerd`、`k8s-init`）。
- `node` / `node_addr`：执行节点信息。
- `event`：生命周期事件（`command_start`、`command_done`、`command_dry_run`）。
- `result`：`running` / `success` / `failed` / `dry_run`。
- `duration_ms`：命令耗时（毫秒）。
- `error_type`：错误分类（`exec_timeout`、`exec_nonzero`、`network_fetch`、`unsupported_os`、`validation_failed`、`unknown`）。

### 本地错误聚合产物
- 目录：`<log-dir>/errors/`（默认 `logs/errors/`）。
- `run-<id>.errors.jsonl`：失败命令事件流（逐行 JSON）。
- `run-<id>.summary.json`：运行级摘要（失败命令数、失败组件、产物路径等）。

### 控制台失败摘要
- 当安装流程存在失败命令时，CLI 会输出简要失败摘要（失败组件、部分失败命令、错误产物路径）。
- 当安装流程成功时，不输出失败摘要块。

### 常见排查步骤
1. 查看控制台摘要中给出的 `errors.jsonl` 与 `summary.json` 路径。
2. 按 `trace_id` 关联 `run.log` 与 output 日志，定位失败命令前后文。
3. 根据 `error_type` 区分：
   - `network_fetch`：镜像/网络抖动类问题；
   - `validation_failed`：配置或后置校验不满足；
   - `exec_nonzero`：命令返回非零退出码；
   - `unsupported_os`：系统发行版/能力不支持。

## 架构与流程
命令流程示例：`devops-infra install base --mirror --dry-run`

说明：CLI 默认使用本地单机执行；内部可通过 Topology/ExecutorFactory 接入 SSH，但当前命令行未暴露配置入口。

```
cmd/install_base.go
  ↓
infra/orchestration/flow.InstallBase(ctx, options)
  ↓
os.Detect (local) → topology.NewSingleNodeTopology
  ↓
executor.NewRuntime(ctx, trace/log) → execfactory.Build(node, runtime)
  ↓
os.NewDriver(osInfo, exec)
  ↓
install_operation/base.New(...).Install()
```

```
┌──────────────────────────────────────────────────────────────┐
│ CLI Layer (cobra)                                            │
│ install base / k8s / k3s / k3d | status / doctor / uninstall │
└──────────────────────────────▲───────────────────────────────┘
                               │
┌──────────────────────────────┴───────────────────────────────┐
│ Infra Orchestration                                          │
│ flow / dependency / order / idempotency                      │
└──────────────────────────────▲───────────────────────────────┘
                               │
┌──────────────────────────────┴───────────────────────────────┐
│ Install Operation                                            │
│ Base: docker / containerd / kernel / tools / mirror          │
│ Platform: k8s / k3s / k3d                                    │
└──────────────────────────────▲───────────────────────────────┘
                               │
┌──────────────────────────────┴───────────────────────────────┐
│ OS Driver + Executor                                         │
│ debian / rhel | apt / yum / systemd / sysctl                 │
│ executor: local (CLI 默认) / ssh (内部可选)                   │
└──────────────────────────────────────────────────────────────┘
```

```
┌──────────────────────────────────────────┐
│         Install Operation Layer          │
│ Base: docker / containerd / kernel       │
│ tools / mirror                           │
│ Platform: k8s / k3s / k3d                │
└──────────────────────▲───────────────────┘
                       │
┌──────────────────────┴───────────────────┐
│            OS Driver + Executor          │
│ debian / rhel | apt / yum / systemd      │
│ sysctl + executor (local / ssh)          │
└──────────────────────────────────────────┘
```
