## ADDED Requirements

### Requirement: Deterministic service health reconciliation
The system SHALL execute a deterministic reconcile state machine for managed services where healthy services are skipped, unhealthy services are restarted, and unrecovered services are reinstalled before failing the workflow.

#### Scenario: Healthy service is skipped
- **WHEN** a managed service is detected as healthy before install action
- **THEN** the workflow skips restart and reinstall for that service

#### Scenario: Unhealthy service is recovered by restart
- **WHEN** a managed service is unhealthy and a restart makes it healthy
- **THEN** the workflow marks the service as recovered and continues without reinstall

#### Scenario: Unhealthy service escalates to reinstall
- **WHEN** a managed service remains unhealthy after restart
- **THEN** the workflow performs exactly one reinstall attempt for that service

#### Scenario: Reconcile fails after restart and reinstall
- **WHEN** a managed service remains unhealthy after one restart and one reinstall attempt
- **THEN** the workflow fails with an explicit reconciliation failure

### Requirement: Bounded recovery attempts
The system SHALL bound automatic recovery attempts per managed service to prevent unbounded loops.

#### Scenario: Restart attempt is bounded
- **WHEN** a managed service first fails health probe
- **THEN** the workflow performs no more than one automatic restart attempt before escalation

#### Scenario: Reinstall attempt is bounded
- **WHEN** restart does not restore service health
- **THEN** the workflow performs no more than one automatic reinstall attempt before final failure

### Requirement: Cross-distro reconcile compatibility
The system SHALL apply reconcile behavior consistently on Fedora, RHEL, Ubuntu, and Debian for managed install services.

#### Scenario: RHEL-family reconcile path succeeds
- **WHEN** reconciliation runs on Fedora or RHEL hosts
- **THEN** service health probe and restart/reinstall actions execute using the supported service manager path

#### Scenario: Debian-family reconcile path succeeds
- **WHEN** reconciliation runs on Ubuntu or Debian hosts
- **THEN** service health probe and restart/reinstall actions execute using the supported service manager path

### Requirement: Workflow-level reconcile coverage
The system SHALL apply reconciliation to key services used by install workflows, including container runtime and kubelet-related services.

#### Scenario: Base workflow reconciles container runtime service
- **WHEN** `install base` processes container runtime components
- **THEN** the workflow reconciles runtime service health before declaring component success

#### Scenario: K8s workflow reconciles kubelet service
- **WHEN** `install k8s` processes kubelet-dependent steps
- **THEN** the workflow reconciles kubelet service health before init/verify success is reported
