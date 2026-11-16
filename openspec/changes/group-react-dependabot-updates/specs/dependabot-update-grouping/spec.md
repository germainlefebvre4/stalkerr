# Spec Delta

## Purpose

Ensures that dependencies which must stay at mutually compatible versions (the frontend React runtime and its type packages) are proposed by Dependabot in a single pull request, so that no update PR breaks CI because of a version mismatch.

## ADDED Requirements

### Requirement: React packages are updated together by Dependabot
Dependabot SHALL propose updates to `react`, `react-dom`, `@types/react` and `@types/react-dom` in the `/frontend` npm project as a single grouped pull request, never as separate pull requests for each package.

#### Scenario: New React release touches several packages
- **WHEN** a new release is available for `react` and `react-dom`
- **THEN** Dependabot SHALL open one pull request that updates both packages (and their `@types/*` counterparts when a matching release exists)
- **AND** the lockfile in that pull request SHALL resolve `react` and `react-dom` to the same version

#### Scenario: Only the type packages have a new release
- **WHEN** a new release exists for `@types/react` or `@types/react-dom` but not for `react` or `react-dom`
- **THEN** Dependabot SHALL update the type packages in the same grouped pull request rather than a separate one

### Requirement: Grouped React update passes frontend CI
A Dependabot pull request that updates React packages SHALL NOT be considered merge-ready unless `react` and `react-dom` resolve to the same version and the frontend CI checks (`npm run lint`, `npm run build`, `npm test`) all pass.

#### Scenario: Mismatched React versions are blocked
- **WHEN** a pull request leaves `react` and `react-dom` at different versions
- **THEN** the frontend CI `Test` check SHALL fail
- **AND** the pull request SHALL NOT be merged until both packages resolve to the same version
