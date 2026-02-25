## ADDED Requirements

### Requirement: Base install compatibility targets
The system SHALL define and validate `install base` run-through behavior on Ubuntu 22.04, Debian 12, Rocky Linux 9, and Fedora 39.

#### Scenario: Ubuntu 22.04 run-through passes
- **WHEN** `install base` is executed on Ubuntu 22.04 with supported options
- **THEN** the run completes successfully and required post-install checks pass

#### Scenario: Debian 12 run-through passes
- **WHEN** `install base` is executed on Debian 12 with supported options
- **THEN** the run completes successfully and required post-install checks pass

#### Scenario: Rocky Linux 9 run-through passes
- **WHEN** `install base` is executed on Rocky Linux 9 with supported options
- **THEN** the run completes successfully and required post-install checks pass

#### Scenario: Fedora 39 run-through passes
- **WHEN** `install base` is executed on Fedora 39 with supported options
- **THEN** the run completes successfully and required post-install checks pass

### Requirement: Base install dry-run replayability
The system SHALL provide deterministic `--dry-run` output for `install base` such that planned command sequences can be inspected and replayed.

#### Scenario: Dry-run outputs command sequence without mutation
- **WHEN** `install base --dry-run` is executed
- **THEN** the output contains the planned ordered command sequence and does not mutate the host state

### Requirement: Base install idempotent rerun behavior
The system SHALL ensure rerunning `install base` on an already configured host does not fail due to already-applied steps and reports skipped or satisfied components.

#### Scenario: Second run reports already-satisfied components
- **WHEN** `install base` is executed on a host where base requirements are already installed
- **THEN** the run completes without fatal errors and reports component states as already installed or skipped
