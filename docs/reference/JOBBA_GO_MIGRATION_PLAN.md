# Jobba: Go conversion, defect fixes, app integration, and verification

Date: 23 October 2026
Status: implementation complete; awaiting environment-specific staging rehearsal and authorized cutover
Workspace: `/Users/yoda/Desktop/jobba`
Architecture reference: `/Users/yoda/Downloads/BACKEND_BLUEPRINT.md`

## 0. Final implementation status

The implementation described by this plan is complete. `Jobba-server-go` is a feature-organized modular monolith with PostgreSQL migrations and repositories for identity, profiles, skills, jobs, applications, marketplace offers/reviews/saved workers, payments, chat, notifications, disputes, and administration. Every legacy Node HTTP route has a Go registration and handler; the accepted differences are documented in `FINAL_PARITY_AUDIT.md`. The Expo application now uses typed API services and real loading, empty, error, and success states for the connected marketplace flows.

Final verification passed on 23 October 2026: the full Go race suite, vet, and build; live PostgreSQL migration application and complete reverse migration; a healthy non-root Docker image; unauthenticated SSE rejection with HTTP 401; 16 Node Jest tests, eight focused Node regression tests, and the Node production build; and mobile TypeScript, lint, 33 endpoint-contract checks, auth-contract checks, and Expo export of 69 routes.

This status is not a production deployment declaration. The remaining work requires deployment-specific authority and external systems: provision production secrets and TLS/database/SMTP/Google/Paystack accounts, rehearse the importer against a protected copy of real data, reconcile the result, exercise provider sandbox behavior, configure ingress/monitoring/backups, and perform the staged cutover in `CUTOVER_ROLLBACK_RUNBOOK.md`. No production deployment, live provider transaction, or production-data import is claimed.

Earlier sections remain as the original source-grounded plan and historical checkpoints. Where an earlier statement says a feature or runtime check was pending, this final status and `FINAL_COMPLETION_REPORT.md` supersede it.

## 1. What this plan now establishes

Jobba is currently an Express/Mongoose backend with Socket.IO, plus an Expo/React Native application. The supplied Go blueprint targets a feature-organized modular monolith using sqlc, SQL migrations, repository wrappers, service interfaces, and zap.

This is therefore three coordinated workstreams:

1. Convert the backend to Go and migrate MongoDB data/queries to PostgreSQL.
2. Correct confirmed logic, authorization, payment, and resource-bounding defects.
3. Repair the mobile authentication/startup flow and connect unfinished mobile features to verified APIs.

GoLand can be the development IDE; Go is the target language.

The architecture document supplies design requirements. Statements in existing documentation are not proof that a feature works or is missing. In particular, `Jobba-server/DOCUMENTATION_AND_GUIDELINES.md` says applications, messaging, payments, reviews, and notifications are missing, but their models/routes/services now exist. Use the code and subsequent tests as the baseline.

### Inspection baseline

| Item | Observed state |
|---|---|
| Backend repository | `Jobba-server`, commit `0a29a71`; clean at inspection |
| Mobile repository | `Jobba-App`, commit `759a4cb`; clean at inspection |
| Parent folder | Not itself a Git repository |
| Backend stack | Express 5, Mongoose 8, Socket.IO 4, Joi, JWT, bcrypt, Google identity, Resend, Paystack |
| Mobile stack | Expo 54, React Native, Expo Router, Zustand, SecureStore, Axios |
| Installed application dependencies | Neither repository contains `node_modules` |
| Existing Go implementation | None identified in the inspected workspace |
| Test execution | Full suites not run; dependencies and isolated test services are not provisioned |
| Planning checks performed | Source inspection, route/model searches, Git status, Jest path checks, isolated JavaScript expression reproduction |

Three agents reviewed architecture/inventory, loops/correctness, and app/verification. Findings below distinguish code-confirmed defects from runtime measurements still required. Severity indicates implementation priority; no claim is made that an issue has been exploited in a deployed environment.

## 2. Immediate priorities

Address these before presenting the application as ready for real users:

1. Restrict privileged account creation and enforce conversation membership on real-time events.
2. Make payment processing and payouts safe under concurrent calls, retries, and ambiguous provider responses.
3. Fix mobile signup contracts, authentication persistence, startup hooks, and logout.
4. Repair the test harness and establish reproducible critical-flow tests.
5. Build the Go foundation, then migrate bounded feature slices with independent verification.
6. Complete mobile feature integration as corresponding APIs become stable.

Do not defer critical correctness fixes until the entire conversion is finished if the current service is serving users. The first implementation checkpoint must establish whether this is a development-only project or a live system; that determines whether urgent Node fixes precede the Go rollout.

## 3. Current feature inventory and Go ownership

All existing REST paths below sit under `/api/v1`.

