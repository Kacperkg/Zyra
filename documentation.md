# Zyra

Zyra is a self-hosted monitoring and ticket-management application for Oracle database health-check emails. It receives assessment emails produced by existing server-side scripts, parses their contents, records every assessment, and raises actionable tickets when a check fails or an expected email is missing.

This document defines the initial product scope and proposed technical design. It is intended to guide an MVP; implementation has not started.

## 1. Goals

Zyra should:

- turn Oracle health-check emails into structured assessments;
- identify failures such as backup, FRA, archive destination, filesystem, ASM, tablespace, and malformed or missing-email problems;
- create tickets only for enabled checks that breach the configured rules;
- preserve the full history of passed and failed assessments;
- let users investigate, discuss, close, and reopen tickets;
- provide fast navigation across clients, databases, assessments, and tickets;
- enforce role-based permissions;
- run as a self-hosted, Dockerized Go and React monorepo.

## 2. Scope

### 2.1 MVP scope

- Oracle daily-check emails, normally expected twice a day (AM and PM).
- Email ingestion into a dedicated Zyra mailbox.
- Subject/sender-based matching of emails to a configured database.
- Storage of the original email and parsed assessment result.
- Configurable check enablement and thresholds per database.
- Detection of malformed/incomplete assessment emails.
- Detection of an email not arriving within a configured schedule window.
- Ticket creation, commenting, closing, and reopening.
- Clients, databases, email-source configuration, schedules, assessments, users, and profiles.
- Admin, trusted, and normal-user roles.
- Light theme by default, with a dark-theme option.
- Docker-based self-hosting.

### 2.2 Explicitly out of scope for the MVP

- SQL checks and SQL issue pages.
- The separate standby-alert email that currently runs every 30 minutes.
- Sending commands to, or making changes on, monitored database servers.
- Automatically fixing database problems.
- SaaS/multi-tenant hosting and billing.
- Native mobile applications.

The data model and parser interfaces should still be extensible so SQL and standby assessments can be added later without redesigning tickets or assessment history.

## 3. Terminology

| Term | Meaning |
| --- | --- |
| Client | A customer or organisation that owns one or more databases. |
| Database | A monitored Oracle database belonging to a client. |
| Email source | Rules used to associate an incoming email with a database, including sender and subject. |
| Assessment | One received and processed health-check email, including its raw source and parsed results. |
| Check result | The outcome of one check within an assessment, such as Filesystem or Backups. |
| Ticket | An actionable issue created from a failed check, malformed email, or missing expected email. |
| Daily check | The current Oracle health assessment, normally received in AM and PM windows. |
| Missing email | Either no email arrived during an expected window, or an email arrived but could not be treated as a complete valid assessment. |

## 4. Users and permissions

Permissions must be enforced by the backend, not only hidden in the UI.

| Capability | Normal | Trusted | Admin |
| --- | :---: | :---: | :---: |
| View dashboard, clients, databases, assessments, and tickets | Yes | Yes | Yes |
| Search and filter | Yes | Yes | Yes |
| Comment on tickets | Yes | Yes | Yes |
| Close and reopen tickets | Yes | Yes | Yes |
| Edit own profile and profile picture | Yes | Yes | Yes |
| Add and edit clients/databases | No | Yes | Yes |
| Edit email sources, schedules, enabled checks, and thresholds | No | Yes | Yes |
| Delete clients/databases/configuration | No | No | Yes |
| Manage users and roles | No | No | Yes |
| View another user's ticket-closure activity | No | No | Yes |

Deletion should preferably be a reversible archive/soft-delete operation so historical assessments and tickets remain intact.

## 5. Primary workflow

1. An existing script runs an Oracle database health check.
2. The script emails its report to the existing operational address.
3. A mail rule forwards relevant messages to a dedicated Zyra mailbox. This isolates ingestion from the human inbox and follows the recommended client/event model.
4. A Zyra email-ingestion worker connects to that mailbox, retrieves the message, and stores the raw email before parsing it.
5. Zyra matches the sender and normalized subject to a configured database and assessment type.
6. The appropriate versioned parser converts the body into structured check results.
7. Zyra evaluates enabled checks and database-specific thresholds.
8. The assessment and all check results are stored whether they passed or failed.
9. Failed results create or update tickets according to the deduplication policy.
10. Users investigate the ticket, leave comments, and close it. A closed ticket can be reopened at any time.

