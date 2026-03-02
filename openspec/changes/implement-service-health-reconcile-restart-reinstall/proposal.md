## Why

Current install flows can skip work when binaries/files already exist, but they do not provide a unified reconcile strategy when services are present but unhealthy. We need deterministic health-based behavior now so operators get automatic recovery (restart/reinstall) instead of repeated manual troubleshooting.

## What Changes

- Introduce a service reconcile contract for install workflows: `healthy -> skip`, `unhealthy -> restart`, `restart failed -> reinstall`.
- Add component-level health checks and reconcile actions for key runtime/control-plane services (containerd, docker mode service, kubelet).
- Add bounded retry/error handling and explicit failure reasons when reconcile cannot restore health.
- Add structured logging fields and summary entries to show whether each component was skipped, restarted, reinstalled, or failed.
- Add tests for normal, degraded, and unrecoverable scenarios across Fedora/RHEL/Ubuntu/Debian families.

## Capabilities

### New Capabilities
- `service-health-reconcile`: Define and enforce install-time health-based reconciliation, including skip/restart/reinstall decision rules and bounded recovery behavior.
- `service-recovery-observability`: Emit consistent lifecycle signals for recovery actions (skip, restart, reinstall, fail) so operators can audit why actions happened.

### Modified Capabilities
- None.

## Impact

- Affected code:
- install component orchestration (`internal/infra/install/base`, `internal/infra/platform/k8s`, `internal/infra/orchestration`)
- OS driver service operations for Debian/RHEL families
- structured logging and failure summary integration
- unit/integration tests for component health reconciliation
- No external API break expected; behavior changes are in install execution strategy and diagnostics.