| Existing capability | Source | Proposed Go feature | Migration notes |
|---|---|---|---|
| Signup, signin, Google identity, verification, reset/change/set password | `src/routes/auth.js`, `src/services/auth.js` | `internal/auth` | Correct public role selection; preserve bcrypt verification; plan session transition |
| User/profile and account status | `src/models/user.js`, auth/profile handlers, admin user service | `internal/user` | Preserve public IDs and status semantics; explicit response DTOs |
| Roles and permissions | `src/routes/role.js`, `src/services/role.js`, `src/models/role.js` | `internal/role` | Separate account type from role ID and permission grants |
| Skill catalogue and skill administration | `src/routes/skill.js`, `src/routes/admin/skill.js` | `internal/skill` | Suitable low-risk read pilot |
| Worker profiles, directory, saved workers | `src/services/workerProfile.js`, `src/services/job/client/worker.js` | `internal/worker` | Geospatial search and user-skill relationships need explicit SQL mapping |
| Client jobs, statistics, lifecycle | `src/services/job/client/index.js` | `internal/job` | Define allowed transitions and atomic assignment |
| Worker search, recommendations, bookmarks | `src/services/job/worker/index.js`, saved-job model | `internal/job` | Replace unbounded exclusion materialization; bounded pagination |
| Applications and selection | `src/services/job/application.js`, application model | `internal/application` | Preserve uniqueness of job/worker; coordinate acceptance with job state |
| Client offers and worker accept/reject | `src/services/job/client/offer.js`, `src/services/job/worker/offer.js` | `internal/offer` | Coordinate offer, application, job, and chat effects |
| Conversations and messages | `src/routes/chat.js`, `src/services/chat.js`, `src/websocket/index.js` | `internal/chat` | Preserve or deliberately replace Socket.IO protocol |
| Notifications | `src/routes/notification.js`, notification service/model | `internal/notification` | Durable state separate from best-effort live delivery |
| Checkout, escrow, payout, bank accounts | `src/routes/payment.js`, `src/services/payment.js`, `src/utils/payment/index.js` | `internal/payment` | Highest-risk migration; integer money, idempotency, reconciliation |
| Reviews and ratings | `src/controllers/review.js`, `src/services/review.js` | `internal/review` | Routes are nested under job routers; preserve reviewer/job uniqueness |
| User/admin disputes | `src/routes/dispute.js`, `src/routes/admin/dispute.js` | `internal/dispute` | Reconcile job disputes with payment disputes and settlement |
| Email and reverse geocoding | `src/services/email.js`, `src/services/location.js` | `internal/email`, `internal/location` adapters | Deadlines, controlled retries, no sensitive payload logging |

These are package responsibilities, not separate services. The target remains one deployable Go application. Avoid importing another feature's concrete repository or HTTP handler; expose narrow consumer-owned contracts and document cross-feature transaction boundaries.

## 4. Confirmed findings and fix backlog

Paths are relative to the workspace. Line references identify the inspected baseline and may move during implementation.

### A. Security, payment, and server correctness

