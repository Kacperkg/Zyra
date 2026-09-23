# RFC-001: Zyra — Replacing the Current Database Check Ticket System

| Field | Value |
| --- | --- |
| Status | Draft — awaiting review |
| Date | 23 September 2026 |
| Decision requested | Review the proposed approach and agree what the proof of concept must demonstrate |
| Product requirements | [documentation.md](documentation.md) |

## 1. Summary

This RFC proposes Zyra as a replacement for the current database-check ticket system, which is slow and has limited room for expansion. Zyra will process reports from existing health-check scripts, evaluate the checks configured for each database, and provide a responsive way to investigate issues and preserve their history.

The MVP is a proof of concept for Oracle daily checks received twice a day. SQL checks and issue pages, plus the separate standby-alert email running every 30 minutes, are committed scope for Release 1.0.

This RFC explains the proposed approach, its tradeoffs, and the decisions needed before implementation. Detailed screens, permissions, thresholds, folder structure, and report examples remain in [the product documentation](documentation.md).

## 2. Problem and motivation

The current system has two reported limitations:

- **Slow operation:** speed is a problem in the existing ticket system, and loading speed is a priority for its replacement.
- **Limited extensibility:** there is little room to expand functionality as monitoring requirements grow.

The existing system's response times and the technical causes of these limitations have not been established. This RFC does not attribute them to a particular language, database, or architecture.

Users still need the operational workflow: identify the client, database, and execution server; inspect report evidence; discuss an issue; and close or reopen tickets. Earlier assessments and closed tickets must remain available so users can investigate recurring problems.

Expansion is a concrete requirement. After the Oracle daily-check PoC, Zyra must accommodate SQL checks and the standby-alert flow. Its design should allow those additions while reusing the common ticket and history workflows.

## 3. Goals and scope

The PoC should demonstrate:

1. Correct interpretation of Oracle daily-check emails using each database's selected checks and thresholds.
2. A usable investigation workflow with comments, close/reopen actions, server context, and retained history.
3. Responsive loading, filtering, navigation, and client search.
4. A design that accommodates additional assessment types without substantial changes to unrelated functionality.
5. Self-hosted, Dockerized operation using the agreed project structure.

The PoC covers Oracle daily checks normally received in AM and PM windows. Release 1.0 adds SQL checks/pages and the standby-alert emails sent every 30 minutes. Their detailed report formats and rules still need to be established.

Direct database administration, automatic remediation, SaaS billing, and native mobile applications are outside the current documented scope.

