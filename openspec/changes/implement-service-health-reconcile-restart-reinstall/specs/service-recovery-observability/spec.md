## ADDED Requirements

### Requirement: Reconcile action observability
The system SHALL emit explicit lifecycle signals for reconciliation outcomes so operators can distinguish skip, restart, reinstall, and unrecoverable failures.

#### Scenario: Healthy skip is observable
- **WHEN** a service is healthy and reconciliation skips recovery actions
- **THEN** logs include an explicit event indicating a health-based skip outcome

#### Scenario: Restart recovery is observable
- **WHEN** a service is unhealthy and restored by restart
- **THEN** logs include an explicit restart-attempt event and a restart-recovered outcome

#### Scenario: Reinstall recovery is observable
- **WHEN** restart does not recover a service and reinstall is attempted
- **THEN** logs include explicit restart-failed and reinstall-attempt events with final outcome

#### Scenario: Unrecoverable failure is observable
- **WHEN** a service remains unhealthy after bounded recovery attempts
- **THEN** logs include an explicit reconcile-failed outcome and failure reason

### Requirement: Failure summary includes reconcile context
The system SHALL include reconciliation context in run-level failure artifacts to improve operator diagnosis.

#### Scenario: Reconcile failure is summarized with service context
- **WHEN** workflow terminates due to unrecoverable service reconciliation
- **THEN** failure artifacts include the affected component/service and the final recovery stage reached

#### Scenario: Successful runs omit failure summary block
- **WHEN** reconciliation completes without unrecoverable failures
- **THEN** the console failure summary block remains omitted

### Requirement: Dry-run reconciliation visibility
The system SHALL represent planned reconciliation actions in dry-run mode without mutating host state.

#### Scenario: Dry-run shows planned restart/reinstall decisions
- **WHEN** install workflow runs with `--dry-run` and detects would-be unhealthy service conditions
- **THEN** output includes deterministic planned reconcile actions without executing service mutations
