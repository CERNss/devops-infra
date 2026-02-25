## Context

A focused bug-fix change is required for three regressions discovered during review: (1) `skip-init` now forces `skip-cni`, (2) failure aggregators are not guaranteed to close on early returns, and (3) runtime-related tests can write artifacts into repository paths.

## Goals / Non-Goals

**Goals:**
- Restore documented independent semantics of `skip-init` and `skip-cni`.
- Guarantee aggregator close/finalization for `install base` and `install k8s` on every return path.
- Make new tests side-effect free and add regression coverage for all three issues.

**Non-Goals:**
- No behavior changes for `k3s`/`k3d`.
- No redesign of aggregator data model.
- No broader workflow redesign beyond these targeted fixes.

## Decisions

- Keep normalization logic declarative and non-mutating for unrelated flags.
  Alternative rejected: auto-forcing `skip-cni` from `skip-init`.
- Register aggregator cleanup (`defer`) immediately after successful creation.
  Alternative rejected: late defer placement near installer execution.
- Make runtime tests use explicit noop sinks or isolated temp paths.
  Alternative rejected: relying on default runtime initialization in tests.

## Risks / Trade-offs

- [Risk] Changing skip flag behavior may expose previously hidden flow assumptions.
  → Mitigation: add direct unit tests for normalization and branch behavior.
- [Risk] Earlier defer placement may output summary in additional failure paths.
  → Mitigation: keep `HasFailures()` guard and assert expected behavior with tests.
- [Risk] Test isolation changes may reduce integration realism.
  → Mitigation: keep scope on unit tests and retain workflow tests elsewhere.

## Migration Plan

1. Remove implicit `skip-cni` mutation from k8s option normalization.
2. Move aggregator cleanup defer to immediately follow successful aggregator creation in both workflows.
3. Update side-effectful tests to use isolated runtime setup.
4. Add/adjust regression tests and run `go test ./...`.

## Open Questions

- None for this patch-level change.