The full functional scope is in [documentation.md](documentation.md#2-scope).

## 4. Agreed constraints

| Area | Agreed choice |
| --- | --- |
| Project | Zyra monorepo |
| API | Go under `zyra-api/` |
| Web application | React under `zyra-web/`, with source in `zyra-web/src/` |
| Development database | PostgreSQL |
| Hosting | Self-hosted and Dockerized |
| Authentication requirements | 15-minute access tokens, a fixed seven-day refresh session requiring sign-in again at expiry, and password recovery |
| Users | Admin, trusted, and normal roles |
| Appearance | Light theme by default, with a dark option |

The [documented folder structure](documentation.md#111-monorepo-layout) applies.

Access tokens can be refreshed without another sign-in during the seven-day session. Refreshing does not extend the deadline measured from the original sign-in. Access tokens issued near that deadline must expire no later than the deadline, when the user must sign in again.

Token storage and format, password hashing, refresh-token rotation, recovery implementation, mailbox access, file storage, production database setup, background processing, caching, queues, security, and operations remain undecided. The agreed choices do not select additional frameworks or infrastructure.

## 5. Proposed design

### 5.1 Separate ingestion, interpretation, and ticket handling

The application should separate the following responsibilities:

1. Receive and retain the email.
2. Match it to a configured database and assessment source.
3. Determine whether it contains a complete assessment.
4. Interpret selected report sections and evaluate their rules.
5. Record the assessment and produce the appropriate ticket findings.
6. Present those findings through the shared ticket workflow.

These are logical responsibilities within the application; their process or deployment arrangement remains open.

This separation is intended to let report formats change without rewriting the ticket interface, and thresholds change without altering mailbox access. The PoC should test whether these boundaries make additions manageable.

### 5.2 Receive existing reports through a dedicated mailbox

The proposed approach uses the existing reports: relevant emails are forwarded to a dedicated address, and Zyra acts as a client of that inbox. A configured sender and subject identify the assessment source, with report metadata providing further database and server context.

Unmatched or ambiguous messages should remain available for review. Receiving the same message again must not create duplicate assessments or tickets.

The provider, access method, polling or event mechanism, retry behaviour, and deduplication implementation remain open. Forwarding must be checked to establish which original sender and subject information is preserved.

### 5.3 Apply configured checks before creating issues

Email content does not determine ticket creation by itself. Zyra evaluates the checks selected for the database and applies the appropriate resource rules.

The supplied examples establish these behaviours:

| Input | Required outcome |
| --- | --- |
| Multiple failing backup rows and `UNDOTBS1` at `2.7%` free, with the relevant checks configured | One grouped Backups ticket and one Tablespace ticket |
| `Backups=NOT_US` | Backups are outside the company's responsibility; no Backups ticket |
| Script errors leave the assessment incomplete | A Missing Email issue classified as received but incomplete |
| A selected Archive Destinations check reports invalid destinations | An Archive Destinations issue with the report evidence |
| No email arrives in its configured window | A Missing Email issue classified as not received |
| A section is not selected/configured for evaluation | No ticket from that section; the raw content remains accessible |

`UNDOTBS1` is a tablespace resource, not a separate check category. The detailed rules and examples are in [Email matching and parsing](documentation.md#6-email-matching-and-parsing).

Grouping findings within one report is separate from handling repeated failures across reports. The single grouped Backups ticket is confirmed. Whether a later failing assessment updates an open ticket or creates a new one remains undecided.

### 5.4 Preserve assessments independently of ticket status

An assessment records the report and its evaluation. A ticket records the investigation. Closing a ticket must not turn its failed source assessment into a passed assessment.

The proposal retains original report content, relevant evidence, effective rule values, and the server identity available when the report was processed. This allows users to understand historical outcomes after configuration changes.

For the supplied format, the hostname can be extracted from `Run by: user@hostname`. How IP addresses are supplied or associated with hostnames remains open. A missing email has no observed execution host; any server context for that issue must come from known configuration.

Storage format, retention, and migration of historical data remain undecided.

### 5.5 Reuse common workflows for expansion

Client/database context, tickets, comments, permissions, and history should be reusable across assessment types. Report interpretation and expected-arrival rules should accommodate the differences between Oracle daily checks, SQL checks, and standby alerts.

The PoC should demonstrate where another assessment type fits. SQL and standby implementation belongs to Release 1.0, using their actual reports and requirements.

This introduces some design effort upfront. Its value depends on showing that a new type can be added without substantial changes to unrelated functionality.

### 5.6 Demonstrate responsiveness

Go and React alone do not establish that Zyra will be fast. The PoC should measure representative operations: dashboard loading, opening a ticket, filtering issues, searching clients, and browsing assessment history.

Expected data volumes, baseline measurements, target response times, and optimisation methods remain to be agreed. Performance improvement should be demonstrated against those measures.

## 6. Rationale and alternatives

These alternatives should inform review. No technical or commercial evaluation has yet established their feasibility.

| Option | Potential benefit | Limitation or question |
| --- | --- | --- |
| Continue with the existing system | Avoids replacement and transition work | Leaves the reported speed and expansion problems unresolved |
| Improve the existing system | May preserve familiar workflows and existing data | Depends on code access, maintainability, and the actual causes of its limitations |
| Adapt an existing monitoring or ticket product | May reduce the common functionality that must be built | Requires evaluation of parsing, configuration, hosting, history, customisation, and cost |
| Build Zyra | Allows the application to be designed around the required workflow and planned additions | Requires implementation, maintenance, and a transition plan |

Zyra is the proposal under consideration because its design can directly address the stated workflow and extension needs. The PoC should provide evidence that justifies continuing. We cannot yet claim that modifying the current system or adopting another product is infeasible.

## 7. Drawbacks and risks

**Parser correctness.** Format changes and incomplete output can cause false or missed issues. The supplied examples should become repeatable validation cases, with additional formats collected before wider use.

**Email dependency.** Delivery delays, forwarding changes, and mailbox failures affect monitoring. Arrival windows and the distinction between absent and incomplete reports need validation.

**Unproven performance.** A replacement could reproduce the current speed problems unless tested with representative data and user actions.

**Unproven extensibility.** Shared abstractions can add complexity without helping actual additions. SQL and standby samples are needed to test whether the proposed boundaries fit.

**History and transition.** Replacement raises questions about existing tickets, comments, and assessments. Export availability and migration requirements are unknown.

**Ongoing ownership.** A custom application requires maintenance of parsing, workflows, dependencies, and deployment. Ownership and operational arrangements remain open.

## 8. Validation and rollout proposal

1. **Establish a baseline:** identify slow operations, expected data volumes, and required expansion scenarios; agree success measures.
2. **Validate interpretation:** verify the supplied report cases, selected-check behaviour, grouped Backups, Tablespace classification, Missing Email, Archive Destinations, and `NOT_US`.
3. **Validate the workflow:** exercise configuration, history, comments, close/reopen, permissions, raw-email access, and hostname/IP display.
4. **Measure performance and extensibility:** test realistic list/history volumes and review the changes needed to support another assessment type.
5. **Review the PoC:** compare results with the agreed measures and record what needs revision before proceeding.
6. **Deliver Release 1.0 additions:** implement SQL checks/pages and the 30-minute standby flow, validating their specific formats and behaviour.

The [PoC acceptance criteria](documentation.md#16-proof-of-concept-acceptance-criteria) provide the functional checklist.

Production transition is not decided here. Whether Zyra runs alongside the current system first, whether history is imported, and how a pilot can be reverted must be agreed before cutover.

## 9. Open questions

| Area | Decision needed |
| --- | --- |
| Current limitations | Which operations are slow, and what specifically prevents expansion? Can the current system be modified? |
| Success measures | What response times, data volumes, and extension effort would make the PoC successful? |
| Mail access | Which provider/access method is used, and how does forwarding preserve matching information? |
| Report completeness | Which sections are required for each script version, and how are optional or unselected sections distinguished from missing output? |
| Schedules | What are the AM/PM windows, timezones, grace periods, and trusted-user editing permissions? |
| Missing Email correlation | How does an incomplete report relate to its expected window so the same run does not produce duplicate missing-email issues? |
| Repeated failures | Does a later failing assessment update an open ticket or create another? |
| Recovery | Does a late valid report close its Missing Email ticket or leave it for review? |
| Resource exclusions | How are ignored backup files handled within a grouped Backups issue? |
| Permissions and notes | Can normal users close/reopen tickets, and are notes ticket-specific, database-specific, or both? |
| Server identity | How are hostnames associated with IP addresses, and how are changes or unmatched hosts handled? |
| Authentication | What token format/storage, hashing, refresh-token rotation, and recovery mechanism will be used? |
| Storage and operations | How are emails/media stored, and what retention, production hosting, security, and operational requirements apply? |
| Transition | Is history imported, is a parallel pilot required, and what is the cutover/reversion plan? |
| Release 1.0 | Which SQL and standby formats, rules, and scheduling behaviours must be supported? |

## 10. Decision and review outcome

**Current status: draft; the overall design has not been approved.**

Review should establish whether this approach addresses the current system's speed and expansion problems, what evidence the PoC must provide, and which questions need resolution before implementation.

After review, record the decision, date, agreed changes, remaining questions, and follow-up proposals here. Detailed implementation decisions can be documented separately when enough information is available to assess them.

## References

- [Zyra product documentation](documentation.md) — detailed requirements, screens, folder layout, and email examples.
- [Rust RFC template](https://github.com/rust-lang/rfcs/blob/master/0000-template.md) — motivation, design, drawbacks, alternatives, and unresolved questions.
- [Design Docs at Google, Malte Ubl](https://www.industrialempathy.com/posts/design-docs-at-google/) — context, goals, design reasoning, and tradeoffs.
