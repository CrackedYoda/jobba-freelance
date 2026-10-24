# Jobba independent verification

Updated: 22 October 2026

Scope: source-level and offline verification of the existing Node backend, Expo app changes, and incremental Go backend against `BACKEND_BLUEPRINT.md` and `JOBBA_GO_MIGRATION_PLAN.md`. This is an integration review, not a release approval.

## Verification result

The current changes materially improve the Node security/payment/chat paths and repair the mobile signup transport contract. The Go foundation and skill pilot compile and test in isolation. One Go routing defect found during this review was fixed and covered by contract tests: the former `GET /api/v1/skills/` subtree pattern also captured unimplemented nested paths such as `/api/v1/skills/{skillId}/workers`; the pattern is now an exact trailing-slash match.

The conversion remains pre-release. The first Go read slice now preserves imported Mongo skill identifiers and retains pagination totals for an out-of-range page. It still implements only one of the two public skill routes. Authentication and job foundations are present but still need database-backed integration. Mobile authentication still has two independent persisted token copies, so the current dependency-free script does not prove durable login/logout behavior under storage failure.

The Go payment slice now has durable payment, bank-account, payout-intent, and webhook-receipt schemas; a real Paystack HTTP adapter behind a provider interface; strict ownership/role checks; exact amount/currency verification; stable initiation and payout references; atomic payout claims; and replay-safe signed webhook handling. Its state tests use local doubles and do not represent Paystack sandbox or PostgreSQL integration evidence.

The migration toolchain now includes a dry-run-first, hash-checkpointed JSONL importer for skills, users, profiles, jobs, applications, payments, conversations/messages, notifications, and disputes. Tests cover all collections, duplicate IDs, malformed JSON, dependency orphans, and checkpoint/source mismatch without MongoDB credentials. A Compose smoke script validates migration up/down behavior against a disposable database, but it could not run here because access to the local Docker daemon socket was denied.

## Offline evidence

| Check | Result |
|---|---|
| Node dependency-free regressions | Passed: 6/6 with `node --experimental-vm-modules --test src/tests/node/backend-regressions.test.cjs` |
| Modified Node syntax checks | Passed for auth, chat, payment, and websocket modules |
| Mobile transport check | Passed with `node --experimental-strip-types scripts/check-auth.mjs` |
| Go foundation/skill/server tests | Passed with isolated writable cache: `GOCACHE=/tmp/jobba-go-cache go test ./internal/server ./internal/skill ./internal/shared ./internal/config` |
| Node and mobile whitespace checks | Passed with `git diff --check` |
| Full Go race suite | Passed after auth/jobs integration: `GOCACHE=/tmp/jobba-go-cache GOPROXY=off go test -race -mod=readonly ./...` |
| Go static/build gates | Passed: `go vet -mod=readonly ./...` and `go build -mod=readonly ./...` with the same offline cache/proxy settings |
| Go payment unit/state tests | Passed as part of the race suite: mismatch rejection, exact fee arithmetic, stable payout reference, ambiguous-transfer tracking, signed-webhook verification, replay no-op, and role rejection |
| Importer unit tests | Passed under the race detector with fixture JSONL covering all supported collections and quarantine/checkpoint failure cases |
| Compose/migration scripts | `bash -n` and `docker compose config --quiet` passed; live execution blocked by Docker daemon socket permissions |
| Database/provider/device checks | Not run: no disposable MongoDB/PostgreSQL, provider sandbox, or Expo runtime is configured |

The Node regression harness directly exercises the current source modules with dependency doubles. It verifies bounded pagination rejection, payout claim exclusion, non-final provider acceptance, ambiguous payout handling, chat receiver membership, and privileged-signup rejection. It does not exercise Express route matching, Mongoose atomicity, Socket.IO transport, or real provider/database behavior.

The mobile script verifies the explicit signup DTO, current nested auth envelope, awaited store method, and malformed session rejection. It replaces the real Zustand/SecureStore implementation with a double, so it does not verify actual persistence ordering, hydration, or cleanup.

## Contract and schema gaps

### P0 before any Go endpoint cutover