| ID / priority | Evidence | Consequence | Planned fix and acceptance |
|---|---|---|---|
| SEC-01 / P0 | `Jobba-server/src/routes/auth.js:7`; `src/controllers/auth.js:34–76`; `src/services/auth.js:47–52` | Public signup accepts `role: admin`, creates an admin account type, and attaches a matching role when present | Public signup accepts only client/worker. Provision admins through an authenticated, audited path. Anonymous admin signup must fail without creating a privileged user |
| SEC-02 / P0 | `Jobba-server/src/websocket/index.js:91–100` | Any authenticated socket can request membership in an arbitrary conversation room; broadcast message contents can reach a nonparticipant who knows the ID | Check active account and membership before join and on sensitive events. Test a third user joining, typing, reading, receiving, and spoofing receipts |
| PAY-01 / P0 | `Jobba-server/src/services/payment.js:151–191`, `:203–258`, `:275–287` | Release uses read/check/write and generates a fresh payout reference each invocation. Concurrent verification/release requests can initiate multiple transfers; local state is marked RELEASED before provider success | Atomic state claim, durable payout intent, stable reference/idempotency key, explicit processing/succeeded/failed/unknown states. Concurrent callbacks and release calls produce one logical payout |
| PAY-02 / P1 | `Jobba-server/src/services/payment.js:211–228` | Verification checks provider success but does not compare returned amount/currency with the stored payment before progressing | Compare reference, currency, and amount in minor units; reject mismatches without changing job/payment state. Preserve evidence for reconciliation |
| PAY-03 / P1 | `Jobba-server/src/routes/payment.js:26`, `:45–51` | `GET /:paymentId` precedes `/banks` and `/bank-accounts`, so those literals reach the payment-detail handler | Register static routes first and validate ID parameters. Route tests must prove banks and bank-account listing reach their intended handler |
| AUTH-01 / P1 | `Jobba-server/src/routes/auth.js:12–16`; account-type controller uses `req.user.id` | Account-type completion is registered before Auth populates the user | Authenticate the onboarding identity with a suitable restricted onboarding token or guard. Test authenticated unfinished accounts and anonymous calls |
| AUTH-02 / P1 | `Jobba-server/src/services/auth.js:56`; `src/websocket/index.js:18–33`; HTTP auth checks in `src/middlewares/auth.js` | OTP generation uses `Math.random`; socket authentication lacks the account-status checks used by HTTP | Secure random OTP generation; expiry/single-use/attempt controls; shared HTTP/socket account policy |
| API-01 / P1 | `Jobba-server/src/middlewares/paginate.js:3–9`; `src/app.js` JSON/urlencoded parser limits | Pagination accepts arbitrary numeric conversion without bounds; request bodies allow up to 1 GB | Validate positive finite integer pages and bounded page sizes; define per-route body limits and reject invalid input before queries |
| CHAT-01 / P1 | `Jobba-server/src/services/chat.js:386–407`; `src/models/job.js:30` | Conversation search populates/reads `title`, but jobs use `jobTitle`; a populated job without `title` can cause TypeError | Use the correct field and null-safe search semantics; test existing job, missing job, and missing title fixtures |
| OPS-01 / P1 | `Jobba-server/src/app.js` async database initialization; `src/bin/www/server.js` listen path; `src/routes/main.js:28` | Server startup is not gated on awaited DB readiness; health reports OK without checking DB | Fail-fast dependency initialization; readiness checks; separate liveness; graceful HTTP/socket/worker shutdown |
| ERR-01 / P1 | `Jobba-server/src/helpers/errorResponse.js`; final error middleware in `src/app.js` | Arbitrary numeric error codes can be treated as HTTP statuses; internal error messages can be returned to clients | Typed application errors, safe unknown-error mapping, validated HTTP status range, redacted server-side causes |
| SEC-03 / P1 | `Jobba-server/src/services/job/worker/index.js:141–163` | Private job access accepts `applicantType: invite` without verifying the requesting worker was invited | Pass authenticated worker identity into the service and verify invitation/assignment eligibility. Unrelated workers must not see private jobs |
| AUTH-03 / P1 | `Jobba-server/src/models/user.js:78–98`; `src/services/auth.js:98`; `src/middlewares/auth.js:39` | Default ACTIVE/unverified accounts, signin checks, and protected-route checks implement different account-state rules | Define one status/verification policy; table-test each state against signin, HTTP, socket, and privileged operations |
| CHAT-02 / P1 | `Jobba-server/src/services/chat.js:120–151` | Sender membership is checked, but client-supplied receiver is not verified as the other conversation participant | Derive receiver from membership on the server; reject inconsistent payloads; test arbitrary/self receiver IDs |
| CHAT-03 / P2 | `Jobba-server/src/controllers/chat.js:40–48`; `src/services/chat.js:46–57` | Omitted `isActive` becomes false, filtering out active conversations by default | Preserve omitted filters; validate true/false explicitly; test omitted, false, true, and malformed values |
| CHAT-04 / P2 | `Jobba-server/src/services/chat.js:149–151,219–248` | Read/modify/save unread counts and separate read resets can lose concurrent updates | Atomic message sequencing/read watermarks or appropriate transaction design; concurrent send/read counts reconcile |
| CHAT-05 / P2 | `Jobba-server/src/models/message.js:85–96`; `src/services/chat.js:262–279` | Saving an older deleted message can reset the conversation's last-message projection | Update preview only for eligible newest messages; explicitly recompute when newest message is deleted |
| CHAT-06 / P2 | `Jobba-server/src/services/chat.js:292–348` | Global archive state conflicts with per-user archivedBy state | Confirm personal/shared archive policy, then store and query consistent per-participant state |
| PAY-04 / P1 | `Jobba-server/src/services/payment.js:72–124`; payment model indexes | Concurrent checkout initialization can create multiple active operations before any record excludes the other | Claim a business idempotency key before provider initialization; unique constraint and stable reference; parallel initiation test |

For PAY-01, duplicate payout is a source-level concurrency risk, not a measured production incident. Verify it with a fake payment provider and controlled concurrent requests before modifying provider integration.

### B. Mobile correctness and compatibility

| ID / priority | Evidence | Consequence | Planned fix and acceptance |
|---|---|---|---|
| APP-01 / P1 | `Jobba-App/app/(auth)/signup-client.tsx:16,35`; worker equivalent; `Jobba-server/src/controllers/auth.js:12–19` | Forms send `phone`; backend requires `phoneNumber` | Define a signup DTO mapping explicitly. Both roles send expected fields and omit confirmation-only fields |
| APP-02 / P1 | `Jobba-App/src/services/authService.ts:38–43`; server response helper | Signup reads `response.data.token`, while server returns `response.data.data.token` | One typed transport decoder for signup/signin; capture real response fixtures and test both |
| APP-03 / P1 | `Jobba-App/src/services/authService.ts:20,43`; `src/store/authStore.ts:37–39` | The asynchronous store login is not awaited before returning/navigating | Await persistence; handle SecureStore failure; navigation occurs only after successful auth state transition |
| APP-04 / P1 | `Jobba-App/app/_layout.tsx:24–44`; `app/index.tsx:23–38` | Early returns occur before hooks; hydration or authentication changes can change hook order | Call hooks unconditionally or split components. Test cold start, hydration, login, logout, role changes, and font failure |
| APP-05 / P1 | `Jobba-App/src/store/authStore.ts` rehydration callback; token helper and persisted store | Hydration flag is directly mutated; token state exists in two storage locations | Explicit reactive hydration completion/error state and one authoritative session design; test partial storage failures and restart |
| APP-06 / P1 | `Jobba-App/components/navigation/CustomDrawerContent.tsx:12–19`; `src/services/authService.ts:18`; login screen submit logging | Logout is wired but redirects even when cleanup fails; token and credential logging occurs in auth paths | Handle cleanup failures explicitly; verify persisted/session state and protected navigation; remove stale TODO and sensitive logs |
| APP-07 / P1 | `Jobba-App/src/services/api.ts:5` | API address is a hardcoded private LAN URL | Environment-specific configuration with validated base URL; device/emulator/staging checks |
| NAV-01 / P2 | `Jobba-App/app/(client)/(tabs)/home/index.tsx:18`; `app/(client)/_layout.tsx:29` | Disputes quick action uses `/(client)/disputes`, but the registered route is under `drawer/disputes` | Correct destination; route-link tests for dashboard actions and drawer entries |

