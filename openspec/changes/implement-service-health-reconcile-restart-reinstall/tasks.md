## 1. Reconcile Framework

- [x] 1.1 Add reusable service health reconcile helper(s) with bounded strategy (`healthy -> restart -> reinstall -> fail`)
- [x] 1.2 Define outcome markers/log context for skip/restart/reinstall/fail actions

## 2. Base Install Integration

- [x] 2.1 Integrate reconcile logic into container runtime related components (`containerd`, official `docker`)
- [x] 2.2 Ensure healthy services are skipped and unhealthy services attempt restart before reinstall
- [x] 2.3 Preserve idempotent behavior across reruns on Fedora/RHEL/Ubuntu/Debian

## 3. K8s Install Integration

- [x] 3.1 Integrate reconcile logic for kubelet-oriented path in `install k8s`
- [x] 3.2 Ensure bounded recovery for kubelet service before declaring init/verify failure
- [x] 3.3 Keep `skip-init`/`skip-cni` semantics unchanged while applying reconcile behavior

## 4. Verification And Documentation

- [x] 4.1 Add/extend unit tests for healthy, restart-recovered, reinstall-recovered, and unrecoverable scenarios
- [x] 4.2 Run focused package tests and full `go test ./...` to validate no regressions
- [x] 4.3 Update README troubleshooting/behavior notes for automatic restart/reinstall reconciliation
