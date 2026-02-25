## 1. K8s Flag Semantics Fix

- [x] 1.1 Remove implicit `skip-cni` mutation from k8s option normalization
- [x] 1.2 Add/adjust tests to verify `skip-init` and `skip-cni` remain independent

## 2. Aggregator Lifecycle Safety

- [x] 2.1 Move `install base` aggregator defer setup to immediately after aggregator creation
- [x] 2.2 Move `install k8s` aggregator defer setup to immediately after aggregator creation
- [x] 2.3 Add tests for early-return paths to ensure aggregator close/finalization is guaranteed

## 3. Test Side-Effect Isolation

- [x] 3.1 Refactor runtime-related tests to avoid default runtime filesystem writes in repository paths
- [x] 3.2 Add assertions that no side-effect artifacts are required for these tests

## 4. Validation

- [x] 4.1 Run `go test ./...` and ensure all tests pass
- [x] 4.2 Re-verify no functional changes for `k3s`/`k3d` command surfaces
