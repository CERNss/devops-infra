## Why

`install k8s` currently fails at `kubeadm init` when swap is enabled, and this is common on Fedora/RHEL (zram) as well as Ubuntu/Debian hosts with swap entries still active. We need deterministic preflight handling now so first-run cluster bootstrap succeeds across these mainstream distributions without manual recovery steps.

## What Changes

- Add explicit swap/zram readiness handling in the `install k8s` preflight phase before `kubeadm init`.
- Ensure runtime swap is disabled and persistent swap sources are handled in an OS-aware way for Fedora, RHEL, Ubuntu, and Debian.
- Add verification behavior and failure messages that clearly distinguish "swap still enabled" from other kubelet/control-plane startup issues.
- Add/update tests for distro-aware preflight behavior and rerun reliability when swap is already disabled.

## Capabilities

### New Capabilities
- `k8s-swap-zram-preflight`: Provide deterministic, cross-distro swap and zram preflight handling for `install k8s` so kubelet can start successfully on first bootstrap.

### Modified Capabilities
- None.

## Impact

- Affected code:
- `internal/infra/platform/k8s/installer.go` (preflight behavior)
- `internal/infra/orchestration/installer_k8s.go` (component flow/option wiring, if needed)
- `cmd/install_k8s.go` and `README.md` (operator-facing behavior/docs, if flags or semantics change)
- Tests for k8s preflight/run-through behavior
- No external API contract changes expected; behavior change is in install workflow reliability and diagnostics.
