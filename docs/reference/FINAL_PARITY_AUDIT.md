# Final Node-to-Go Route Parity Audit

Date: 2026-10-22

Scope: every HTTP endpoint mounted by `Jobba-server/src/routes`, compared with the current Go `net/http` registrations, handlers, PostgreSQL migrations, and legacy response envelope. `P` means route and core behavior are present in Go, and `Δ` means present with an intentional or material contract difference. Go-only routes are listed separately.

## Executive result

The Go backend covers the main auth, job, application, marketplace, worker-profile, payment, chat, notification, dispute, skill-admin, role-admin, and user-admin flows. The PostgreSQL migrations contain backing tables for every implemented Go feature: identity/OTP/password recovery, worker profiles and skills, jobs/saved jobs, applications, offers/reviews/saved workers, payments/bank accounts/payout intents/webhook receipts, chat, notifications, and disputes.

All Node HTTP routes now have Go registrations and working handlers. The final compatibility additions are complete: authenticated account-type selection, public skill-to-worker lookup, the legacy payment-dispute adapter, and the `/api/v1/health` readiness alias. The complete `NewComplete` route table is covered by a composition regression so ambiguous `net/http.ServeMux` patterns fail during tests instead of at production startup.

No HTTP route-parity blocker remains. Production cutover still requires the normal operational prerequisites: applied PostgreSQL migrations, configured JWT/Google/Paystack/email secrets, imported legacy data, and deployment probes targeting one of the documented health endpoints. Client teams must also adopt the documented REST-polling chat protocol; Go intentionally does not emulate the Node Socket.IO wire protocol.

One isolated Go defect was fixed during this audit: malformed payment pagination such as `page=abc` or `limit=abc` now returns 400 instead of silently using defaults. A focused regression test covers malformed, negative, and excessive values.

## Route-by-route matrix

### Root and health

| Node route | Status | Go route / evidence | Notes |
|---|---:|---|---|
| `GET /api/v1` | P | `GET /api/v1` in `internal/server/server.go` | Root response is implemented. |
| `GET /api/v1/health` | P | Compatibility alias to Go readiness; `/livez` and `/healthz` remain available. |

### Authentication and profile

| Node route | Status | Go route / notes |
|---|---:|---|
| `POST /api/v1/auth/signup` | P | Same path; rate limited; rejects privileged public roles. |
| `POST /api/v1/auth/signin` | P | Same path; Go also aliases `/auth/login`. |
| `POST /api/v1/auth/forgot-password` | P | Same path and recovery migration support. |
| `POST /api/v1/auth/reset-password` | P | Same path and recovery migration support. |
| `POST /api/v1/auth/google` | Δ | Registered only when Google auth is configured; otherwise 404. |
| `POST /api/v1/auth/account-type` | P | Same authenticated path; limited to one-time selection of `client` or `worker`. |
| `POST /api/v1/auth/resend-verification` | P | Same path. |
| `POST /api/v1/auth/verify-email` | P | Same path. |
| `GET /api/v1/auth/profile` | P | Same path, authenticated. |
| `PUT /api/v1/auth/profile` | P | Same path, authenticated. |
| `POST /api/v1/auth/change-password` | P | Same path, authenticated. |
| `POST /api/v1/auth/set-password` | P | Same path, authenticated. |

### Roles and admin users

| Node route | Status | Go route / notes |
|---|---:|---|
| `GET /api/v1/role/all` | Δ | Same path; Go enforces admin account type but does not reproduce Node resource-operation RBAC checks. |
| `POST /api/v1/role/new` | Δ | Same path; same RBAC difference. |
| `GET /api/v1/role/single/:id` | Δ | Same path; same RBAC difference. |
| `PUT /api/v1/role/update/:id` | Δ | Same path; same RBAC difference. |
| `DELETE /api/v1/role/delete/:id` | Δ | Same path; same RBAC difference. |
| `GET /api/v1/admin/users` | P | Same path, admin guarded. |

### Client jobs, applications, workers, offers, and reviews

| Node route | Status | Go route / notes |
|---|---:|---|
| `POST /api/v1/jobs/client` | P | Same path and client guard. |
| `GET /api/v1/jobs/client` | P | Same path with bounded pagination. |
| `GET /api/v1/jobs/client/statistics` | P | Same path. |
| `GET /api/v1/jobs/client/reviews/:userId` | P | Same path. |
| `GET /api/v1/jobs/client/:jobId` | P | Same path and ownership checks. |
| `PUT /api/v1/jobs/client/:jobId` | P | Same path. |
| `DELETE /api/v1/jobs/client/:jobId` | P | Same path. |
| `PUT /api/v1/jobs/client/:jobId/close` | P | Same path. |
| `PUT /api/v1/jobs/client/:jobId/complete` | P | Same path. |
| `PUT /api/v1/jobs/client/:jobId/assign` | P | Same path. |
| `GET /api/v1/jobs/client/:jobId/applications` | P | Same path. |
| `PUT /api/v1/jobs/client/application/:applicationId/accept` | P | Same path. |
| `POST /api/v1/jobs/client/:jobId/review` | P | Same path. |
| `GET /api/v1/jobs/client/workers` | P | Same path and query filters. |
| `GET /api/v1/jobs/client/workers/saved-workers` | P | Same path. |
| `GET /api/v1/jobs/client/workers/:workerId` | P | Same path. |
| `POST /api/v1/jobs/client/workers/:workerId/save` | P | Same toggle behavior. |
| `POST /api/v1/jobs/client/offers` | P | Same path. |