APP-04 is a concrete hook-order problem. This audit has not demonstrated an infinite render loop. APP-05's exact hydration symptom still requires execution against the installed Zustand/Expo versions.

## 5. Loop and performance investigation

The reviewed source does not establish a general infinite-loop defect. The important cases are work growing with stored data, missing bounds, repeated side effects, and app lifecycle errors.

| Area | Observed code | Remediation | Measurement/test |
|---|---|---|---|
| Conversation search | `src/services/chat.js:388–412` loads and populates all conversations, then filters in JavaScript | Filter in SQL with membership scope, indexed search where justified, stable ordering, and bounded results | Increasing conversations per user; query plans, response size, allocations, p95/p99 |
| Recommendation exclusions | `src/services/job/worker/index.js:189–196` materializes every applied job ID into `$nin` | SQL `NOT EXISTS`/anti-join against indexed applications; preserve skill/location/privacy rules | Increasing application history; explain plan, correctness and memory comparison |
| Location matching/search | Recommendation regexes and geocoding response scans | Normalize region fields, escape literal search as appropriate, index actual filters; stop scans when complete | Match semantics, large datasets; geocoder CPU vs network latency |
| Pagination | Unbounded/invalid `page` and `limit` | Bounded defaults/maxima, stable order; cursor pagination where compatible | Negative, zero, NaN, fractional, huge limit/page; no unbounded response |
| External I/O | Paystack Axios client has no explicit timeout; geocoding request likewise | Go client timeouts/context deadlines, bounded retry ownership, provider-aware reconciliation for writes | Slow/failed provider, cancellation, retry exhaustion, ambiguous payout result |
| Repeated payment actions | Callback, webhook, and release can overlap | Durable idempotency and atomic state transitions | Concurrent duplicate webhook/release calls; crash before/after provider acceptance |
| Socket connection state | `users` map stores one socket ID per user; disconnect deletes user entry | Per-user connection set/rooms and correct last-connection semantics | Two devices; disconnect one; remaining user stays online |
| Mobile startup/navigation | Conditional hooks, asynchronous persistence, logout failure handling | Explicit startup/auth state machine and guarded navigation | Cold start, slow storage, expired session, repeated taps, logout/restart |

Keep bounded transformations such as worker work-history mapping over five reviews and a small role-action list. Replacing ordinary loops with goroutines is not an acceptance criterion.

### Performance protocol

1. Create representative small, medium, and large fixtures for conversations, messages, applications, jobs, and notifications.
2. Record resource limits, input size, concurrency, duration, query count, latency percentiles, CPU, memory, and errors.
3. Establish product targets after measuring the baseline.
4. Compare old and Go behavior with the same fixtures and comparable resources.
5. Require each optimization to preserve permissions, ordering, duplicates, transaction semantics, and partial-failure behavior.
6. Test overload and recovery; queues, goroutines, retries, and memory must remain bounded.

No latency or speedup claim is made in this planning artifact.

## 6. Mobile implementation matrix

The app currently exposes signin/signup API calls in `src/services/authService.ts`. Many routed screens are placeholders or render static dashboard content. A backend rewrite alone will not complete those journeys.

| Journey | Current state in inspected app | Planned app work |
|---|---|---|
| Login/signup | Forms and API functions exist, with defects above | Correct DTOs, response parsing, secure persistence, error/loading states |
| Verification/reset/Google onboarding | Backend paths exist; complete corresponding UI flows were not found | Add screens/deep links and account-state handling when these flows enter the release scope |
| Startup and role routing | Root/index route by persisted state | Stable hydration and server session validation; protect client/worker routes |
| Logout | Local logout wired; cleanup failure still redirects; no server-session revocation found | Test cleanup and restart; add revocation when refresh sessions are introduced |
| Home dashboards | Components and sample/static content | Feed real statistics, jobs/projects, loading/empty/error states |
| Job search, posting, applications, offers | Route shells/limited UI; no corresponding API service layer found | Implement typed services and one end-to-end journey per Go slice |
| Chat | Messages screens exist; no Socket.IO client dependency declared | Implement conversations/messages and selected real-time transport after protocol decision |
| Wallet/payment | Wallet screens exist; no complete payment client flow found | Checkout callback, verified status, history, bank accounts, escrow actions |
| Reviews/disputes/profile/settings | Backend coverage varies; app screens largely incomplete | Define release slice and wire authorized operations with validations |
| Growth/support/analytics | Navigation surfaces do not imply backend functionality | Define product behavior before creating new endpoints |

