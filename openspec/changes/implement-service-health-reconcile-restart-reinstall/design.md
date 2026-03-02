## Context

The current installer framework executes components in order and uses `IsInstalled()` to skip already-satisfied steps. This is effective for existence checks (binary/file present) but insufficient for runtime correctness: a service can be installed yet unhealthy (inactive, crash-looping, broken config). In those cases, today’s behavior either skips too early or fails late in validation without an automatic recovery path.

This change introduces a health-based reconciliation layer for core services used by `install base` and `install k8s`, with explicit behavior across Fedora/RHEL and Ubuntu/Debian families.

## Goals / Non-Goals

**Goals:**
- Define a deterministic per-service reconcile state machine: `healthy -> skip`, `unhealthy -> restart`, `restart failed/unhealthy -> reinstall`, `reinstall still unhealthy -> fail`.
- Apply reconcile logic to key services first: `containerd`, `docker` (official mode), and `kubelet`.
- Keep behavior idempotent on repeated runs and bounded (no infinite restart/reinstall loops).
- Emit clear lifecycle/summary signals that indicate which reconcile action was taken and why.

**Non-Goals:**
- Full self-healing for every package/component in the project.
- Introducing long-running background controllers or daemonized reconciliation.
- Reworking SSH topology behavior beyond current single-run install execution.

## Decisions

### 1) Introduce explicit health probes separate from installation presence checks

Decision:
- Keep `IsInstalled()` for lightweight presence checks.
- Add health probes (service-active + critical command checks) as a separate reconcile gate in install flow.

Why:
- Presence and health are different signals.
- Preserves existing component contracts while adding runtime correctness.

Alternative considered:
- Fold health checks into every `IsInstalled()`.
  - Rejected because `IsInstalled()` is currently used for skip semantics and dry-run planning; mixing runtime health may produce surprising control flow.

### 2) Use bounded reconcile attempts: restart once, reinstall once

Decision:
- For each managed service:
  1) Probe health.
  2) If unhealthy, attempt one restart.
  3) Re-probe; if still unhealthy, run reinstall path once.
  4) Re-probe; fail with typed error if still unhealthy.

Why:
- Deterministic behavior, easy to reason about, avoids loops.

Alternative considered:
- Multiple restart retries with exponential backoff.
  - Rejected for now to avoid long install latency and added complexity; can be added later if required.

### 3) Reuse existing installer primitives for reinstall

Decision:
- Reinstall action reuses existing component `Install()` behavior (including package install/update and service enable/restart).

Why:
- Minimizes duplicated logic and keeps distro-specific behavior in existing installer implementations.

Alternative considered:
- Add dedicated reinstall scripts per service.
  - Rejected due to duplication and drift risk.

### 4) Standardize observability for reconcile outcomes

Decision:
- Add explicit outcome labels in logs/summaries (e.g. `reconcile_skip_healthy`, `reconcile_restart`, `reconcile_reinstall`, `reconcile_failed`).
- Preserve existing failure aggregation artifacts while enriching component action traces.

Why:
- Operators need to distinguish “fresh install” from “recovered unhealthy service”.

Alternative considered:
- Keep current generic command success/failure only.
  - Rejected because it hides action intent and complicates troubleshooting.

## Risks / Trade-offs

- [Risk] False-negative health probes can trigger unnecessary reinstall.
  - Mitigation: keep probes minimal/stable (`systemctl is-active` + required binary checks), and scope initial rollout to core services only.

- [Risk] Reinstall path may be destructive on hosts with customized service config.
  - Mitigation: preserve existing backup behavior where present; document reconcile policy and allow future opt-out flag if needed.

- [Risk] Added orchestration complexity can introduce regressions in skip behavior.
  - Mitigation: add table-driven tests for healthy/unhealthy/recovery/failure paths and assert idempotent rerun behavior.

## Migration Plan

1. Implement reconcile helper and integrate it into base/k8s service-bearing components.
2. Add unit tests for state transitions and failure classification.
3. Update README/operator docs to explain automatic restart/reinstall behavior.
4. Roll out with current default behavior enabled (no flag change).
5. If unexpected regressions occur, rollback by bypassing reconcile helper and returning to existing `IsInstalled()->Install()` flow.

## Open Questions

- Should we expose a CLI flag to disable automatic reinstall and allow restart-only mode?
- Should reconcile outcomes be surfaced in a dedicated summary section (beyond current failure summary)?
- Should kubelet reconcile be skipped automatically when `--skip-init` is true, or still enforced when kubelet package is managed?
