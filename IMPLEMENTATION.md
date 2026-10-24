# Jobba Freelance — Full Implementation Specification

**Status:** authoritative build specification
**Deadline:** 48 hours
**Workspace:** `/Users/yoda/Desktop/jobba-freelance`
**Frontend:** Next.js App Router and TypeScript
**Settlement:** USDC on Solana; SOL pays network fees

## 1. Outcome

Deliver a responsive web marketplace that demonstrates both complete paths below on Solana devnet.

```text
Happy path
client signs in with wallet
→ posts project
→ freelancer signs in and proposes
→ client selects proposal and funds USDC escrow
→ freelancer accepts funded contract
→ parties chat
→ freelancer submits work
→ client releases escrow
→ both see confirmed transaction

Dispute path
active funded contract
→ participant opens dispute
→ ordinary release is locked
→ admin reviews terms, messages, delivery, and evidence
→ admin awards freelancer or refunds client
→ decision and confirmed transaction appear in the case history
```

This ordering is final for the MVP. Funding creates the on-chain escrow in `Funded`; freelancer acceptance changes it to `Active`. A client may cancel and recover funds only while the escrow remains `Funded` and undisputed.

## 2. Non-negotiable scope

### Must ship

- Wallet discovery and connection through Reown AppKit/WalletConnect.
- Server-issued, single-use Sign-In With Solana challenge.
- Secure, HttpOnly session cookie; wallet connection by itself is not authentication.
- Client, freelancer, and admin roles.
- Freelancer profiles, skills, rate, bio, availability, and portfolio links.
- Project creation, browsing, search, categories, budgets, and skill filters.
- Proposals with accept, reject, and withdraw behavior.
- One fixed-price contract and one USDC escrow vault per accepted proposal.
- Fund, accept, submit, release, cancel-before-acceptance, dispute, refund, and award transitions.
- Durable contract chat with realtime subscription and polling recovery.
- Admin dispute queue, evidence view, decision form, and audit trail.
- Transaction progress, retry guidance, and Solana Explorer links.
- Deterministic demo seed/reset and a deployable public URL.

### Cut before deadline pressure

Cut these in order: reviews, saved items, presence, typing indicators, file upload, native SOL pricing, split dispute decisions, multiple milestones. Never cut signature verification, escrow authorization, integer money handling, dispute locking, or transaction reconciliation.

## 3. Fixed technical decisions