### 5.1 Recommended ingestion approach

For the MVP, Zyra should act as a mailbox client using a dedicated mailbox and a provider-supported protocol/API (for example IMAP with TLS or Microsoft Graph, depending on the chosen mail provider). Polling is acceptable initially. The ingestion interface should be provider-neutral so webhooks or a different mail provider can be added later.

Each message must be idempotent. Zyra should store the provider message ID and a deterministic content hash, and must not create a second assessment or ticket when the same email is forwarded or fetched twice.

Processing should use durable jobs with retry and a dead-letter state. A transient parsing or mailbox failure must not silently lose an email.

## 6. Email matching and parsing

### 6.1 Email-source configuration

Each database may have one or more assessment email sources containing:

- expected sender address;
- expected subject or an explicit safe matching pattern;
- assessment type, initially `daily_check`;
- parser type and parser version;
- timezone;
- enabled/disabled state;
- one or more expected schedule windows;
- optional grace period.

Example subject: `XXXX_XXX IFSPRD - Daily Check`.

Matching should normalize harmless differences such as leading/trailing whitespace and repeated forwarding prefixes (`FW:`, `Fwd:`), but should not use broad fuzzy matching that could assign a report to the wrong database.

Unmatched email should be retained in an ingestion-review queue for an administrator; it should never be discarded.

### 6.2 Parser contract

A parser receives the immutable raw message and returns:

- database identity discovered in the report, if present;
- execution hostname and IP address, if present or resolved through configured server metadata;
- assessment timestamp and type;
- parser version;
- overall parsing status;
- a collection of normalized check results;
- warnings for missing, truncated, duplicated, or unrecognised sections.

Each check result contains:

- stable check type;
- display title;
- result: `passed`, `failed`, `warning`, `ignored`, `not_evaluated`, or `unknown`;
- measured value(s) and unit, where applicable;
- configured threshold used for evaluation;
- short summary suitable for a ticket title;
- relevant raw excerpt or structured evidence.

The parser extracts facts; the rule evaluator decides whether those facts create a ticket. Keeping these separate makes parser testing and future rule changes safer.

The daily-check parser must also handle the format shown in the first supplied anonymised email sample:

- mail-system warning banners may appear before the report and must be ignored as transport content, not interpreted as part of the assessment;
- `ReportOn`, `PkgVersion`, `UniqueID`, `Database`, database version, and script information are report metadata;
- checks are introduced by stable labels such as `Datafiles=`, `Backups=`, and `Tablespaces=`;
- `OK!` means the report found no issue for that check, but an enabled/ignored rule must still be applied and stored;
- multi-line sections use trailing `\` characters as formatting/continuation markers; these must not become part of parsed values;
- tabular sections must be parsed into individual resource findings rather than stored only as one block of text;
- the `=@=` marker terminates the database-statistics block before the filesystem section;
- Windows paths, spacing, case differences, old Oracle versions, and forwarded email formatting must be preserved or normalized safely without breaking section recognition;
- `ReportOn` is the assessment time. The received-at time is stored separately.

For this report format, `Script Info` identifies where the check ran. From a value such as `Run by: username@hostname`, the parser must extract only the portion after the final `@` as the observed execution hostname; the username is not part of the connection target. Zyra then matches the observed hostname to a configured database server record containing a canonical hostname and optional IP address. Both the observed value and the matched server identity must be stored. If no server matches, the assessment remains processable but shows an explicit `Unmatched host` warning for administrative review rather than silently attaching an incorrect IP.

Zyra should not depend on live DNS lookup when a user opens a ticket. Hostname and IP displayed on an assessment or ticket should be immutable snapshots taken when the assessment is processed, with links to the current server configuration. This ensures an older ticket still shows the original connection target after a server migration or IP change.

An external-email caution banner is untrusted message content. It is neither an instruction to Zyra nor evidence of a database failure.

### 6.3 Initial check catalogue

The initial Oracle catalogue should support the checks visible in the reference system and the requested core checks:

- ASM space;
- archive destinations;
- backups;
- clusterware;
- datafiles;
- extents;
- failed jobs;
- filesystem;
- indexes;
- invalid objects;
- max lag;
- recovery area/FRA space;
- segments;
- tablespaces;
- malformed/incomplete email;
- missing expected email.

Check identifiers should be machine-friendly stable values such as `archive_destinations` and `filesystem`; display names can change without breaking history.

### 6.4 Worked parsing example

The first anonymised sample is a daily check produced by package version `2.5` for database `ifsprd`. It contains, among other data:

```text
Backups=
Datafiles needing backup:
RMAN  E:\ORADATA\IFSPRD\APEX01.DBF  22-SEP-2026 00:42:16
...

