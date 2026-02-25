## ADDED Requirements

### Requirement: Independent skip-init and skip-cni semantics
The system SHALL treat `skip-init` and `skip-cni` as independent flags and SHALL NOT implicitly mutate one flag from the other.

#### Scenario: Skip init without skip cni preserves skip-cni value
- **WHEN** `install k8s` is configured with `skip-init=true` and `skip-cni=false`
- **THEN** normalized options preserve `skip-cni=false`

#### Scenario: Explicit skip cni remains respected
- **WHEN** `install k8s` is configured with `skip-cni=true`
- **THEN** CNI install path is skipped regardless of init mode
