## Why

Recent refactors introduced three correctness issues: `skip-init` now implicitly forces `skip-cni`, failure aggregators are not guaranteed to close on early returns, and new tests can pollute the repository with generated `logs/` and `trace/` directories. These need to be fixed now to avoid behavior regressions and unstable developer workflows.

## What Changes

- Restore flag semantics so `install k8s --skip-init` does not automatically change `--skip-cni` behavior.
- Fix failure aggregator lifecycle in both `install base` and `install k8s` so summaries and file handles are finalized even when returning early.
- Make runtime/context related tests side-effect free (no repository writes during test runs).
- Add targeted regression tests for all three issues.

## Capabilities

### New Capabilities
- `k8s-flag-semantics-consistency`: Ensure `skip-init` and `skip-cni` remain independently controllable and test-covered.
- `failure-aggregator-lifecycle-safety`: Ensure aggregator close/finalize behavior is guaranteed for all return paths.
- `test-side-effect-isolation`: Ensure unit tests do not create persistent artifacts in the repository workspace.

### Modified Capabilities
None. There are currently no existing capability specs under `openspec/specs/`.

## Impact

- Affected code:
  - `internal/infra/orchestration/installer_k8s.go`
  - `internal/infra/orchestration/flow/install_base.go`
  - `internal/infra/executor/runtime_test.go` and related tests
- Affected behavior:
  - k8s CLI flag semantics
  - failure summary reliability on error paths
  - repository cleanliness after running tests
- No API surface expansion and no `k3s`/`k3d` scope changes.
