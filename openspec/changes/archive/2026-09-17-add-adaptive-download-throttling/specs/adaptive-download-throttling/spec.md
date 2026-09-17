## Purpose

Combines the weekly bandwidth schedule (`download-bandwidth-schedule`) and Jellyfin's active-playback state into a single effective download policy, and enforces it — as a shared rate limit or a full stop — across the `download` and `resume-downloads` commands.

## ADDED Requirements

### Requirement: Independent Criteria
The weekly schedule and Jellyfin active-playback detection SHALL each be independently configurable — either, both, or neither may be active at a given time — and the effective policy SHALL be computed from whichever of the two are currently enabled/contributing, without requiring the other to be configured.

#### Scenario: Only the schedule is configured
- **WHEN** Jellyfin active-playback detection is disabled and the weekly schedule has an active `throttle` window
- **THEN** the effective policy SHALL be `throttle`

#### Scenario: Neither is configured
- **WHEN** no schedule window is defined and Jellyfin active-playback detection is disabled
- **THEN** the effective policy SHALL always be `none`, matching current (pre-change) download behavior

### Requirement: Effective Policy Combination
The system SHALL compute one effective policy from the weekly schedule's currently active action and the Jellyfin signal's currently contributed action, selecting the most restrictive of the two using the ordering `stop` > `throttle` > `none`.

#### Scenario: Schedule says throttle, Jellyfin says stop
- **WHEN** the weekly schedule's active action is `throttle` and the Jellyfin signal is currently contributing `stop`
- **THEN** the effective policy SHALL be `stop`

#### Scenario: Both say throttle
- **WHEN** both the weekly schedule and the Jellyfin signal are currently contributing `throttle`
- **THEN** the effective policy SHALL be `throttle`, enforced at the single shared throttle rate (see Single Shared Throttle Rate)

### Requirement: Active-Playback Definition
The system SHALL consider Jellyfin to have active playback only when at least one reported session has a currently playing media item that is not paused. A session with no playing item, or with a playing item that is paused, SHALL NOT count as active playback.

#### Scenario: Idle session
- **WHEN** Jellyfin reports a connected session with no currently playing item
- **THEN** the system SHALL treat this as no active playback

#### Scenario: Paused playback
- **WHEN** Jellyfin reports a session whose playback is paused
- **THEN** the system SHALL treat this as no active playback

#### Scenario: Actively playing session
- **WHEN** Jellyfin reports a session actively playing a media item
- **THEN** the system SHALL treat this as active playback

### Requirement: Configurable Jellyfin Action
When Jellyfin active-playback detection is enabled, the system SHALL apply a configured action — `throttle` or `stop` — as the Jellyfin signal's contribution to the effective policy whenever active playback is detected. When active-playback detection is disabled, the Jellyfin signal SHALL contribute `none` at all times.

#### Scenario: Configured to stop
- **WHEN** Jellyfin active-playback detection is enabled with action `stop`, and active playback is currently detected
- **THEN** the Jellyfin signal SHALL contribute `stop` to the effective policy

#### Scenario: Detection disabled
- **WHEN** Jellyfin active-playback detection is disabled
- **THEN** the Jellyfin signal SHALL always contribute `none`, regardless of Jellyfin's actual playback state

### Requirement: Fail-Open on Jellyfin Unreachable
When Jellyfin cannot be reached to determine active-playback state, the system SHALL treat this as no active playback, so a Jellyfin outage never causes downloads to throttle or stop on its account.

#### Scenario: Jellyfin unreachable
- **WHEN** a Jellyfin active-playback check fails (timeout, connection error, or authentication error)
- **THEN** the Jellyfin signal SHALL contribute `none` to the effective policy for that check

### Requirement: Single Shared Throttle Rate
The system SHALL apply one configured throttle rate as an aggregate cap shared across all concurrently active transfers, regardless of whether the `throttle` policy was triggered by the weekly schedule, by Jellyfin active-playback, or by both simultaneously. The system SHALL NOT apply a separate rate per signal or per transfer.

