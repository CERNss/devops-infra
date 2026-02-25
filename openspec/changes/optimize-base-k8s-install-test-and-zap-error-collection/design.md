## Context

Current `install base` and `install k8s` flows are usable but operationally hard to diagnose across distributions when failures happen in different layers (executor, OS driver, component installer, platform installer). Logging exists, but field conventions and failure summarization are not yet a strict contract for operators. The project also lacks a reliable cross-distro run-through validation definition for the four priority targets: Ubuntu 22.04, Debian 12, Rocky Linux 9, and Fedora 39.

This design implements the proposal scope in three ordered phases:
1. `C`: observability and error collection first
2. `A`: base install reliability second
3. `B`: k8s install reliability third

`k3s` and `k3d` remain out of scope and unchanged.

## Goals / Non-Goals

**Goals:**
- Establish a stable structured log schema for `install base` and `install k8s` based on Zap fields.
- Add local machine-readable error aggregation artifacts and concise console failure summaries.
- Define and implement repeatable run-through checks for `install base` on Ubuntu 22.04, Debian 12, Rocky Linux 9, and Fedora 39.
- Define and implement repeatable run-through checks for `install k8s` on the same targets.
- Preserve backward compatibility for existing CLI usage and default execution behavior.

**Non-Goals:**
- No functional expansion of `install k3s` or `install k3d`.
- No multi-node orchestration feature delivery in this change.
- No external observability backend integration (for example Elasticsearch, Loki, SaaS log collectors).
- No redesign of OpenSpec workflow or schema.

## Decisions

### Decision 1: Introduce a canonical install event schema at executor boundary
- Choice: normalize command lifecycle and error metadata where commands are executed (`internal/infra/executor` + middleware sinks), then propagate consistent fields to Zap and local artifacts.
- Rationale: this is the narrowest choke point covering base and k8s flows without duplicating instrumentation in every installer.
- Required fields: `trace_id`, `command`, `component`, `node`, `start_time`, `duration_ms`, `result`, `error_type`, `error_message`, `stdout_path`, `stderr_path`.
- Alternatives considered:
  - Instrument each installer independently: rejected because schema drift is likely and maintenance cost is higher.
  - Only parse existing plain logs after the fact: rejected because correlation and error typing are incomplete.

### Decision 2: Build a two-layer failure artifact model (event stream + summary)
- Choice: write local JSONL error event stream plus one concise run summary file.
- Artifact plan:
  - event stream: append-only records for failed steps
  - summary: run-level counters, failed components, failed commands, artifact paths
- Rationale: JSONL is suitable for streaming and debugging; summary is optimized for operator quick triage.
- Alternatives considered:
  - Summary only: rejected because detail loss blocks root-cause analysis.
  - Full custom database: rejected as unnecessary complexity for current scope.

### Decision 3: Keep console output short and deterministic
- Choice: print failure summary only when run fails; include top failed steps and local artifact locations.
- Rationale: operators need fast signal without overwhelming stdout; details stay in files.
- Alternatives considered:
  - Always print verbose diagnostics: rejected due to noisy output and poor readability.

### Decision 4: Add explicit error taxonomy for collection and reporting
- Choice: map failures into stable classes such as `exec_timeout`, `exec_nonzero`, `network_fetch`, `unsupported_os`, `validation_failed`, and `unknown`.
- Rationale: typed failures improve triage and enable deterministic acceptance checks.
- Alternatives considered:
  - Raw error strings only: rejected because classification becomes inconsistent and brittle.

### Decision 5: Implement cross-distro run-through verification as layered checks
- Choice: define three layers of acceptance for each target distro:
  - build/test baseline (`go test ./...`)
  - dry-run determinism checks
  - real install verification checks
- Rationale: this separates fast feedback from environment-dependent validation.
- Alternatives considered:
  - only real install tests: rejected due to slow feedback and hard debugging.

### Decision 6: Stabilize `install base` via explicit post-install validation contract
- Choice: add/standardize checks for kernel prerequisites, tools availability, container runtime state, and idempotent rerun behavior.
- Rationale: prevents false positives where commands succeed but environment is not ready.
- Alternatives considered:
  - rely on command exit code only: rejected because partial success is common in infra setup.

### Decision 7: Stabilize `install k8s` via mode-aware verification contract
- Choice: validate init path and skip-init path separately; validate CNI outcomes for `flannel`, `calico`, and `none/skip-cni`.
- Rationale: current k8s flow branches by flags and requires branch-specific verification.
- Alternatives considered:
  - one generic post-check set: rejected because it produces false failures on valid skip paths.

### Decision 8: Preserve CLI compatibility and isolate behavior changes to observability/reliability internals
- Choice: avoid breaking flag semantics; implement improvements under existing command surfaces.
- Rationale: reduces migration cost and risk for existing users.
- Alternatives considered:
  - introduce a new command family for v2 behavior: rejected as unnecessary for this scope.

## Risks / Trade-offs

- [Risk] Added logging and artifact writes increase I/O overhead on long runs.
  - Mitigation: keep artifact schema compact, write summaries once per run, and gate verbose payload fields.

- [Risk] Error taxonomy may misclassify some failure paths initially.
  - Mitigation: include fallback `unknown` class and add targeted unit tests for classifier coverage.

- [Risk] Cross-distro run-through may fail due to upstream mirror/network instability rather than code regressions.
  - Mitigation: classify external dependency failures explicitly and separate infra-noise from deterministic regressions.

- [Risk] Fedora/Rocky package and service differences can create branch complexity.
  - Mitigation: centralize distro-conditional logic in OS driver and validation adapters, not scattered installer logic.

- [Risk] K8s real install checks are expensive and environment-sensitive.
  - Mitigation: stage checks by depth (smoke first, full run-through second) and preserve reproducible command traces.

## Migration Plan

1. Implement phase `C` first:
- add canonical install event model and Zap field contract
- add local error event stream + run summary writer
- add failure console summary rendering
- add classifier and unit tests

2. Implement phase `A` second:
- codify base post-install checks and idempotency assertions
- refine distro-specific handling for the four target distros
- add dry-run determinism verification tests for base

3. Implement phase `B` third:
- codify k8s pass criteria per path (default init, skip-init)
- codify CNI verification for flannel/calico/none
- add dry-run and run-through verification tests for k8s

4. Deployment strategy:
- ship changes behind existing command surfaces without new mandatory flags
- maintain current defaults for logging/tracing toggles

5. Rollback strategy:
- revert to previous logging sinks and summary rendering paths if regressions occur
- keep install orchestration behavior unchanged so rollback is code-level, not operationally disruptive

## Open Questions

- Should the run summary include only failed steps, or both failed and skipped steps by default?
- What is the exact retention policy for local error artifacts under `logs/` and `trace/` directories?
- Should dry-run determinism checks ignore timestamps and transient path fragments by normalization rules?
- For k8s run-through on CI, do we require full kubeadm init on every distro per PR, or gate full runs to scheduled pipelines?