Track two independent statuses for every capability: **backend migration status** and **mobile integration status**. A placeholder screen cannot satisfy end-to-end acceptance.

## 7. Target architecture

Proposed source location: `Jobba-server-go/`, created when implementation begins. Keep the existing Node repository available for characterization, urgent fixes if necessary, and rollback. Establish whether the Go source will become a new repository or be integrated into `Jobba-server` before the first implementation commit; do not silently create nested repositories.

```text
Jobba-server-go/
  cmd/api/main.go
  internal/
    config/
    auth/ user/ role/ skill/ worker/
    job/ application/ offer/
    chat/ notification/ payment/ review/ dispute/
    shared/
    middleware/
    server/
    database/
    email/ location/ paystack/
  db/
    migrations/
    queries/
    generated/
  docs/
  sqlc.yaml
  go.mod
  .env.example
```

Each business feature follows the supplied `domain.go`, `interfaces.go`, `repository.go`, `service.go`, `handler.go`, `dto.go`, `error.go`, and `service_test.go` convention.

### Required implementation rules

- Services own business rules and ownership checks, including socket-originated operations.
- Repository wrappers own SQL, transactions, and generated/domain conversion.
- For workflows spanning repositories, define an atomic use-case repository operation or transaction boundary through feature contracts. Independent transactions do not provide shared atomicity. Keep sqlc/pgx transaction types out of service APIs and network calls out of long-running database transactions.
- Handlers bind/validate, call a service, and respond.
- Infrastructure satisfies feature interfaces; framework contexts do not cross into service APIs.
- Go files within one package do not enforce layer isolation. Review/static checks must enforce domain purity and generated-type containment.
- Shared AppError sentinels are immutable; `WithError` clones; `errors.Is` matches stable codes; wrapped causes never reach the client.
- Config validates required dependencies, secrets, environment values, and production CORS before accepting traffic.
- zap logs carry request/job correlation with redaction.
- `/healthz` checks critical dependencies with deadlines; separate liveness avoids dependency-driven restart storms.
- Shutdown stops intake, drains bounded work, stops background tasks, then closes pools.
- Migration files are schema authority. Commit sqlc output; never edit generated files manually.
- Choose supported tool/dependency versions using official documentation during implementation; this plan does not pin versions without verification.

## 8. API and real-time compatibility strategy

### REST

Current responses generally use:

```json
{
  "status": true,
  "success": true,
  "message": "Human-readable message",
  "data": {}
}
```

The blueprint prescribes `{ "data": ... }` and `{ "error": { "code", "message" } }`. Preserve `/api/v1` through an explicit compatibility presenter while mobile code migrates. Adopt the canonical envelope at a versioned boundary or coordinated client release. Record this temporary presenter as a deliberate blueprint deviation.

Freeze route methods/paths, statuses, pagination metadata, populated objects, `_id` versus `id`, timestamps, null fields, numeric amounts, errors, and event formats in fixtures. Correct known security and correctness bugs deliberately; do not preserve them as parity requirements.

### Authentication additions versus existing behavior

Current JWTs contain `{id, role}` with a default seven-day expiry (`Jobba-server/src/services/auth.js:20–23`). No refresh/logout route or rate-limiting implementation was found in the inspected source. The app has a request interceptor only; it does not currently have a refresh-retry loop.

The blueprint's refresh-token families, reuse detection, and per-identity/per-IP rate limits are new implementation scope. Define legacy-token coexistence or reauthentication, atomic rotation, server-side revocation/logout, and mobile storage before introducing them. If mobile refresh is added, use one in-flight refresh operation shared by waiting requests, a bounded retry, and an explicit logout path on failure.

OTP storage is currently bcrypt-hashed, not plaintext. Replace insecure generation with cryptographic randomness and design expiry, attempt limits, resend behavior, identity/action binding, and single-use consumption together. If keyed OTP storage replaces bcrypt, document key management and migration rather than assuming a plain fast hash protects a low-entropy code.

### Socket.IO

The existing server uses Socket.IO at `/socket.io`; raw WebSockets are not a drop-in wire replacement. The inspected mobile package does not declare a Socket.IO client, but other consumers remain unknown.

Before implementing chat transport:

1. Inventory every real client and event payload.
2. Test a Go implementation against the exact protocol features used, including heartbeat, reconnect, room semantics, and acknowledgments; or coordinate a new protocol with all clients.
3. If a temporary Node socket gateway is necessary, document it as a transitional exception to the one-binary target and define its retirement.
4. Require server-side membership checks regardless of transport.
5. Define message idempotency, ordering, unread/read state, reconnect resynchronization, and multi-device presence.

Do not select an unverified Go Socket.IO library solely by name.

## 9. MongoDB-to-PostgreSQL migration

This database change is required by the selected SQL-oriented target, but the source code does not reveal whether existing data is disposable or live. Preserve data by default until that is established.

### Proposed relational mapping

