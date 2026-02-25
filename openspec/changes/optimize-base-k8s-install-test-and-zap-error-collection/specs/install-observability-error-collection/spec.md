## ADDED Requirements

### Requirement: Unified structured installation logs
The system SHALL emit structured logs for `install base` and `install k8s` with a stable schema including at minimum `trace_id`, `command`, `component`, `node`, `duration_ms`, `result`, and `error_type` (when failed).

#### Scenario: Successful command execution emits normalized fields
- **WHEN** an installation command step completes successfully
- **THEN** the emitted log entry contains the required structured fields with `result=success`

#### Scenario: Failed command execution emits normalized failure fields
- **WHEN** an installation command step fails
- **THEN** the emitted log entry contains the required structured fields with `result=failed` and a non-empty `error_type`

### Requirement: Local error aggregation artifacts
The system SHALL persist aggregated error events for each installation run to local files in machine-readable format.

#### Scenario: Failure creates aggregated local error artifact
- **WHEN** any step in `install base` or `install k8s` fails
- **THEN** the run output includes a local error aggregation artifact containing each failed step and its correlation identifiers

#### Scenario: Successful run still produces traceable artifact set
- **WHEN** an installation run completes without failures
- **THEN** the run output includes local structured logs and trace references with zero aggregated failures

### Requirement: Operator-facing console failure summary
The system SHALL print a concise console summary for failed runs that includes failed components, failed commands, and the local artifact locations.

#### Scenario: Console summary shown for run with failures
- **WHEN** an installation run ends with one or more failures
- **THEN** the console output includes a failure summary and explicit paths to local diagnostic artifacts

#### Scenario: Console summary omitted for successful run
- **WHEN** an installation run ends successfully
- **THEN** the console output does not include a failure summary block