Tablespaces=
Name       Total MB  Max. MB  Used MB  %Free
UNDOTBS1   31744     31744    30900     2.7%
```

As shown by the selected/excluded checks reference screen, each database controls which sections of the email Zyra evaluates. For this database, Backups and Tablespaces are selected for evaluation. The same email therefore produces exactly two tickets:

| Ticket | Parsed evidence | Ticketing rule |
| --- | --- | --- |
| Backups | The `Backups` section contains multiple datafiles needing backup and their last-completed timestamps. | The Backups check is selected and the non-empty failure list breaches its rule. Create **one grouped Backups ticket** for the assessment, with every affected file attached as structured evidence. Never create one ticket per datafile from this section. |
| Tablespace (`UNDOTBS1`) | `UNDOTBS1` has `2.7%` free in the `Tablespaces` section. | This is a Tablespace check. The configured minimum-free threshold is breached, so create one Tablespace ticket for resource `UNDOTBS1`. `UNDO` is part of the affected tablespace's name, not a separate check type. |

The email also contains unusable indexes and filesystem usage values, including drive `E:` at `91%`. Indexes and Filesystem are **not selected/configured** for this database example, so their sections are not evaluated and cannot create tickets. Their contents remain available in the immutable raw email; Zyra may record only that the sections were skipped as `not_evaluated`. Once an administrator or trusted user selects and configures those checks, later assessments can evaluate them. Sections containing `OK!` count as passed only when that check is selected for evaluation. Therefore, the presence of a non-empty section alone is not a universal ticket rule.

In summary, the expected outcome for this assessment is exactly:

1. One Tablespace ticket for `UNDOTBS1` at `2.7%` free.
2. One Backups ticket containing the complete list of affected datafiles.

No additional ticket is created for each backup row, and no Indexes or Filesystem ticket is created because those checks are not selected for this database.

The expected high-level parser output is:

```json
{
  "assessmentType": "daily_check",
  "database": "ifsprd",
  "reportTime": "2026-09-23T11:27:59",
  "packageVersion": "2.5",
  "ticketCandidates": [
    { "checkType": "backups", "resources": "parsed from all backup rows" },
    { "checkType": "tablespace", "resource": "UNDOTBS1", "freePercent": 2.7 }
  ],
  "ticketsCreated": 2
}
```

The Backups candidate above is an abbreviated representation; the actual output contains the structured method, path, and last-completed time from every parsed row. Database timezone must be applied before converting `ReportOn` to a UTC timestamp.

### 6.5 Missing and malformed emails

Zyra must distinguish:

1. **Not received:** no matching valid email arrived during the configured window plus grace period.
2. **Received but malformed:** an email arrived, but required sections were absent, truncated, or unparsable.

The missing-email scheduler evaluates every enabled email source in its configured timezone. A schedule can define separate AM and PM windows, for example `18:30–22:00`. A valid assessment satisfies only the appropriate window. Each missed window creates at most one ticket.

If a delayed valid email arrives after a missing-email ticket was created, Zyra should link the assessment to that ticket and mark it as no longer missing. Automatic closure is a product decision still to be confirmed; the safer MVP default is to leave the ticket open for a user to review.

## 7. Check configuration and thresholds

Configuration is per database. Each known check is selected or excluded, following the included/excluded interaction shown in the reference screen. Only selected checks are evaluated against their rules and can create tickets. An excluded or not-yet-configured section is skipped; its content remains available in the raw email but does not need to be parsed into resource-level results.

Initial rule shapes include:

| Check | Configuration |
| --- | --- |
| Filesystem | Filesystem/mount name, enabled or ignored, maximum usage percentage. |
| Tablespace | Tablespace name, enabled or ignored, minimum required free percentage. |
| ASM space | Disk group/name, enabled or ignored, minimum required free percentage. |
| Backup | Backup target/name, enabled or ignored, and parser-specific success criteria. |
| Archive destinations | Destination, enabled or ignored, and acceptable status. |
| FRA/recovery area | Enabled or ignored and maximum usage/minimum free threshold. |
| Missing email | Expected schedule window, timezone, grace period, enabled or ignored. |

Checks or resources without matching configuration should be skipped as `not_evaluated` and may be displayed for configuration review. Zyra must not silently choose a threshold or create a ticket for an unconfigured check, mount, tablespace, disk group, or backup target.

All configuration changes should be audited with actor, timestamp, before/after values, and an optional reason. An assessment must retain the effective rule snapshot used to calculate its result so historical outcomes do not change when thresholds are edited later.

## 8. Tickets

### 8.1 Ticket fields

- unique numeric ticket number;
- status: `open` or `closed` initially;
- check/issue type;
- assessment type;
- title and summary;
- client, database, and execution server snapshot, including observed hostname, canonical hostname, and IP address when configured;
- source assessment and check result, where applicable;
- created and updated timestamps;
- closed timestamp and closing user;
- reopen timestamp/history;
- comments and participants;
- links to up to five recent similar issues.

Suggested title format: `Backups check failed for <client> / <database> – Daily Check`.

For the worked email example, suitable titles are:

- `Backups check failed for <client> / IFSPRD – Daily Check`
- `Tablespace check failed for <client> / IFSPRD – UNDOTBS1 at 2.7% free`

### 8.2 Creation and deduplication

The exact repeat-failure policy must be configurable or confirmed before implementation. Recommended MVP behaviour:

- create one open ticket per database, assessment type, check type, and resource identity;
- when the same failure occurs while that ticket is open, attach the new occurrence to it and update `last seen` and occurrence count;
- when the previous ticket is closed and the failure happens again, create a new ticket and link it as a similar issue;
- never merge unrelated resources (for example two different filesystems) solely because their check type is the same.

### 8.3 Ticket list

The Oracle issues page lists both current and historical tickets and supports:

- filters for client, database, ticket/check type, created date, assessment type, and status;
- sorting by ticket number, check type, created date/time, client, database, and assessment type;
- pagination;
- a clear empty state and loading state;
- persistent filters in the URL so a view can be bookmarked or shared.

Default column order:

1. Ticket number
2. Check type
3. Created date
4. Created time
5. Client/customer
6. Database
7. Assessment type
8. Status

Closed tickets remain searchable and visible.

### 8.4 Ticket detail

The ticket page should show:

- issue title, status, ticket number, client, database, execution hostname, IP address, and timestamps;
- a link to the source assessment and a read-only view of the sanitized raw email;
- parsed evidence and all linked occurrences;
- chronological system events and user comments;
- `Comment` and `Comment and close` actions;
- a `Reopen` action for closed tickets;
- a right-hand context panel containing notes, linked client, linked database, participants, and the five most recent similar issues.

Comments should support a controlled rich-text subset: paragraphs, headings/body sizes, bold, italic, underline, lists, alignment, links, text colour, images, and GIFs. Content must be sanitized on the server. Uploaded files require type/size limits and should be served from authenticated storage; arbitrary embedded HTML or JavaScript is not allowed.

Every status change is an immutable timeline event showing who performed it and when.

## 9. Screens and navigation

### 9.1 Shared navigation

The same responsive navigation appears throughout the application and includes:

- dashboard;
- Oracle issues;
- clients;
- databases/search;
- user profile and theme control;
- administration links when authorised.

### 9.2 Dashboard

The MVP dashboard contains an `Open Oracle Issues` summary card with the current count and a link to the filtered Oracle issues page.

A future `Open SQL Issues` card may use the same component but is not part of the MVP and should not imply that SQL monitoring is active.

### 9.3 Client page

- client details;
- searchable list of configured databases;
- links to each database;
- add/edit actions for trusted users and admins;
- archive/delete action for admins only.

### 9.4 Database page

- database identity, configured server hostname(s), IP address(es), client, and status;
- enabled/excluded check configuration;
- check-specific threshold tables;
- assessment email sources and schedules;
- assessment history with date, type, and passed/failed result;
- filters for assessment history;
- participants who have commented on this database's tickets;
- five recent similar/relevant issues.

Selecting an assessment email source opens a detail page or panel where trusted users and admins can edit its sender, subject rule, parser, and expected schedules.

### 9.5 User profile and administration

Users can change their own display details, password, theme, and profile picture. Admins can list users, manage roles/status, and view the tickets closed by a selected user.

## 10. Authentication and account recovery

- Short-lived signed access tokens with a target lifetime of 24 hours.
- Refresh tokens with a target lifetime of 7 days.
- Refresh-token rotation on every use, with reuse detection and server-side revocation.
- Refresh tokens stored in `Secure`, `HttpOnly`, `SameSite` cookies; do not store them in browser local storage.
- Passwords hashed with Argon2id using deployment-appropriate parameters.
- Rate limiting for login, refresh, and password-recovery endpoints.
- Password recovery using a single-use, hashed, time-limited token delivered by email.
- Password reset revokes all existing refresh sessions for that user.
- Admins can disable an account and revoke its sessions.
- Authentication and security-sensitive actions are audit logged.

The 24-hour access-token lifetime is a requested starting point, but it increases the window in which a stolen access token remains useful. Before production, consider a 15–60 minute access token while retaining the 7-day rotating refresh session.

## 11. Proposed architecture

### 11.1 Monorepo layout

```text
Zyra/
├── apps/
│   ├── api/               # Go HTTP API and background workers
│   └── web/               # React frontend
├── internal/              # Go domain/application packages
│   ├── auth/
│   ├── assessments/
│   ├── email/
│   ├── parsers/
│   ├── rules/
│   └── tickets/
├── packages/              # Shared frontend packages/types if needed
├── migrations/            # Database migrations
├── deploy/                # Docker and self-hosting configuration
├── docs/                  # Future detailed design/operations documents
└── documentation.md
```

This is a proposed layout, not a fixed implementation constraint.

### 11.2 Components

- **React web app:** TypeScript single-page application, responsive UI, route-level code splitting, accessible components, light/dark themes.
- **Go API:** JSON API, authorization, configuration, tickets, comments, search, uploads, and reporting.
- **Worker processes:** mailbox polling, parsing, ticket generation, and missing-email schedule evaluation. They may initially run from the same Go binary in separate process modes.
- **PostgreSQL:** primary transactional store and initial full-text/trigram search.
- **Object storage:** raw MIME messages, sanitized rendered bodies, profile images, and comment attachments. An S3-compatible store allows self-hosted MinIO or external S3; local filesystem storage can be an MVP adapter for single-node installs.
- **Optional Redis:** deferred until measurements justify it; useful later for distributed jobs, caching, or rate limiting.

### 11.3 Suggested runtime flow

```text
Mail server
    │
    ▼