| Mongo model/data | SQL destination proposal | Important constraints |
|---|---|---|
| User, Role | users, roles, role_permissions | Unique normalized identities; account type distinct from role; no public privilege assignment |
| Skill and user skills | skills, user_skills | Unique catalogue names; valid foreign keys |
| Saved workers | saved_workers | Unique client/worker pair |
| WorkerProfile embedded arrays | worker_profiles plus bounded JSONB or child tables | Choose according to query/update needs; one profile per worker |
| Job and skills | jobs, job_skills | Lifecycle constraints; valid budgets/dates; owner and optional assigned worker |
| Application | applications | Unique job/worker; acceptance coordinated with job assignment |
| SavedJob | saved_jobs | Unique worker/job |
| Offer | offers | Sender/receiver/job validity; legal accept/reject transitions |
| Conversation and participant state | conversations, conversation_members | Participant-pair/job uniqueness defined explicitly; per-user archive/read state |
| Message | messages | Conversation membership, stable ordering, idempotent client message ID |
| Notification | notifications | Recipient-scoped queries and bounded history |
| Payment, bank account | payments, payout_attempts, bank_accounts, provider_events | Stable references; exact money; unique default account per worker; transition/idempotency constraints |
| Review | reviews | Unique job/reviewer; eligible reviewer and reviewee |
| Dispute | disputes, dispute_evidence as needed | Valid job participants; active-dispute uniqueness and resolution policy |
| OTP/reset/OAuth data | verification_codes, password_reset_tokens, external_identities | Expiry and single-use checks; unique provider identity; safe secret storage |
| Settings | settings/configuration tables only if used | Audit actual consumers before carrying unused structure |

### Identity and data semantics

- Existing documents mix Mongo `_id` values with `uniqueId` UUIDs. Preserve public IDs initially, for example using validated legacy text keys, or use a complete mapping table if internal UUIDs are chosen. Do not blindly replace identifiers.
- Preserve bcrypt hashes and test verification of existing `$2*` variants; do not force password resets solely because of language conversion.
- Convert monetary values using an explicit rounding policy to integer minor units. Preserve historic amounts/fees; do not recalculate past obligations from current fee configuration.
- Preserve timestamps, enum values, nullable/missing semantics, and reference relationships through explicit transformations.
- Mongo TTL indexes do not become PostgreSQL TTL behavior automatically. Check expiry in queries and run bounded cleanup.
- Audit geospatial queries before choosing PostGIS. Preserve longitude/latitude order, distance units, and matching behavior; a `2dsphere` index is not a reason to invent new geo functionality.
- Replace denormalized application arrays and redundant relationships only after reconciliation identifies their authoritative source.
- Quarantine invalid/orphaned records for review rather than silently discarding them.

### Migration sequence

1. Export a protected snapshot and record counts, schema variants, duplicate identities, and orphan references.
2. Build SQL migrations and deterministic import transformations.
3. Import in dependency order with bounded, restartable batches.
4. Reconcile counts, references, monetary totals, statuses, and sampled payloads.
5. Run application and migration tests against the imported copy.
6. Decide cutover mode based on live traffic and downtime requirements.
7. Freeze legacy writes for a maintenance-window cutover when acceptable, import final changes, verify, then switch ownership.
8. If uninterrupted coexistence is required, design explicit replication/change capture and reconciliation first. Do not use informal dual writes or route writes between independent unsynchronized databases.

Incremental development of Go features does not automatically mean safe endpoint-by-endpoint production cutover across two data stores. Read-only shadow comparisons can use snapshots; write routing requires an authoritative-data design.

### Rollback boundary

Before Go accepts writes, routing back to Node is straightforward if the old database remains intact. After Go-only writes begin, rollback requires reverse replication/reconciliation or forward repair. Define that boundary, retention window, and recovery procedure before release. A SQL down migration does not restore data to MongoDB.

## 10. Payment and job state design

Payment is not a generic CRUD conversion. Create a state transition table covering:

- Job creation, assignment, application acceptance, activation, completion, cancellation, and dispute.
- Checkout initialized, provider payment verified, escrow held, release requested, payout processing, payout confirmed, failed, reversed, and outcome unknown.
- Which party may invoke each transition and which invariant must hold.
- Which transitions need one database transaction.
- Which external effects need a durable intent/outbox and later confirmation.

Persist payment/payout intent and a stable reference before calling the provider. Avoid holding a SQL transaction open during network I/O. A timeout after provider acceptance must lead to reconciliation with the same reference, not a fresh payout.

Required invariants:

1. One logical payout per authorized release.
2. Payment amount/currency/reference match the stored obligation.
3. Failed or uncertain external outcomes cannot be reported as completed settlement.
4. Accepted application, assigned worker, and payment beneficiary agree.
5. Duplicate/out-of-order webhooks cannot regress state.
6. Job dispute and payment dispute policies agree on whether funds may be released.
7. Money/fee arithmetic is exact and auditable.

Payment tests use an isolated fake provider first, then sandbox-provider verification when configured. No planning test should send real payments.

## 11. Implementation phases and exit gates