### Worker jobs, applications, offers, and reviews

| Node route | Status | Go route / notes |
|---|---:|---|
| `GET /api/v1/jobs/worker` | P | Same path. |
| `GET /api/v1/jobs/worker/search` | P | Same path. |
| `GET /api/v1/jobs/worker/recommended` | P | Same path. |
| `GET /api/v1/jobs/worker/saved` | P | Same path. |
| `GET /api/v1/jobs/worker/applications` | P | Same path. |
| `GET /api/v1/jobs/worker/category/:category` | P | Same path. |
| `GET /api/v1/jobs/worker/reviews/:userId` | P | Same path. |
| `GET /api/v1/jobs/worker/:jobId` | P | Same path. |
| `POST /api/v1/jobs/worker/:jobId/save` | P | Same path. |
| `DELETE /api/v1/jobs/worker/:jobId/save` | P | Same path. |
| `POST /api/v1/jobs/worker/:jobId/apply` | P | Same path. |
| `POST /api/v1/jobs/worker/:jobId/review` | P | Same path. |
| `POST /api/v1/jobs/worker/offers/:offerId/accept` | P | Same path. |
| `POST /api/v1/jobs/worker/offers/:offerId/reject` | P | Same path. |

### Worker profiles and skills

| Node route | Status | Go route / notes |
|---|---:|---|
| `GET /api/v1/workers/:id/profile` | P | Same path and authenticated client/worker/admin access. |
| `GET /api/v1/workers/profile/me` | P | Same path and worker guard. |
| `PUT /api/v1/workers/profile` | P | Same path and worker guard. |
| `GET /api/v1/skills` | P | Go registers both trailing-slash variants. |
| `GET /api/v1/skills/:skillId/workers` | P | Same public path with bounded pagination and legacy-compatible worker output. |
| `POST /api/v1/admin/skills` | P | Same path and admin guard. |
| `PUT /api/v1/admin/skills/:skillId/status` | P | Same path and admin guard. |

### Chat

| Node route | Status | Go route / notes |
|---|---:|---|
| `POST /api/v1/chat/conversations` | P | Same path; participant/job validation backed by chat tables. |
| `GET /api/v1/chat/conversations` | P | Same path and boolean/pagination filters. |
| `GET /api/v1/chat/conversations/search` | P | Same path. |
| `GET /api/v1/chat/conversations/:conversationId` | P | Same participant visibility rule. |
| `POST /api/v1/chat/conversations/:conversationId/archive` | P | Same path. |
| `POST /api/v1/chat/conversations/:conversationId/unarchive` | P | Same path. |
| `POST /api/v1/chat/conversations/:conversationId/urgent` | P | Same path. |
| `GET /api/v1/chat/conversations/:conversationId/messages` | P | Same path and pagination. |
| `POST /api/v1/chat/conversations/:conversationId/messages` | P | Same path. |
| `POST /api/v1/chat/conversations/:conversationId/read` | P | Same path. |
| `DELETE /api/v1/chat/messages/:messageId` | P | Same path and sender ownership. |

The Node Socket.IO event surface is outside HTTP route parity. Go deliberately uses authenticated REST reads/writes plus polling through the conversation, message, and unread-count endpoints. It does not emulate Socket.IO events such as `conversation:join`, typing, delivery, online/offline, or pushed messages. This is the documented cutover protocol, so clients must use polling rather than attempt a Socket.IO connection to Go.

### Notifications

| Node route | Status | Go route / notes |
|---|---:|---|
| `GET /api/v1/notifications` | P | Same path and recipient-scoped list. |
| `PATCH /api/v1/notifications/:id/read` | P | Same path and recipient ownership. |
| `PATCH /api/v1/notifications/read-all` | P | Same path. |
| `DELETE /api/v1/notifications/clear-all` | P | Same path. |
| `DELETE /api/v1/notifications/:id` | P | Same path and recipient ownership. |
| `POST /api/v1/notifications/send` | P | Same path and admin check. |

### Payments

| Node route | Status | Go route / notes |
|---|---:|---|
| `POST /api/v1/payment/webhook` | P | Same public signed webhook; Go additionally stores idempotent webhook receipts. |
| `GET /api/v1/payment` | P | Same path; malformed pagination now returns 400. |
| `POST /api/v1/payment/initiate` | P | Same path; Go service enforces client role and ownership. |
| `GET /api/v1/payment/verify/:reference` | P | Same path and participant visibility. |
| `POST /api/v1/payment/:paymentId/release` | P | Same path; Go service enforces client role and atomic payout claim. |
| `POST /api/v1/payment/:paymentId/dispute` | P | Legacy adapter validates participant ownership and `IN_ESCROW` state, then uses the shared dispute creation workflow and marks the payment disputed. |
| `GET /api/v1/payment/banks` | P | Same path. |
| `POST /api/v1/payment/bank-accounts` | P | Same path; Go enforces worker role in service. |
| `GET /api/v1/payment/bank-accounts` | P | Same path and worker role. |
| `PATCH /api/v1/payment/bank-accounts/:accountId/default` | P | Same path and owner-scoped update. |
| `DELETE /api/v1/payment/bank-accounts/:accountId` | P | Same path and owner-scoped delete. |
| `GET /api/v1/payment/:paymentId` | P | Same path and participant visibility. |