Mailbox ingestion worker ──► Raw immutable email storage
    │
    ▼
Database/source matcher ──► Versioned parser ──► Rule evaluator
                                                   │
                          ┌────────────────────────┴───────────────┐
                          ▼                                        ▼
                  Assessment history                         Ticket service
                                                                   │
                                                                   ▼
                                                           React application
```

## 12. Core data model

The following entities are expected; fields may be refined during implementation.

- `users`, `roles`, `user_roles`, `refresh_sessions`, `password_reset_tokens`
- `clients`
- `databases`
- `database_servers` (canonical hostname, optional IP address, aliases, active state, and connection/display notes)
- `assessment_email_sources`
- `assessment_schedules`
- `check_definitions`
- `database_check_settings`
- `resource_thresholds`
- `raw_emails`
- `assessments`
- `assessment_check_results`
- `tickets`
- `ticket_occurrences`
- `ticket_comments`
- `ticket_events`
- `attachments`
- `audit_events`

Important integrity rules:

- client/database names may repeat globally but should be unique within their appropriate parent scope;
- a raw email provider ID and content hash support idempotency;
- an expected schedule window can produce at most one missing-email occurrence;
- an assessment stores the observed execution hostname plus the matched server hostname/IP snapshot used by its tickets;
- assessments and check results are append-only operational history;
- tickets and comments must not be cascade-deleted when a user is disabled;
- stored timestamps use UTC, while schedules retain an IANA timezone such as `Europe/London`.

## 13. API outline

The API should be versioned under `/api/v1`. A possible first surface is:

```text
POST   /auth/login
POST   /auth/refresh
POST   /auth/logout
POST   /auth/forgot-password
POST   /auth/reset-password