| Concern | Decision |
|---|---|
| Web | Next.js 16 App Router, React 19, TypeScript |
| Styling | Tailwind CSS; add shadcn components only when used |
| Backend boundary | Next.js route handlers and server-only service modules |
| Database | PostgreSQL (local Docker for now; the user's own hosted Postgres later) |
| Database access | `pg` pool in server-only modules with hand-written SQL; SQL migrations are authoritative |
| Browser database access | None; the browser talks only to route handlers |
| Validation | Zod schemas shared by forms and route handlers |
| Server state | TanStack Query after first API slice |
| Wallet UI | Reown AppKit with Solana adapter |
| Authentication | SIWS/off-chain message fallback plus server session cookie |
| Solana program | Anchor/Rust, PDA-controlled canonical USDC token account |
| Solana web boundary | One adapter module isolates wallet/web3/Anchor types |
| Chat | PostgreSQL messages + polling (SSE/LISTEN-NOTIFY later if time allows) |
| Tests | Vitest for units, Anchor tests for program, Playwright for critical journeys |
| Deployment | Vercel, hosted PostgreSQL, Solana devnet |

Avoid adding a second ORM, Supabase, Redux, a separate API deployment, or a general event bus during the hackathon.

## 4. Repository architecture

```text
jobba-freelance/
├── CURRENT.md
├── IMPLEMENTATION.md
├── README.md
├── docs/
│   ├── FREELANCE_SOLANA_WEBAPP_48H_PLAN.md
│   ├── ARCHITECTURE.md
│   ├── MIGRATION_INVENTORY.md
│   └── reference/
├── jobba-freelance-frontend/
│   ├── app/
│   │   ├── api/v1/
│   │   ├── dashboard/
│   │   ├── projects/
│   │   ├── freelancers/
│   │   ├── contracts/[contractId]/
│   │   ├── messages/
│   │   └── admin/disputes/[disputeId]/
│   ├── components/
│   ├── features/
│   ├── lib/{auth,db,solana,validation}/
│   ├── public/
│   └── tests/
└── jobba-freelance-backend/
    ├── database/migrations/
    ├── programs/freelance_escrow/
    └── docs/
```

Dependency direction is `route/UI → feature service → database or Solana adapter`. UI components never query PostgreSQL directly. Database escrow rows are projections; finalized program accounts and token balances are authoritative.

## 5. Domain language conversion

| Existing Jobba term | Freelance product term |
|---|---|
| Client | Client |
| Worker/handyman | Freelancer |
| Job | Project |
| Application | Proposal |
| Offer/assignment | Contract |
| Work category | Skill category |
| Service fee in naira | Fixed USDC budget |
| Payment/release | Escrow funding/release |
| Job completion evidence | Deliverable submission |

Use the new terms in URLs, database names, UI labels, analytics, and documentation. Compatibility aliases are unnecessary because this is a separate product.

## 6. Core state model

### Project

`draft → open → contracted → completed | cancelled | disputed`

### Proposal

`pending → accepted | rejected | withdrawn`

Accepting a proposal reserves it while the client prepares the funding transaction. Mark it accepted only when funding is confirmed; rejected proposals can then be updated transactionally.

### Contract database projection

`funding_pending → funded → active → submitted → released | cancelled | disputed → resolved`

### Escrow program

`Funded → Active → Submitted → Released`

Alternative terminal branches:

- `Funded → Cancelled` by client before freelancer acceptance.
- `Funded | Active | Submitted → Disputed → Refunded | Awarded` by configured admin resolution.

Every mutation checks actor, current state, contract identifier, mint, amount, and canonical terms hash. No endpoint may skip a transition.

## 7. Wallet authentication

1. Browser connects through Reown AppKit.
2. `POST /api/v1/auth/challenge` receives the public key and returns a random nonce plus exact sign-in message.
3. The message binds domain, URI, chain, wallet, issued-at time, expiry, nonce, and a fixed statement.
4. Wallet signs using SIWS where supported, with `signMessage` compatibility behind one adapter.
5. `POST /api/v1/auth/verify` verifies the exact bytes and public key server-side.
6. The nonce is consumed atomically and cannot be replayed.
7. Server creates a hashed session record and sends a `Secure`, `HttpOnly`, `SameSite=Lax` cookie.
8. Wallet/account/cluster change invalidates client auth state and requires a new challenge.

Challenge TTL: 5 minutes. Session TTL: 24 hours for the demo. Rate-limit challenge and verify routes by IP plus wallet.

## 8. PostgreSQL implementation

Create migrations in this order:

1. Extensions, enums, updated-at helper, and immutable audit helper.
2. Profiles, profile roles, skills, and portfolio links.
3. Projects and project skills.
4. Proposals and unique `(project_id, freelancer_id)` constraint.
5. Contracts, submissions, and canonical terms hash.
6. Conversations, participants, messages, read cursor, and unique contract conversation.
7. Disputes, evidence, decisions, and append-only admin audit.
8. Auth nonces and hashed sessions.
9. Chain actions, signatures, events, reconciliation cursor, and indexes.
10. RLS policies for realtime chat reads.

Money is `bigint` atomic units. USDC uses six decimals; never store authoritative amounts in JavaScript `number`, SQL floating point, or formatted decimal strings. Store public keys, mint, program ID, escrow PDA, token program, signature, slot, block time, commitment, and cluster with every chain action.

All list endpoints use cursor pagination. Mutations accept an `Idempotency-Key`. State changes use row locking or optimistic version checks. A unique action identifier prevents duplicate escrow creation and release.

## 9. Solana escrow program

### Accounts

- Config PDA: admin authority, pending admin, accepted USDC mint, token program, treasury, fee basis points, paused flag, version.
- Escrow PDA: version, contract ID hash, terms hash, client, freelancer, mint, gross amount, fee amount, state, timestamps, bumps.
- Vault: canonical associated token account owned by the escrow PDA.

### Instructions

- `initialize_config`
- `propose_admin` / `accept_admin`
- `set_paused`
- `create_and_fund_escrow`
- `accept_escrow`
- `submit_work`
- `approve_and_release`
- `cancel_before_acceptance`
- `open_dispute`
- `resolve_dispute`

Use checked integer arithmetic. Define whether the platform fee is deducted from the gross award; for the MVP it is deducted on freelancer payout and a full client refund returns the gross amount. The config PDA is the runtime authority for fee and mint values. Recipient/treasury associated-token accounts are created idempotently with the initiating signer paying rent. Close the empty vault and escrow account only after emitting the terminal event and recording the destination for reclaimed rent.

The admin may choose full refund or full award only. No maintenance instruction can transfer escrow funds. Pause blocks new funding while allowing exits and dispute resolutions.

Each demo seed gets a fresh contract UUID so terminal PDAs do not collide across rehearsals.

## 10. Chain transaction lifecycle

For every wallet transaction:

1. Server authorizes the actor and constructs canonical inputs.
2. Server creates a `chain_actions` row with unique action ID and expiry.
3. Server returns a serialized transaction and expected account/amount summary.
4. Browser displays the summary, simulates, asks the wallet to sign, and submits.
5. Browser confirms using signature plus blockhash and last-valid block height.
6. Server independently fetches the transaction and program account at `confirmed`.
7. Reconciler promotes the action to finalized and applies the database projection.
8. UI renders pending, confirmed, finalized, expired, or failed states and permits safe idempotent retry.

Never trust a client-provided signature without verifying its program, instruction, accounts, amount, mint, and action ID.

## 11. API groups

- `/api/v1/auth/{challenge,verify,session,logout}`
- `/api/v1/profiles`, `/api/v1/freelancers`, `/api/v1/skills`
- `/api/v1/projects`, `/api/v1/projects/:id/proposals`
- `/api/v1/proposals/:id/{withdraw,fund}`
- `/api/v1/contracts/:id/{accept,submit,release,cancel}`
- `/api/v1/contracts/:id/transactions/:action`
- `/api/v1/conversations/:id/messages`
- `/api/v1/disputes`, `/api/v1/disputes/:id/evidence`
- `/api/v1/admin/disputes`, `/api/v1/admin/disputes/:id/resolve`
- `/api/v1/chain/actions/:id`, `/api/v1/chain/reconcile`

Each route gets a Zod request/response schema, explicit role matrix, stable error code, and at least one authorization test. Chain-building endpoints return transactions; confirmation/reconciliation endpoints update projections only after RPC verification.

## 12. Responsive interface

Build mobile first at 360px, verify at 768px and 1440px. The required surfaces are landing, project discovery/detail, freelancer profile, role-specific dashboard, project creation, proposal composer/review, contract workspace, chat, wallet/transaction status, and admin dispute workspace.

The contract workspace is the central screen: participants, terms hash, amount, escrow state, next allowed action, deliverable, messages, dispute entry point, and transaction history. On small screens use a bottom action bar; on desktop use a sticky summary rail.

Use familiar marketplace patterns but create original branding, copy, layout, and assets. Do not reproduce Upwork trademarks or exact pages.

## 13. Chat and disputes

One conversation exists per contract. Only client, freelancer, and assigned admin may read it. Messages use server timestamps and `(created_at, id)` cursor order. The browser inserts through authenticated route handlers; Realtime broadcasts committed rows. Poll after three seconds without an event and reconcile optimistic messages by client-generated ID.

Opening a dispute records a category, statement, evidence URLs, current terms hash, submission snapshot, and escrow address. It also submits the on-chain dispute instruction and locks normal release. Admin decisions require a written reason, current on-chain state check, preview of financial outcome, wallet signature, confirmation, and append-only audit record.

## 14. Security requirements

- Keep RPC keys, database service key, migration URL, and reconciliation secret server-only.
- Vercel never receives client or freelancer private keys.
- Keep deployer/admin keys out of Git and local example files.
- Restrict Reown project origins.
- Validate configured program, mint, token program, admin, treasury, and fee against the Config PDA at startup/deployment.
- Enforce CSRF-resistant same-site sessions, strict origin checks on mutations, rate limits, input length limits, and URL allowlists.
- Sanitize rendered user content and never render arbitrary HTML from messages or deliverables.
- Record all admin reads of sensitive dispute evidence and every state-changing action.
- Label the program unaudited and use devnet for the hackathon.

## 15. Verification gates

| Deadline | Gate |
|---|---|
| H2 | Clean install/build, CI, env doctor, wallet UI smoke, frozen transition matrix |
| H6 | Anchor fund/accept/submit/release/dispute tests and generated IDL |
| H10 | SIWS, migrations, terms hash, and transaction builder integrate locally |
| H16 | Complete local happy path |
| H22 | Complete devnet happy path and tested reset/funding runbook |
| H28 | Complete dispute path |
| H34 | Responsive core, durable chat, feature freeze |
| H40 | Clean-clone setup, migration, seed, build, E2E, deployment pass |
| H46 | Rehearsed demo, backup video, fresh demo state |
| H48 | Submission buffer complete |

Required tests cover nonce replay/expiry/domain mismatch, route roles, duplicate proposals, contract races, message participants, all valid and invalid program transitions, wrong mint/ATA/signer/amount, double release, release during dispute, unauthorized admin, overflow, and the two end-to-end demo journeys.

## 16. Execution order

### Hours 0–2

- Finish workspace, environment contract, CI, status log, route shell, and visual tokens.
- Pin Solana/Anchor/Rust/Reown packages after running the compatibility spike.
- Decide exact devnet USDC mint and acquire test funds; otherwise label and use a deterministic Mock USDC mint.

### Hours 2–6

- Build Config/Escrow PDAs and program instructions.
- Test authorization, transitions, integer amounts, and token transfers.
- Export IDL and deployment manifest.

### Hours 6–16

- Apply database migrations and seed.
- Implement SIWS session flow and profiles.
- Implement projects, proposals, contract creation, and transaction builder.
- Finish one complete local economic loop.

### Hours 16–28

- Deploy program to devnet and run funding/release path.
- Add deterministic reset and reconciliation.
- Implement dispute UI, instructions, and admin audit path.

### Hours 28–34

- Add durable chat and required responsive screens.
- Deploy the first complete preview and freeze features.

### Hours 34–48

- Run clean-clone setup and all suites.
- Fix only demo blockers, correctness, responsiveness, and accessibility.
- Seed two demo journeys, rehearse, record a backup, reset, and submit.

## 17. Definition of done

The project is complete when a judge can connect two wallets, authenticate, perform the happy path with real devnet token movement, perform the dispute path with an allowlisted admin, refresh at every step without losing state, see durable chat and transaction history, and use the core pages on mobile and desktop. A new developer must be able to follow the repository docs from a clean clone without private knowledge.

## 18. Resume protocol

`CURRENT.md` is mandatory operational state. Update it:

- before starting a new phase;
- after completing a meaningful unit of work;
- whenever a command, dependency, credential, or external service blocks progress;
- before ending a work session.

It must always say what is being done, what is complete, the last verification result, active blockers, and the exact next command or task.