#### Scenario: Two transfers in progress while throttled
- **WHEN** the effective policy is `throttle` and two transfers are in progress at once
- **THEN** their combined throughput SHALL be capped at the configured throttle rate, not each individually capped at that rate

### Requirement: No Cap Under a None Policy
When the effective policy is `none`, transfers SHALL proceed without any throttle-rate cap, unchanged from current behavior.

#### Scenario: No signal active
- **WHEN** neither the weekly schedule nor Jellyfin is contributing a restriction
- **THEN** transfers SHALL run at unrestricted speed, as before this change

### Requirement: No New Claims While Stopped
While the effective policy is `stop`, the `download` and `resume-downloads` commands SHALL NOT begin transferring any item that has not already started, deferring it to a later run (see `media-download-scheduling`). An item deferred this way SHALL remain eligible to be claimed normally once the effective policy is no longer `stop`.

#### Scenario: Stop begins before an item starts
- **WHEN** the effective policy is `stop` and a worker has no in-progress transfer
- **THEN** the worker SHALL NOT begin a new transfer, and SHALL leave that item available for a later run to attempt

### Requirement: Aborting an In-Progress Transfer on Stop
When the effective policy transitions to `stop` while a transfer is in progress, the system SHALL abort that transfer within a short, bounded delay, independent of the file's remaining size.

#### Scenario: Stop begins mid-transfer
- **WHEN** the effective policy becomes `stop` while a large file is partway through downloading
- **THEN** the transfer SHALL be aborted without waiting for it to complete

### Requirement: Policy Abort Is Not a Failure
An item whose transfer was aborted because the effective policy was `stop` SHALL be recorded with a status distinct from a network or transfer failure. It SHALL NOT count against that item's retry attempt budget, and SHALL NOT trigger a failure notification.

#### Scenario: Aborted transfer does not consume retries
- **WHEN** an item's transfer is aborted due to the effective policy being `stop`
- **THEN** its remaining retry attempt budget SHALL be unchanged, and no failure notification SHALL be sent for this occurrence

### Requirement: Restart From Zero (v1 Scope)
When an item whose transfer was previously aborted by policy is attempted again, the system SHALL restart its transfer from the beginning. Preserving and resuming from the bytes already transferred before the abort is out of scope for this capability.

#### Scenario: Retrying a policy-aborted item
- **WHEN** an item previously aborted by policy is attempted again after the effective policy is no longer `stop`
- **THEN** the transfer SHALL start over from 0% rather than continuing from its previous progress

### Requirement: Applies to Both Download Entry Points
The effective policy SHALL be enforced identically by both the `download` command and the `resume-downloads` command, since both perform the same kind of file transfer.

#### Scenario: Resume command respects the policy
- **WHEN** the effective policy is `stop` and the `resume-downloads` command runs
- **THEN** it SHALL NOT begin transferring any incomplete download, the same as the `download` command would not

### Requirement: Jellyfin Poll Interval
The system SHALL re-check Jellyfin's active-playback state on a configurable interval, independent of any individual transfer's own progress-reporting cadence.

#### Scenario: Long-running transfer picks up a later Jellyfin state change
- **WHEN** Jellyfin playback starts partway through an already-in-progress transfer
- **THEN** the system SHALL detect it at the next Jellyfin poll and apply the resulting policy without waiting for the current transfer to finish

### Requirement: Schedule Evaluation Is Not Delayed By Jellyfin Polling
The weekly-schedule portion of the effective policy SHALL be evaluated at least as frequently as the Jellyfin poll interval, so a schedule boundary is never delayed by a slower Jellyfin check.

#### Scenario: Schedule boundary reached between Jellyfin polls
- **WHEN** a schedule window's start time is reached partway between two scheduled Jellyfin polls
- **THEN** the schedule's contribution to the effective policy SHALL update at that boundary without waiting for the next Jellyfin poll
