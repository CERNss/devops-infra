## ADDED Requirements

### Requirement: Unit tests must not write artifacts to repository paths
The test suite SHALL avoid creating persistent `logs/` or `trace/` directories under source folders during normal execution.

#### Scenario: Runtime-related tests are side-effect free
- **WHEN** runtime/context unit tests run
- **THEN** repository source directories remain free of generated log and trace artifacts

#### Scenario: Tests with filesystem needs use isolated temporary directories
- **WHEN** a test requires file output paths
- **THEN** it uses temporary directories or noop sinks instead of repository paths