| Phase | Work | Exit evidence |
|---|---|---|
| 0. Baseline and test repair | Restore reproducible installs; repair harness; map routes/events; establish live-data status | Non-watch test command, isolated dependencies, captured contracts, baseline report |
| 1. Urgent fixes | SEC/PAY/APP priorities; coordinate Node fixes if live | Reproductions and regression tests; independently reviewed patches |
| 2. Go foundation | Config, errors, logging, router, readiness, shutdown, DB/migrations/sqlc, CI | Clean build, safe startup/error behavior, real test DB |
| 3. Pilot | Public skill catalogue reads with legacy presenter; imported catalogue fixture | Existing contract parity, query limits, independent verification |
| 4. Identity and profiles | Auth/user/role/worker; mobile auth/startup fixes | Signup/login/verification/reset/status/ownership and restart flows pass |
| 5. Job marketplace | Job search/posting/lifecycle, applications/offers, saved entities, reviews | Worker/client end-to-end journey and concurrent-selection invariants |
| 6. Messaging | Chat/notification persistence and transport | Membership, ordering, duplicate delivery, multi-device/reconnect tests |
| 7. Payments and disputes | Provider adapter, payout state machine, reconciliation, settlement policy | Duplicate/concurrent/failure cases and financial invariants verified |
| 8. App integration | Connect remaining in-scope screens as APIs stabilize | Real loading/empty/error/success states; no sample data for accepted features |
| 9. Release rehearsal | Data import, protected staging, load/failure tests, cutover and recovery | Reproducible runbook with thresholds and responsible operator |
| 10. Cutover and retirement | Authorized rollout, observation, old-service retirement | No unowned routes/jobs/clients; reconciled data; recovery window satisfied |

Phases 4–8 overlap only where interfaces are stable. Payment defect remediation begins early; the full Go payment feature is released after supporting identity/job invariants are ready.

## 12. Multi-agent implementation and verification

Use one coordinator and three workers. Assign work by bounded feature and owned paths.

| Role | Responsibility | Acceptance handoff |
|---|---|---|
| Coordinator | Architecture, dependency graph, shared wiring/manifests/migration order, integration | Traceable task and final gate decision |
| Current-app / contract agent | Existing behavior, mobile journeys, fixtures, defect reproductions | Source-backed contract and expected differences |
| Conversion / implementation agent | Assigned Go feature and relevant regression tests | Patch, run commands, migrations, compatibility notes |
| Independent verifier | Reproduce defects, review architecture/security, run integration/failure checks | Pass/fail with evidence and unresolved limits |

Rotate roles by phase; the verifier must not be the primary author of the patch accepted. Multiple implementers may work on isolated features when a slot is available, but independent verification remains a separate gate.

Shared files such as `go.mod`, router wiring, config, migrations, and sqlc output have one integration owner at a time. Generate sqlc output after integrating schema/query changes. Avoid competing edits across the two existing repositories.

Task card:

```text
ID / feature / baseline commit:
Observed problem and source references:
Blueprint requirements:
Owned paths and dependency interfaces:
Compatibility preserved / intentional changes:
Implementation and data effects:
Regression and integration commands:
Performance/failure cases:
Independent verifier:
Rollout/rollback conditions:
Evidence links and status:
```

Status: planned → characterized → implementing → implemented → independently verified → integrated → released → observed.

## 13. Verification plan and existing test blockers

### Confirmed harness issues

- `Jobba-server/jest.config.js` points to `./tests/setup.js`; the existing file is `src/tests/setup.js`.
- `src/tests/unit/emailVerification.test.js:11` mocks nonexistent `models/permission`.
- `src/tests/unit/auth.test.js` tests addition, not authentication.
- The test script uses `jest --watchAll`; CI needs a deterministic terminating command.
- `src/tests/setup.js` exports a connect/cleanup function rather than registering normal test hooks; it contains database-drop cleanup. Separate unit tests from integration setup and require a disposable database with an explicit target guard before enabling cleanup.
- Backend/application dependencies are not installed locally, so no full suite or mobile lint result is claimed.

### Baseline repair sequence

1. Install from lockfiles in a controlled development environment.
2. Repair Jest paths/module mocks and align test tooling versions only as needed.
3. Separate pure service tests from database integration suites.
4. Replace placeholder tests with meaningful authentication and contract tests.
5. Provision disposable MongoDB for baseline tests and PostgreSQL for Go tests.
6. Stub email, Google identity/maps, and Paystack; no live side effects in normal CI.
7. Record failures that predate conversion.

### Required regression cases

| Area | Mandatory cases |
|---|---|
| Signup | Correct phone field/envelope; client/worker success; anonymous admin rejected; duplicate identity; email failure recovery |
| Session/mobile | SecureStore slow/failure; awaited login; cold hydration; logout/restart; role guards; no credential logs |
| Authorization | Suspended/deleted users; wrong owner; arbitrary conversation join; spoofed socket sender/receiver/read receipt |
| Jobs/applications | Private-job visibility; only eligible invited worker where applicable; duplicate application; concurrent accept/assign; invalid transitions |
| Search | Correct jobTitle; missing populated job; large conversation history; bounded pagination; large application history |
| Payments | Amount/currency mismatch; repeated webhook; parallel release; provider timeout after success; reversal; retry with same reference |
| Data | Legacy IDs/password hashes; orphan/duplicate records; exact money; TTL expiry; import restart; reconciliation |
| Runtime | Dependency startup failure; readiness vs liveness; cancel/timeout; overload; graceful socket/HTTP/background drain |
| App features | Each accepted screen calls actual API and handles loading/empty/error/success; refresh/reconnect does not duplicate actions |