Payment responses retain Go's precise minor-unit fee fields and now also expose Node-compatible `platformFeePercent`, `platformFeeAmount`, `workerReceivableAmount`, nested `paystackData`, payout reference/status, and escrow release fields. Persisted `SUCCEEDED` payout intents serialize as Node's legacy `SUCCESS` value.

### Disputes

| Node route | Status | Go route / notes |
|---|---:|---|
| `POST /api/v1/disputes` | P | Same path, participant roles only. |
| `GET /api/v1/disputes` | P | Same path; users see involved disputes. |
| `GET /api/v1/disputes/:id` | P | Same path and participant/admin visibility. |
| `POST /api/v1/disputes/:id/evidence` | P | Same path. |
| `GET /api/v1/admin/disputes` | P | Same path and admin-only result set. |
| `GET /api/v1/admin/disputes/:id` | P | Served by the generic dispute detail authorization, allowing admin. |
| `PUT /api/v1/admin/disputes/:id/assign` | Δ | Same path; Go accepts `admin`, while Node also names `superadmin` and `moderator` account types that its current user schema does not define. |
| `PUT /api/v1/admin/disputes/:id/resolve` | Δ | Same role difference. |

## Go-only HTTP routes

These do not block Node compatibility but should be documented for clients:

| Go-only route | Purpose |
|---|---|
| `POST /api/v1/auth/login` | Alias for sign-in. |
| `GET /api/v1/jobs` | Public job list. |
| `GET /api/v1/jobs/context/:jobId` | Authenticated consolidated job context; the static `context` prefix avoids ambiguous `ServeMux` overlap with client/worker job-detail routes. |
| `GET /api/v1/jobs/client/offers` | Client offer list. |
| `GET /api/v1/jobs/client/offers/:offerId` | Client offer detail. |
| `GET /api/v1/jobs/worker/offers` | Worker offer list. |
| `GET /api/v1/jobs/worker/offers/:offerId` | Worker offer detail. |
| `GET /api/v1/chat/unread-count` | Aggregate unread count. |
| `PATCH /api/v1/admin/users/:id/status` | Admin status mutation. |
| `PATCH /api/v1/admin/users/:id/account-type` | Admin account-type mutation. |
| `GET /metrics` | Operational metrics. |
| `GET /livez` | Liveness probe. |
| `GET /healthz` | Readiness probe. |

## Security and semantic findings

- Go payment authorization is primarily enforced in the service rather than route wrappers. Initiation/release/bank-account operations reject the wrong role, and list/detail/verify scope records to the actor. This is secure but differs structurally from the corrected Node route guards.
- Go intentionally applies a stricter, simpler admin policy: admin and role-management endpoints require the `admin` account type. Node's per-resource operation permissions and its undeclared `superadmin`/`moderator` labels are not reproduced. This is an accepted cutover policy, not a missing route; deployments that require delegated limited-admin roles would need a future authorization feature.
- Go chat routes are authenticated but do not use an outer worker/client role guard. Repository/service participant checks protect conversation data; an admin token can reach the handlers but cannot access a conversation without membership.
- Go webhook processing is stronger than Node in replay handling because `payment_webhook_receipts` records event identity and completion.
- Go validates request bodies with size limits and unknown-field rejection in core auth/payment flows. Some chat/dispute handlers use plain decoders, so strict unknown-field behavior is not uniform.

## Remaining blockers

No missing or conflicting HTTP route remains in the audited `NewComplete` server. The remaining cutover work is operational rather than an implementation gap: run migrations and legacy import, configure external-service credentials, point health probes at the deployed Go endpoints, and ensure chat clients use the agreed REST-polling protocol. Delegated per-resource admin permissions would be a separate future feature if product policy changes from the accepted admin-only model.

## Verification

Run from `Jobba-server-go`:

```text
go test ./...
go vet ./...
```

Run from `Jobba-server` without installed dependencies:

```text
node --experimental-vm-modules --test src/tests/node/backend-regressions.test.cjs
```

The final command exercises the real Node pagination, payment, chat, auth, and payment-route modules with dependency doubles; it does not copy their implementations.

Final results:

- `GOCACHE=/tmp/jobba-go-build-cache go test ./...`: passed across all Go packages.
- `GOCACHE=/tmp/jobba-go-build-cache go vet ./...`: passed with no diagnostics.
- Node dependency-free regression suite: 8 tests passed, 0 failed.
- `gofmt -d` on the edited Go files produced no output.
