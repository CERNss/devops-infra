## Why

The current installation paths for `install base` and `install k8s` are functional but not yet reliable and observable enough for repeatable cross-distro deployment validation. This change is needed now to make installation outcomes diagnosable, testable, and consistently reproducible on priority target environments.

## What Changes

- Improve install observability first (priority `C`): unify Zap structured logging fields and add local error collection with console summaries for failed runs.
- Improve `install base` stability and validation second (priority `A`): optimize installation flow, idempotency checks, and distro-specific handling for Ubuntu 22.04, Debian 12, Rocky Linux 9, and Fedora 39.
- Improve `install k8s` stability and validation third (priority `B`): optimize kubeadm-driven install flow and verification criteria on the same target environments.
- Define explicit "run-through" acceptance criteria covering compile checks, dry-run replay, and real installation validation for base and k8s workflows.
- Keep `k3s` and `k3d` related content unchanged in this change; no scope expansion to those workflows.

## Capabilities

### New Capabilities
- `install-observability-error-collection`: Standardize install runtime logging and error event collection using Zap with consistent structured fields, persisted local artifacts, and operator-facing console summaries.
- `base-install-runthrough-reliability`: Define and enforce repeatable base-install validation behavior and pass criteria across the priority Linux distributions.
- `k8s-install-runthrough-reliability`: Define and enforce repeatable kubeadm install validation behavior and pass criteria across the priority Linux distributions.

### Modified Capabilities
None. There are currently no existing capability specs under `openspec/specs/`.

## Impact

- Affected code areas:
  - CLI command paths for `install base` and `install k8s`
  - orchestration flow and installer components under `internal/infra/orchestration`, `internal/infra/install/base`, and `internal/infra/platform/k8s`
  - logging, trace, and executor runtime integrations under `internal/middleware` and `internal/infra/executor`
- Affected behaviors:
  - structured logging format and severity usage
  - error reporting output (local files + console summary)
  - install validation and run-through test workflow on target OS distributions
- Dependencies and systems:
  - continue using Zap as the logging backend
  - no planned functional changes to `k3s`/`k3d` install surfaces in this change
