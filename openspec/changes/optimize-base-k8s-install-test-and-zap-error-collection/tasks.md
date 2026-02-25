## 1. C Phase: Observability Schema and Error Taxonomy

- [x] 1.1 Define a canonical install event schema (fields, types, required/optional rules) for `install base` and `install k8s`
- [x] 1.2 Implement a shared error classifier that maps failures to `exec_timeout`, `exec_nonzero`, `network_fetch`, `unsupported_os`, `validation_failed`, and `unknown`
- [x] 1.3 Add unit tests for error classification coverage and unknown-fallback behavior

## 2. C Phase: Executor and Middleware Structured Logging Integration

- [x] 2.1 Refactor local executor command lifecycle reporting to emit canonical structured fields for success/failure
- [x] 2.2 Refactor SSH executor command lifecycle reporting to emit the same canonical structured fields
- [x] 2.3 Propagate `component` and `node` metadata through runtime context so logs can be correlated across layers
- [x] 2.4 Standardize Zap field keys and severity usage for install lifecycle events (`info`/`warn`/`error`)
- [x] 2.5 Add tests validating required fields are present in emitted logs for both success and failure paths

## 3. C Phase: Local Error Aggregation Artifacts and Console Summary

- [x] 3.1 Implement JSONL failure event sink for per-run local error aggregation artifacts
- [x] 3.2 Implement run summary artifact generation (counts, failed components/commands, artifact paths)
- [x] 3.3 Integrate artifact writers into install flow lifecycle without changing existing CLI flags
- [x] 3.4 Implement concise console failure summary rendering with artifact path hints
- [x] 3.5 Add tests to ensure failure summary appears only on failed runs and is omitted on successful runs

## 4. A Phase: Base Install Validation Contract and Idempotency

- [x] 4.1 Define and implement explicit post-install checks for kernel prerequisites, common tools, and container runtime readiness
- [x] 4.2 Improve `install base` idempotent rerun behavior to mark already-satisfied components without hard failures
- [x] 4.3 Add/update tests for base component state detection (`IsInstalled`) and rerun outcomes
- [x] 4.4 Implement deterministic `install base --dry-run` verification rules (ordering and normalization)
- [x] 4.5 Add dry-run verification tests for base workflow output consistency

## 5. A Phase: Base Cross-Distro Run-through (Ubuntu/Debian/Rocky/Fedora)

- [ ] 5.1 Add distro-specific validation adapters for Ubuntu 22.04 and Debian 12 base run-through checks
- [ ] 5.2 Add distro-specific validation adapters for Rocky Linux 9 and Fedora 39 base run-through checks
- [ ] 5.3 Add automated smoke scripts/tests for real base install verification per target distro
- [x] 5.4 Ensure external mirror/network failures are classified and reported separately from deterministic logic failures

## 6. B Phase: K8s Install Validation Contract (Init, Skip-Init, CNI)

- [x] 6.1 Implement default `install k8s` pass checks (kubeadm init success, kubelet active, kubeconfig setup when enabled)
- [x] 6.2 Implement `--skip-init` path checks that validate only applicable steps
- [x] 6.3 Implement CNI verification checks for `flannel`, `calico`, and `none/--skip-cni`
- [x] 6.4 Add/update tests for k8s installer branch behavior and verification output
- [x] 6.5 Implement deterministic `install k8s --dry-run` verification rules and tests

## 7. B Phase: K8s Cross-Distro Run-through (Ubuntu/Debian/Rocky/Fedora)

- [ ] 7.1 Add run-through automation for Ubuntu 22.04 and Debian 12 k8s install validation
- [ ] 7.2 Add run-through automation for Rocky Linux 9 and Fedora 39 k8s install validation
- [ ] 7.3 Add result capture that links k8s run-through failures to local error artifacts and console summaries

## 8. Final Verification and Delivery

- [ ] 8.1 Validate `go test ./...` remains green after all C/A/B changes
- [ ] 8.2 Execute full dry-run validation for `install base` and `install k8s` and confirm deterministic output checks pass
- [ ] 8.3 Execute full run-through verification for the four target distros and record pass/fail evidence
- [ ] 8.4 Confirm no functional behavior changes were introduced for `k3s` and `k3d` command surfaces
- [x] 8.5 Update project documentation with logging schema, artifact locations, and troubleshooting workflow