GET    /dashboard
GET    /tickets
GET    /tickets/{ticketId}
POST   /tickets/{ticketId}/comments
POST   /tickets/{ticketId}/close
POST   /tickets/{ticketId}/reopen

GET    /clients
POST   /clients
GET    /clients/{clientId}
PATCH  /clients/{clientId}

GET    /databases
POST   /databases
GET    /databases/{databaseId}
PATCH  /databases/{databaseId}
GET    /databases/{databaseId}/assessments
GET    /databases/{databaseId}/check-settings
PUT    /databases/{databaseId}/check-settings

GET    /assessments/{assessmentId}
GET    /assessments/{assessmentId}/raw-email

GET    /users/me
PATCH  /users/me
POST   /users/me/avatar
GET    /admin/users
GET    /admin/users/{userId}/closed-tickets
```

List endpoints must use server-side filtering, stable sorting, and cursor or indexed page-based pagination.

## 14. Security and privacy requirements

- TLS is required outside local development.
- Raw email content is untrusted input and must never be executed.
- HTML email and rich-text comments must be sanitized with an allowlist; external images should be blocked or proxied to avoid tracking.
- Raw-email access requires authorization and should redact or safely render dangerous content.
- Validate attachment content type, extension, size, and image dimensions.
- Apply CSRF protection where cookie authentication is used.
- Use strict CORS, security headers, and a Content Security Policy.
- Do not write email bodies, passwords, tokens, or sensitive attachment data to application logs.
- Store secrets outside images and source control, using environment variables or Docker secrets.
- Encrypt backups and document restore procedures.
- Audit configuration, permission, login, ticket-status, and administrative changes.
- Define a retention policy for raw emails, attachments, audit records, and user data before production.

## 15. Performance and usability

Loading speed is a primary requirement. Initial targets:

- API p95 under 300 ms for common cached/indexed reads under normal internal load;
- useful page content visible within 2 seconds on a typical business connection;
- ticket/client/database search feedback within 300 ms after debounce;
- no unbounded list or timeline queries;
- dashboard counts calculated with indexed queries or refreshed summary data;
- database indexes designed around ticket status, client, database, check type, assessment type, and created time;
- route-level frontend code splitting and lazy loading for the rich-text editor;
- responsive layouts and keyboard-accessible controls;
- WCAG 2.1 AA colour contrast in both themes.

Search should start with PostgreSQL indexes and trigram/full-text search. A separate search service should be introduced only if real usage demonstrates the need.

## 16. Observability and operations

- Structured application logs with correlation IDs.
- Metrics for mailbox connection, ingestion lag, messages received, unmatched messages, parse failures, assessments, tickets created, missing windows, and job retries.
- Health and readiness endpoints for containers.
- Alert when mailbox ingestion has not succeeded within a configured interval.
- Docker Compose development/self-hosting setup with persistent volumes.
- Automated database migrations and documented backup/restore process.
- Graceful shutdown so in-progress jobs can be retried safely.

## 17. Testing strategy

- Unit tests for every parser section and rule evaluator.
- Golden-file tests using anonymised real email fixtures.
- A golden-file test for the supplied package-version `2.5` daily check that asserts two tickets only: Backups and Tablespace (`UNDOTBS1`).
- Assertions that the same fixture skips Indexes and Filesystem as `not_evaluated`, without resource-level evaluation or tickets, because those checks are not selected/configured in this example.
- An assertion that all backup rows are grouped into one Backups ticket rather than creating a ticket for each affected datafile.
- Assertions that the external-email caution banner, line-continuation characters, and `=@=` delimiter do not corrupt report metadata or check sections.
- Assertions that `Run by: user@hostname` produces the observed hostname, matches the correct configured database server, and snapshots its canonical hostname/IP onto both created tickets.
- Tests for truncated, reordered, duplicated, forwarded, HTML-only, and unexpected email bodies.
- Schedule tests across timezones, daylight-saving changes, grace periods, and late arrivals.
- Idempotency and concurrency tests for duplicated messages and workers.
- Permission tests for every protected backend operation.
- Integration tests using PostgreSQL and a test mailbox/email fixture source.
- End-to-end tests for login, ticket filtering, commenting, close/reopen, database configuration, and password recovery.
- Security tests for XSS in email/comments, malicious attachments, refresh-token reuse, and access to another user's restricted operation.

Production parser fixtures must be anonymised and must not contain client credentials or sensitive database information.

## 18. MVP acceptance criteria

The MVP is ready for an internal pilot when:

1. An authorised user can sign in, refresh a session, log out, and recover a password.
2. Admins can manage users/roles; trusted users can configure clients and databases without deleting them; normal users have read-only configuration access.
3. A configured daily-check email is ingested exactly once, matched to the correct database, and stored with its raw source.
4. Supported checks are parsed into durable results and evaluated using the database's effective configuration.
5. Passed and failed assessments remain visible in history.
6. Enabled failures create correctly titled tickets; ignored checks do not.
   The supplied anonymised example creates exactly one Backups ticket and one Tablespace ticket for `UNDOTBS1`. Its unselected Indexes and Filesystem sections are skipped as `not_evaluated`; their content remains available in the raw email, and they create no tickets.
7. A missing or malformed email creates the appropriate ticket only once per expected window/message.
8. Users can filter and sort Oracle tickets, view an issue timeline, comment, comment-and-close, close, and reopen.
9. Ticket pages link to the client, database, source assessment/raw email, participants, and five similar issues.
   They also show the execution hostname and configured IP address captured when the assessment was processed.
10. Admins can see which tickets a user closed.
11. Light and dark themes work, with light as the default, and users can update their profile picture.
12. The application starts through documented Docker commands, persists its data, exposes health checks, and has a tested restore procedure.
13. Permission, parser, schedule, and core end-to-end tests pass.

## 19. Proposed delivery phases

### Phase 0 — Discovery and fixtures

- Collect anonymised examples of successful and failed daily-check emails.
- Confirm mail provider and supported access method.
- Confirm exact subject/sender patterns and AM/PM schedules.
- Define the first parser grammar and check catalogue.

### Phase 1 — Foundation

- Monorepo, Docker development environment, PostgreSQL migrations.
- Authentication, refresh sessions, password recovery, roles, users, and audit events.
- Client/database configuration and basic responsive application shell.

### Phase 2 — Assessment pipeline

- Dedicated mailbox ingestion, immutable raw storage, matching, idempotency, and retry states.
- Daily-check parser, rule evaluation, check settings, thresholds, and assessment history.
- Missing-email scheduler.

### Phase 3 — Ticket workflow

- Dashboard, Oracle issue list, filters/sorting, ticket detail, comments, close/reopen, participants, and similar issues.
- Search, profile pictures, theme preference, and admin closure reporting.

### Phase 4 — Hardening and pilot

- Security review, performance measurements, accessibility pass, operational metrics, backups/restores, parser-fixture expansion, and internal pilot.

## 20. Decisions needed before implementation

1. Which mail provider hosts the dedicated inbox, and should the MVP use IMAP, Microsoft Graph, Gmail API, or another supported API?
2. What are the exact AM/PM schedule windows, timezone, and allowed grace periods for each database?
3. What real email formats and script versions must the first parser support?
4. Which checks are mandatory for the first pilot, and what are their precise pass/fail rules?
5. Should a late valid email automatically close its missing-email ticket or only add a recovery event for manual review?
6. Should repeated failures update one open ticket (the recommendation here) or create a ticket per assessment?
7. Can normal users close/reopen tickets, or should that be limited to trusted users and admins?
8. Are client/database notes global notes, ticket-specific notes, or both?
9. What retention period and access rules apply to raw emails and attachments?
10. Is single-node local storage sufficient initially, or is S3-compatible object storage required from day one?

## 21. Future extensions

- SQL assessment checks and SQL issue dashboard/list.
- Thirty-minute standby-alert ingestion and dedicated alert behaviour.
- Additional mail providers and push/webhook ingestion.
- Notifications and escalation rules.
- Service-level reporting and trend analytics.
- Parser plug-ins/version migration tools.
- External ticketing or chat integrations.

---

The two supplied screenshots are treated as interaction references rather than designs to reproduce exactly. The useful concepts retained here are per-database selected/excluded checks, a chronological ticket conversation, prominent status, comment-and-close behaviour, linked client/database context, participants, and recent similar issues. Zyra's final UI should modernise these patterns while preserving their operational clarity.