1. **Authentication compatibility is not yet end to end.** The app expects `{data:{token,user}}` and routes on `user.accountType`. The Node response supplies that shape, but the mobile `User` type declares `id` and `phone` while Mongo serializes `_id`, `uniqueId`, and `phoneNumber`. Freeze the actual auth fixture and choose explicit public identity/phone fields before the Go auth response is wired.

Resolved during verification: the skill query now selects `COALESCE(legacy_id, id::text)` as the compatibility `_id`, and pagination runs a separate count query before the bounded list query. Handler and pagination tests cover a legacy ID and an empty page beyond the last page.

### P1 before declaring auth persistence fixed

1. **The app stores the same token twice.** `storeToken` writes `authToken`, while Zustand persist writes the session (including `token`) under `auth-storage`. The request interceptor reads only `authToken`; startup routing reads only the Zustand user. Partial write/delete failures can therefore produce contradictory authorization and navigation state.
2. **Awaiting `login()` does not prove the Zustand write finished.** The action awaits `authToken`, then calls `set`; Zustand's async persistence is not awaited by this action. The current script mocks `login()` itself and therefore proves the service awaits its dependency, not that the real persisted session is durable before navigation.
3. **Hydration failure cleanup is incomplete.** The callback clears in-memory state but does not remove a stale standalone `authToken`; subsequent API requests may still carry it.
4. **Mobile validation differs from backend validation.** Login accepts six-character passwords, while the user schema requires eight for created passwords. This does not prevent legacy login, but signup validation and backend rules should be captured separately so error behavior is intentional.

### P1/P2 conversion coverage gaps

- `GET /api/v1/skills/:skillId/workers` exists in Node but is not implemented in Go. It now correctly returns 404 rather than accidentally returning the skill list.
- The Node compatibility presenter returns legacy envelopes for all skill outcomes. Go currently uses a legacy envelope for query-validation failures but a canonical `{error:{code,message}}` envelope for service failures. Decide whether `/api/v1` error parity is required and freeze both fixtures.
- Node exposes `/api/v1/health`; Go exposes `/livez` and `/healthz`. Keep the stronger Go probes, but retain or deliberately retire the old path after consumer/deployment inventory.
- The Node API root returns `{message:"WELCOME TO JOBBA API!!!"}`; Go returns `{data:{service:"jobba-api"}}`. This is low risk unless deployment checks or external consumers use it.
- Node still accepts request bodies up to 1 GB and starts database initialization asynchronously. These previously identified operational defects remain outside the current regression coverage.
- The Go skill repository directly calls pgx while a sqlc query file also exists. This remains a documented temporary blueprint deviation until generated code is available.

## Prioritized remaining slices

1. Finish and independently test Go auth/user/role compatibility: legacy bcrypt variants, public identifiers, status/verification policy, token claims, DTO fixtures, and negative authorization cases.
2. Generate sqlc code, replace the temporary direct pgx skill repository, and test both public skill routes against PostgreSQL fixtures.
3. Collapse mobile session persistence to one authoritative record and test real SecureStore failure permutations, cold hydration, logout/restart, and route guards.
4. Add route-level Node characterization for auth and payment static paths, plus Socket.IO client tests for outsider joins, spoofed events, reconnect, and multiple devices.
5. Build job/application/offer SQL transitions with uniqueness and concurrent acceptance tests before wiring corresponding mobile screens.
6. Build chat/notification persistence and transport only after the Socket.IO compatibility decision is recorded.
7. Run the payment migrations against disposable PostgreSQL, add repository concurrency tests, exercise Paystack sandbox verification/webhooks, and then integrate disputes with the payout hold/release policy.
8. Rehearse data import/reconciliation and cutover/rollback on a protected snapshot before any Go write ownership.

## Release blockers

- No real Mongo-to-PostgreSQL import has been executed or reconciled.
- No full Node Jest run, Expo lint/typecheck, device lifecycle test, or database integration suite has passed in this workspace.
- No verified inventory exists for external REST/Socket.IO consumers or the legacy `/health` path.
- Payment concurrency behavior is covered only by source-level doubles, not by database uniqueness/transactions and a controlled provider.
- Live-data retention, deployment ownership, and maintenance-window requirements remain undecided.
