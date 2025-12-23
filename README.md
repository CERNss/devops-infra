
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
