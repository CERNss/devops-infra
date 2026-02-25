## ADDED Requirements

### Requirement: Failure aggregator lifecycle safety on all return paths
The system SHALL close and finalize failure aggregators for install workflows even when the function returns early due to setup errors.

#### Scenario: Early return still finalizes aggregator
- **WHEN** install workflow returns before installer execution
- **THEN** aggregator close/finalization is executed

#### Scenario: Summary remains available for failed runs
- **WHEN** one or more command failures occur in a workflow
- **THEN** summary and error artifact files are flushed and available after workflow exit