### Go CI gates

Run formatting checks, `go vet ./...`, `go test ./...`, `go test -race ./...` on supported infrastructure, and `go build ./...`. Add deterministic sqlc regeneration checks, real-database integration, selected benchmarks, and pinned dependency scanning.

The independent verifier checks both blueprint compliance and current-app compatibility. Passing unit tests alone cannot accept transactional payment behavior, database migrations, or mobile integration.

## 14. Release and operational plan

- Build the Go binary/container with reproducible inputs.
- Replace deployment configuration deliberately: `Jobba-server/vercel.json` references `bin/www/server.js`, while source/build entrypoints use `src/bin/www/server.js` and `build/bin/www/server.js`. Confirm actual deployment rather than assuming this file is authoritative.
- Choose hosting that supports the selected real-time transport and graceful shutdown.
- Configure secrets externally, TLS, proxy trust, bounded body sizes, CORS, and service deadlines.
- Add dashboards for error rate, latency, DB pools, goroutines, message delivery, provider failures, payout reconciliation, and data-import progress.
- Set numerical rollout thresholds from measured workload/product requirements.
- Rehearse traffic switching, data ownership, webhook routing, in-flight work, and rollback/forward recovery.
- Retain the old deployment and protected backup until the defined recovery window ends.

No production deploy, data mutation, or provider request was performed during this audit.

## 15. Initial actionable task board

| Ticket | Scope | Dependency | Done when |
|---|---|---|---|
| BASE-01 | Repair and run baseline harness | Dependency installation/test environment | Meaningful unit tests terminate; isolated integration setup |
| CONTRACT-01 | Route/event fixtures and app API matrix | Source audit | Existing envelope/ID/status/event behavior recorded |
| SEC-01/02 | Public role creation and socket membership | BASE-01 | Negative authorization cases pass |
| PAY-01/02/03 | Release safety, provider comparison, static routes | BASE-01 | Controlled concurrency/provider/route regressions pass |
| APP-01–07 | Signup, persistence, hooks, logout, logs, URL config | CONTRACT-01 | Device/emulator auth lifecycle passes |
| DATA-01 | Mongo inventory and SQL mapping | Data snapshot availability | Import/reconciliation design handles real schema variants |
| GO-01 | Blueprint foundation and CI | CONTRACT-01, schema decisions | Foundation gates pass |
| GO-02 | Skill read pilot | GO-01, DATA-01 | Legacy contract and query limits verified |
| GO-03 | Auth/user/role/worker | GO-02 | Existing passwords and mobile sessions handled correctly |
| GO-04 | Jobs/applications/offers/reviews | GO-03 | Core marketplace journey and concurrency tests pass |
| GO-05 | Chat/notifications | GO-03, transport decision | Authorized live messaging and reconnect behavior |
| GO-06 | Payments/disputes | GO-04, PAY remediation | Idempotent settlement and reconciliation |
| APP-08 | Connect agreed business screens | Corresponding GO tickets | End-to-end product acceptance per journey |
| RELEASE-01 | Staging import/load/failure/cutover rehearsal | Required GO/APP tickets | Runbook and recovery demonstrated |
| RELEASE-02 | Authorized rollout and observation | RELEASE-01 | Thresholds and data invariants hold |

Estimate after baseline repair and the skill pilot. Separate data migration, mobile completion, and independent verification effort from Go code conversion; they are substantial deliverables.

## 16. Decisions that remain open

These do not block planning or isolated implementation preparation, but must be settled before dependent release work:

- Is the current backend live, and must existing data/accounts be retained?
- Which mobile journeys constitute the first release versus later product completion?
- Which consumers besides this Expo app use REST or Socket.IO?
- Is a maintenance-window database cutover acceptable, or is continuous replication required?
- What are the intended invitation, dispute-resolution, escrow-release, and refund rules?
- Where should Go source live in version control, and which deployment environment will run it?

Default planning assumptions: preserve data, preserve existing public contracts through a compatibility boundary, implement one Go modular monolith, and keep rollout reversible up to a clearly documented data-write boundary.

## 17. Completion criteria

The work is complete only when:

- Required backend routes/events are migrated or explicitly retired.
- The architecture is implemented, with deviations recorded.
- Confirmed in-scope defects have independent regression evidence.
- Accepted mobile journeys use real APIs and survive restart/failure paths.
- Data import and money/state invariants reconcile.
- Load, security, cancellation, race, and dependency-failure checks pass their agreed criteria.
- Deployment, observation, recovery, and operational ownership are documented and exercised.

As of 23 October 2026, the implementation, compatibility closure, and local verification gates are complete as recorded in `FINAL_COMPLETION_REPORT.md`. Environment-specific staging import/reconciliation, external provider sandbox testing, deployment configuration, authorized cutover, observation, and production ownership remain required; no production deployment is claimed.
