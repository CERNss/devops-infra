## ADDED Requirements

### Requirement: K8s install compatibility targets
The system SHALL define and validate `install k8s` run-through behavior on Ubuntu 22.04, Debian 12, Rocky Linux 9, and Fedora 39.

#### Scenario: Ubuntu 22.04 k8s run-through passes
- **WHEN** `install k8s` is executed on Ubuntu 22.04 with supported options
- **THEN** kubeadm-based installation completes and required post-install checks pass

#### Scenario: Debian 12 k8s run-through passes
- **WHEN** `install k8s` is executed on Debian 12 with supported options
- **THEN** kubeadm-based installation completes and required post-install checks pass

#### Scenario: Rocky Linux 9 k8s run-through passes
- **WHEN** `install k8s` is executed on Rocky Linux 9 with supported options
- **THEN** kubeadm-based installation completes and required post-install checks pass

#### Scenario: Fedora 39 k8s run-through passes
- **WHEN** `install k8s` is executed on Fedora 39 with supported options
- **THEN** kubeadm-based installation completes and required post-install checks pass

### Requirement: K8s init outcome verification
The system SHALL define pass criteria for `install k8s` that include kubeadm init completion, kubelet active state, and root kubeconfig availability when kubeconfig setup is enabled.

#### Scenario: Default init path satisfies required checks
- **WHEN** `install k8s` is executed with default init behavior
- **THEN** kubeadm init succeeds and required service/config checks are satisfied

#### Scenario: Skip-init path preserves validation scope
- **WHEN** `install k8s --skip-init` is executed
- **THEN** the workflow skips init-specific checks and validates only applicable steps

### Requirement: CNI plugin deployment verification
The system SHALL define verification behavior for configured CNI modes `flannel`, `calico`, and `none`.

#### Scenario: Flannel mode applies flannel manifests
- **WHEN** `install k8s --cni=flannel` is executed
- **THEN** flannel deployment steps are applied and reported as successful

#### Scenario: Calico mode applies calico manifests
- **WHEN** `install k8s --cni=calico` is executed
- **THEN** calico deployment steps are applied and reported as successful

#### Scenario: None mode skips CNI deployment
- **WHEN** `install k8s --cni=none` or `--skip-cni` is executed
- **THEN** CNI deployment is skipped and the skip state is explicitly reported
