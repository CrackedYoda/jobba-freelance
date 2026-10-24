# Solana Freelance Marketplace — 48-Hour Implementation Plan

**Status:** implementation blueprint for a separate repository
**Deadline model:** working hackathon submission in 48 hours
**Product direction:** an Upwork-inspired freelance marketplace with wallet-native identity, USDC escrow on Solana, job-linked chat, and administrator dispute resolution
**Suggested folder name:** `Jobba-freelance-web`

## 1. Product decision

Build a new responsive web application rather than modifying the completed Jobba mobile or Go projects. Reuse their marketplace lessons—authorization, explicit state machines, idempotency, dispute evidence, and compatibility testing—but do not copy their code or handyman-oriented terminology.

The hackathon demo must prove one complete economic loop:

```text
Client connects wallet
  → posts a freelance project
  → freelancer connects wallet and submits a proposal
  → client accepts proposal and funds USDC escrow on Solana
  → freelancer accepts the funded contract
  → both parties chat
  → freelancer submits a deliverable
  → client releases USDC
```

It must also prove a dispute branch:

```text
Funded contract
  → either party opens dispute
  → funds are locked from normal release
  → admin reviews terms, chat, submission, and evidence
  → admin refunds client or awards freelancer
  → result and Solana transaction are visible to both parties
```

### Currency scope

- **USDC is the contract currency and escrow asset.** It gives clients and freelancers stable pricing.
- **SOL pays Solana transaction fees** and is displayed in the wallet panel.
- **Native-SOL contract payments are a stretch goal only after the USDC path is fully tested.** Supporting both native SOL and SPL-token vaults doubles the on-chain account and test surface.
- The app must never call a custom token “USDC.” Use Circle’s configured devnet USDC mint for the hosted demo, or clearly label a local-validator mint as `Mock USDC`.

## 2. Hackathon scope boundary

### Must ship

- Responsive landing, marketplace, dashboards, contract workspace, chat, and admin console.
- Reown AppKit, powered by WalletConnect, for Solana wallet discovery and connection.
- Server-issued, single-use Sign In With Solana challenge and signature verification.
- Wallet-based onboarding with `client`, `freelancer`, or both modes.
- Freelancer profiles: title, bio, skills, rate, portfolio links, availability.
- Client project creation.
- Project browsing, search, categories, budgets, and skill filters.
- Proposal submission, review, accept/reject/withdraw.
- One fixed-price milestone per contract for the judged demo.
- Real USDC funding into a program-controlled escrow vault.
- Freelancer contract acceptance and work submission.
- Client release.
- Either-party dispute creation and response.
- Admin full refund or full freelancer award.
- Persistent one-to-one contract chat.
- Transaction confirmation states and Solana Explorer links.
- Seed/reset command and deterministic demo data.
- Deployable web app, API, database migrations, and Solana program.

### Stretch only

- Several milestones per contract.
- Split dispute awards.
- Native SOL-denominated contracts.
- File uploads; use validated URLs first.
- Ratings and reviews.
- Saved projects and saved freelancers.
- Typing indicators and online presence.
- Email or wallet push notifications.
- Platform fee configuration UI.

### Explicitly defer

- Hourly tracking and timesheets.
- Agencies, teams, and multiple freelancers per contract.
- Fiat cards, bank accounts, naira conversion, Paystack, and local payout rails.
- Cross-chain deposits, swaps, bridges, and arbitrary tokens.
- KYC, sanctions screening, tax forms, and production compliance automation.
- DAO arbitration, appeals, multiple arbitrators, and decentralized juries.
- Automatic deadline release/refund.
- Contract amendments after funding.
- Production mainnet launch or claims of audited escrow.

## 3. Technical architecture

### Recommended stack

| Layer | Choice | Why it fits 48 hours |
|---|---|---|
| Web | Next.js App Router + TypeScript | Responsive UI, server routes, SSR, and one deployable project |
| Styling | Tailwind CSS + shadcn/ui + Lucide | Fast, consistent marketplace UI without building a design system |
| Wallet | Reown AppKit + Solana adapter | WalletConnect network plus Solana account and transaction support |
| Authentication | Sign In With Solana + secure server session cookie | Wallet ownership becomes the identity without passwords |
| Forms | React Hook Form + Zod | Shared validation and fast form implementation |
| Server state | TanStack Query | Transaction retries, optimistic chat, loading/error states |
| Database | Hosted PostgreSQL, preferably Supabase | Fast provisioning, SQL constraints, migrations, and optional Realtime |
| Database access | Drizzle ORM or Supabase server client | Typed queries and migration-friendly schema |
| Chat | PostgreSQL messages + Supabase Realtime; 3-second polling fallback | Durable history with a low-effort realtime path |
| On-chain program | Anchor/Rust | Fast Solana PDA, token CPI, events, and program tests |
| Solana client | `@solana/kit`/Wallet Standard-compatible APIs plus Anchor client where needed | Modern Solana primitives and message verification |
| Deployment | Vercel + Supabase + Solana devnet | Fast preview deployments and a public judging URL |
| Tests | Vitest + Playwright + Anchor tests | Unit, journey, and program-state coverage |

Reown AppKit currently supports Solana in React/Next.js and provides WalletConnect wallet discovery. A Reown project ID must be restricted with an origin allowlist. Wallet connection alone is not login; the backend must verify a signed challenge. See [Reown AppKit overview](https://docs.reown.com/appkit/overview) and [Reown SDK capabilities](https://docs.reown.com/web3inbox/frontend-integration/android/core/installation).

### System diagram

```text
Browser / Next.js UI
├── Reown AppKit → Solana wallet
├── SIWS signature → Next.js auth route
├── REST/route handlers → marketplace services
├── Realtime subscription/polling → chat service
└── signed transaction → Solana RPC

Next.js server
├── session and nonce service
├── project/proposal/contract service
├── chat and dispute service
├── Solana transaction/indexing service
└── PostgreSQL repositories

PostgreSQL
├── profiles, skills, projects, proposals
├── contracts, submissions, conversations, messages
├── disputes and admin audit log
├── chain actions and indexed events
└── SIWS nonces and sessions

Solana program
├── Config PDA
├── Escrow PDA
├── escrow-owned canonical USDC ATA
├── fixed authorization and state transitions
└── versioned events
```

### Repository structure

```text
Jobba-freelance-web/
├── app/
│   ├── (public)/
│   │   ├── page.tsx
│   │   ├── projects/page.tsx
│   │   ├── projects/[projectId]/page.tsx
│   │   └── freelancers/[profileId]/page.tsx
│   ├── (app)/
│   │   ├── dashboard/page.tsx
│   │   ├── projects/
│   │   ├── proposals/
│   │   ├── contracts/[contractId]/page.tsx
│   │   ├── messages/
│   │   ├── profile/page.tsx
│   │   └── wallet/page.tsx
│   ├── admin/
│   │   ├── page.tsx
│   │   └── disputes/[disputeId]/page.tsx
│   └── api/v1/
│       ├── auth/
│       ├── projects/
│       ├── proposals/
│       ├── contracts/
│       ├── conversations/
│       ├── disputes/
│       └── chain/
├── features/
│   ├── auth/
│   ├── profiles/
│   ├── projects/
│   ├── proposals/
│   ├── contracts/
│   ├── escrow/
│   ├── chat/
│   ├── disputes/
│   └── admin/
├── components/
│   ├── ui/
│   ├── layout/
│   ├── wallet/
│   └── marketplace/
├── server/
│   ├── auth/
│   ├── db/
│   ├── repositories/
│   ├── services/
│   ├── solana/
│   └── validation/
├── programs/
│   └── freelance_escrow/
├── database/
│   ├── migrations/
│   └── seed.ts
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── docs/
│   ├── architecture.md
│   ├── escrow.md
│   └── demo-script.md
└── .env.example
```

Dependency direction:

```text
pages/routes → feature services → repositories/adapters
                               ↘ PostgreSQL
                               ↘ Solana RPC/program
```

Components do not query PostgreSQL directly. The Solana index is a projection; confirmed program accounts and token balances remain the authority for escrow state.

## 4. Wallet signup and authentication

Solana’s current guidance recommends readable off-chain messages where supported, with `signMessage` fallback. A signed message proves wallet control without a fee or transaction. See [Solana message authentication](https://solana.com/docs/frontend/messages-and-auth).

### Authentication flow

1. User selects **Connect wallet**.
2. Reown AppKit opens WalletConnect-compatible wallet selection.
3. Browser receives the selected Solana address and cluster.
4. Browser calls `POST /api/v1/auth/challenge`.
5. Server generates a cryptographically random nonce and stores only its hash with:
   - expected wallet when known;
   - exact domain and URI;
   - cluster;
   - issued and expiry times;
   - `used_at = null`.
6. Prefer Wallet Standard `solana:signIn` when advertised. Otherwise sign the exact server message with `signMessage`.
7. Browser sends the address, challenge ID, signature, and exact signed bytes to `POST /api/v1/auth/verify`.
8. Server verifies Ed25519 signature, address, domain, URI, cluster, nonce, expiry, and unused state.
9. In one transaction, server consumes the nonce and creates/rotates the session.
10. Server sets a short-lived `HttpOnly`, `Secure` in production, `SameSite=Lax` session cookie.
11. New wallets enter onboarding; existing wallets load their profile and modes.

Example statement:

```text
Jobba Freelance wants you to sign in with your Solana wallet.
Signing in costs nothing and does not submit a transaction.
```

### Rules

- Never treat a connected wallet as authenticated before signature verification.
- Never accept a nonce created by the browser.
- Consume each nonce once, including under concurrent verification.
- Bind session to wallet and cluster; rotate on account/network changes.
- Signing login must never request token approval or transfer authority.
- Admin access uses an environment allowlist of wallet addresses for the hackathon.
- A wallet can enable client and freelancer modes; it cannot self-assign admin.

## 5. Freelance marketplace model

Replace handyman concepts throughout:

| Old concept | New concept |
|---|---|
| Worker | Freelancer |
| Client job for local work | Remote or hybrid freelance project |
| Onsite task | Fixed-price digital contract |
| Work type | Category, skills, experience level, delivery time |
| Assignment | Accepted proposal and contract |
| Payment release | On-chain milestone release |
| Evidence | Deliverable manifest and dispute evidence |

Suggested categories:

- Web and mobile development
- UI/UX and graphic design
- Writing and translation
- Video and animation
- Marketing and growth
- Data and AI
- Blockchain development
- Virtual assistance

### Off-chain state machines

```text
Project:
Draft → Open → Hired → Completed
          ↓       ↓
      Cancelled  Disputed

Proposal:
Submitted → Accepted
    ↓           ↓
Withdrawn     Contract created
    ↓
Rejected

Contract:
Awaiting freelancer acceptance → Active → Work submitted → Completed
                  ↓                 ↓              ↓
               Declined          Disputed ←───────┘

Dispute:
Open → Counterparty responded → Under review → Resolution submitted
                                             → Resolved on-chain
```

The database must reject invalid transitions even when the UI hides the relevant button.

## 6. PostgreSQL data model

Use UUID primary keys, UTC timestamps, explicit check constraints, and unique business keys.

### `profiles`

- `id`
- `wallet_address` unique
- `display_name`
- `avatar_url`
- `headline`
- `bio`
- `country`
- `timezone`
- `hourly_rate_usdc_minor` optional display metadata
- `client_enabled`
- `freelancer_enabled`
- `status`: `ACTIVE | SUSPENDED`
- timestamps

### `skills` and `profile_skills`

- normalized unique skill names
- unique `(profile_id, skill_id)`

### `projects`

- client profile
- title, description, category
- experience level
- budget in USDC minor units
- delivery expectation
- visibility and status
- timestamps

### `project_skills`

- unique `(project_id, skill_id)`

### `proposals`

- project and freelancer
- cover letter
- proposed USDC minor amount
- delivery days
- status
- unique `(project_id, freelancer_id)`

### `contracts`

- project, client, freelancer, accepted proposal
- terms version and canonical terms hash
- gross USDC minor amount
- fee BPS and frozen fee amount
- escrow PDA and vault ATA
- cluster, program ID, mint
- state
- accepted timestamps
- unique project/accepted proposal

### `submissions`

- contract
- note
- canonical deliverable manifest JSON
- manifest SHA-256
- timestamps

### `conversations` and `messages`

- one conversation for project/proposal/contract participants
- text-first messages
- optional validated attachment URL
- participant-only database policy
- cursor pagination

### `disputes`, `dispute_evidence`, and `admin_audit_log`

- contract and escrow reference
- initiator/counterparty
- reason and statements
- evidence URL plus evidence hash
- on-chain resolution choice/signature
- immutable admin audit entries

### `chain_actions` and `chain_events`

`chain_actions` stores client action IDs, requested transaction, signature, blockhash, last valid height, confirmation state, and error. Enforce unique `(cluster, wallet_address, action_id)`.

`chain_events` stores decoded program events uniquely by `(cluster, signature, event_index)` and records commitment.

### `auth_nonces` and `sessions`

- hashed nonce only
- exact expected SIWS fields
- expiry and atomic consumption
- hashed session token or framework-managed secure session

## 7. Solana escrow program

Solana tokens live in token accounts. USDC uses the original SPL Token Program, and the correct token program must be included when deriving associated token accounts. See [Solana token accounts](https://solana.com/docs/tokens) and [Solana payments](https://solana.com/docs/payments/how-payments-work).

### Immutable deployment manifest

```text
cluster
program_id
rpc_http_url
rpc_ws_url
explorer_base_url
usdc_mint
token_program_id
platform_treasury_owner
admin_authority
fee_bps
deployment_version
```

Do not accept the mint, token program, treasury, or admin from request input.

### Config PDA

```text
seeds = ["config"]
version
admin
pending_admin
usdc_mint
token_program
treasury_owner
fee_bps
paused
bump
```

### Escrow PDA

For the MVP, one contract equals one on-chain milestone:

```text
seeds = ["escrow", client_pubkey, contract_id_hash]
version
client
freelancer
contract_id_hash
terms_hash
mint
gross_amount
fee_amount
funded_at
accepted_at
submitted_at
evidence_hash
state
bump
```

Use the escrow PDA as token authority. Its canonical USDC ATA is the vault. The program signs token transfers with PDA seeds.

### Program instructions

#### `initialize_config`

- Deployment authority only.
- Stores exact mint, token program, admin, treasury, and bounded fee BPS.

#### `create_and_fund_escrow`

- Client signs.
- Deterministic PDA prevents a second escrow for the same client/contract.
- Creates vault ATA idempotently.
- Freezes freelancer, amount, fee, mint, and terms hash.
- Uses `transfer_checked` with configured decimals.
- Transfers full gross amount atomically.
- Emits `EscrowFunded`.

#### `accept_escrow`

- Freelancer signs once.
- Before acceptance, client may cancel/refund.
- After acceptance, client cannot unilaterally refund.

#### `submit_work`

- Freelancer only.
- `Active → Submitted`.
- Stores the hash of the off-chain deliverable manifest.

#### `approve_and_release`

- Client only after submission.
- Atomically transfers net amount to freelancer ATA and fee to treasury ATA.
- State becomes terminal.

#### `open_dispute`

- Client or freelancer.
- Only funded/active/submitted states.
- Locks normal release/refund paths.
- Stores evidence manifest hash.

#### `resolve_dispute`

- Configured admin only.
- Only a disputed escrow.
- MVP choice: full client refund or full freelancer award.
- Stretch choice: checked client/freelancer split whose sum equals gross amount.
- Recipients are stored parties and configured treasury; the admin cannot redirect funds.

#### `cancel_before_acceptance`

- Client only before freelancer acceptance.
- Full refund and terminal state.

#### `set_paused`, `propose_admin`, `accept_admin`

- Minimal maintenance controls.
- Never add an arbitrary admin drain instruction.

### On-chain state machine

```text
Funded → Active → Submitted → Released
   ↓        ↓         ↓
Cancelled  Disputed ←─┘
              ↓
        Refunded / Awarded
```

Every money-moving transition and both token transfers commit atomically.

### Amount rules

- Store USDC as six-decimal integer base units.
- Parse decimal input without JavaScript floating point.
- Reject zero, negative, more than six decimals, and overflow.
- Freeze fees when funding occurs.
- Use checked arithmetic in the program.

## 8. Chain transaction lifecycle

```text
Draft action
→ fetch fresh blockhash + lastValidBlockHeight
→ derive and verify accounts
→ build transaction
→ simulate
→ wallet reviews and signs
→ submit
→ persist signature immediately
→ confirm using the original blockhash tuple
→ mark confirmed
→ mark finalized after finality
→ reconcile program state and indexed event
```

Rules:

- A returned signature is not success.
- On timeout, query signature status and escrow account before rebuilding.
- If a blockhash expires, rebuild only after proving the state did not transition.
- A retry with the same action ID returns the existing chain action.
- Program state, not a browser callback, decides whether funds moved.
- Display `Awaiting wallet`, `Submitted`, `Confirmed`, `Finalized`, and `Failed` distinctly.

## 9. API surface

```text
POST   /api/v1/auth/challenge
POST   /api/v1/auth/verify
POST   /api/v1/auth/logout
GET    /api/v1/me
PATCH  /api/v1/me

GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/:projectId
PATCH  /api/v1/projects/:projectId
POST   /api/v1/projects/:projectId/publish

POST   /api/v1/projects/:projectId/proposals
GET    /api/v1/projects/:projectId/proposals
GET    /api/v1/proposals/mine
POST   /api/v1/proposals/:proposalId/withdraw
POST   /api/v1/proposals/:proposalId/accept

GET    /api/v1/contracts
GET    /api/v1/contracts/:contractId
POST   /api/v1/contracts/:contractId/accept
POST   /api/v1/contracts/:contractId/funding-transaction
POST   /api/v1/contracts/:contractId/submit
POST   /api/v1/contracts/:contractId/release-transaction
POST   /api/v1/contracts/:contractId/sync

GET    /api/v1/conversations
GET    /api/v1/conversations/:conversationId/messages
POST   /api/v1/conversations/:conversationId/messages
POST   /api/v1/conversations/:conversationId/read

POST   /api/v1/contracts/:contractId/disputes
GET    /api/v1/disputes/:disputeId
POST   /api/v1/disputes/:disputeId/respond
POST   /api/v1/disputes/:disputeId/evidence

GET    /api/v1/admin/disputes
GET    /api/v1/admin/disputes/:disputeId
POST   /api/v1/admin/disputes/:disputeId/resolution-transaction
POST   /api/v1/admin/disputes/:disputeId/sync

GET    /api/v1/chain/actions/:actionId
POST   /api/v1/chain/reconcile
```

Use a consistent envelope and machine-readable codes:

```json
{
  "data": {},
  "error": null,
  "meta": {}
}
```

## 10. Responsive UX and pages

Use Upwork’s marketplace mental model, not its name, logo, copy, colors, or exact visual design.

### Public pages

- Landing page with one promise: hire or earn using on-chain USDC escrow.
- Browse projects.
- Project detail.
- Public freelancer profile.
- How escrow works.

### Client workspace

- Dashboard.
- Post/edit project.
- My projects.
- Proposal review.
- Contract workspace.
- Messages.
- Wallet and transaction activity.

### Freelancer workspace

- Recommended/browse projects.
- Proposal composer.
- My proposals.
- Active contracts.
- Work submission.
- Messages.
- Wallet and earnings activity.

### Admin workspace

- Open dispute queue.
- Case workspace.
- User/project status lookup.
- Audit history.

### Contract workspace

This is the central screen:

- contract parties and agreed terms;
- escrow status and USDC amount;
- chain transaction timeline;
- deliverable submission;
- chat panel/link;
- dispute actions;
- Explorer links;
- sticky primary action matched to the current state.

### Admin dispute workspace

- Case ID, amount, escrow PDA, and chain status.
- Client and freelancer profiles/wallets.
- Immutable contract terms.
- Chronological chat and transaction timeline.
- Submission and evidence grouped by party.
- Resolution note.
- Explicit preview of recipient and amount.
- Admin wallet confirmation.
- Locked final resolution and Explorer link.

### Responsive targets

- 360 px mobile.
- 768 px tablet.
- 1280/1440 px desktop.
- Sticky bottom primary action on mobile.
- Filter drawer on mobile, sidebar on desktop.
- Messages use list-to-thread navigation on mobile and split view on desktop.
- No horizontal overflow in wallet, contract, or dispute views.

## 11. Chat plan

- Persistent PostgreSQL message history is authoritative.
- Only project/contract participants can list or send.
- Subscribe only to conversations the session belongs to.
- Use Supabase Realtime for message inserts if available by Hour 24.
- Refetch on reconnect, focus, and visibility changes.
- Fall back to bounded polling every three seconds without changing the UI.
- Text messages only for the core submission.
- Deliverable and evidence URLs are separate records, not trusted HTML in chat.
- Defer typing, presence, edits, deletes, reactions, calls, and group chat.

## 12. Admin and dispute security

- Admin wallets come from immutable deployment configuration/environment allowlist.
- The database may record the admin profile, but it cannot grant admin authority by itself.
- The on-chain program independently verifies the configured admin public key.
- The admin can act only on `Disputed` escrows.
- Destinations are the stored client/freelancer/treasury accounts.
- Resolution writes an off-chain audit record and an on-chain event.
- UI must clearly disclose centralized hackathon arbitration.
- No admin drain, arbitrary recipient, silent terms edit, or resolution rollback.

## 13. Security and correctness tests

### Wallet/session tests

- Invalid, altered, expired, and reused nonce.
- Wrong domain, URI, cluster, wallet, or signature.
- Concurrent verification of the same nonce.
- Account or network switch.
- Non-admin wallet accessing admin routes.

### Marketplace tests

- Only clients create/publish projects.
- Only freelancers propose.
- One proposal per freelancer/project.
- Only project owner accepts.
- Competing proposal acceptance race produces one winner.
- Conversation access is participant-only.
- Invalid state transitions fail server-side.

### Program tests

- Wrong signer for every instruction.
- Fake mint or wrong token program.
- Noncanonical vault/destination ATA.
- Wrong PDA seeds, client, freelancer, or terms hash.
- Duplicate funding/release/refund/resolution.
- Release before submission.
- Cancel after acceptance.
- Dispute after release.
- Admin resolution without dispute.
- Arbitrary recipient substitution.
- Zero/overflow amounts and fee boundaries.
- Underfunded vault.
- Pause behavior.
- Every competing state-transition race.

### Journey tests

1. Client connects, signs in, creates project.
2. Freelancer connects, signs in, proposes.
3. Client accepts; freelancer accepts.
4. Client funds real devnet USDC escrow.
5. Parties exchange messages.
6. Freelancer submits; client releases.
7. Separate seeded escrow enters dispute; admin refunds or awards.

## 14. Forty-eight-hour execution plan

### Team workstreams

Use three parallel tracks with integration checkpoints. One person may cover multiple tracks if necessary.

- **Track A — Solana:** program, wallet transactions, event decoding.
- **Track B — Marketplace:** Next.js, auth, database, projects/proposals/contracts.
- **Track C — UX and operations:** responsive UI, chat, admin, tests, deployment, demo.

### Hours 0–2: freeze scope

- Create separate repository/folder.
- Add README, environment example, CI skeleton, lint/typecheck.
- Register Reown project and origin allowlist.
- Provision Supabase/PostgreSQL and Vercel projects.
- Create three devnet wallets: client, freelancer, admin.
- Confirm official devnet USDC availability and prefund SOL fees.
- Freeze one-milestone MVP and demo script.

**Gate:** no secondary features until one wallet can connect, sign a SIWS challenge, and sign a devnet transaction.

### Hours 2–6: proof of hardest path

**Track A**

- Scaffold Anchor program.
- Implement config plus create/fund and release instructions.
- Prove transfer into PDA-controlled vault on local validator.

**Track B**

- Implement SIWS nonce/session flow.
- Create profiles, projects, proposals, and contracts schema.
- Seed skills and demo profiles.

**Track C**

- Build responsive shell, wallet button, role switcher, design tokens.
- Draft landing, browse, dashboard, and contract workspace layouts.

**H6 gate:** local-validator USDC fund and release test passes. If it does not, stop UI expansion and fix escrow.

### Hours 6–14: marketplace vertical slice

- Project list, filters, detail, and create form.
- Freelancer profile and proposal form.
- Client proposal list and accept action.
- Atomic accept creates contract/conversation and rejects competing proposals.
- Freelancer contract acceptance.
- Role-aware dashboards.
- Program negative tests for wrong signer/mint/account/state.

### Hours 14–22: real escrow integration

- Build and simulate funding transaction.
- Wallet signs and submits.
- Persist action/signature/blockhash tuple.
- Confirm and reconcile escrow state.
- Contract workspace shows exact amount, parties, mint, and Explorer link.
- Implement submission and release transaction.
- Index program events idempotently.
- Complete devnet happy path once.

**H22 gate:** client-to-freelancer real USDC path works end-to-end on devnet.

### Hours 22–30: chat and disputes

- Persistent participant-only chat.
- Realtime subscription or polling fallback.
- Submission note and deliverable URL.
- Open dispute and counterparty response.
- Admin dispute queue and case workspace.
- On-chain dispute and admin resolution.
- Seed one already-disputed contract for the presentation.

### Hours 30–38: integration and responsive polish

- Mobile navigation and sticky actions.
- Loading, empty, wallet rejection, wrong-network, pending, timeout, and retry states.
- Reconciliation after refresh.
- 360/768/1440 px checks.
- Accessibility labels, focus, contrast, and keyboard navigation.
- Deterministic reset/seed command.
- First hosted deployment.

### Hours 38–43: freeze and verify

- Feature freeze at H38.
- Full program tests.
- Unit/integration tests.
- Playwright happy and dispute paths.
- Fresh database migration and seed rehearsal.
- Fresh Vercel preview deployment.
- Recheck Reown allowlist and environment variables.

### Hours 43–46: demo rehearsal

- Run happy path twice.
- Run dispute path twice.
- Confirm wallet SOL/USDC balances.
- Confirm Explorer links.
- Test separate browser profiles for client, freelancer, and admin.
- Record fallback demo video.
- Capture architecture diagram and screenshots.

### Hours 46–48: submission buffer

- Critical fixes only.
- Reset demo state and wallets.
- Tag judged commit.
- Verify public URL in a private browser.
- Submit repository, URL, video, screenshots, and known limitations.

## 15. Critical path and cut order

Critical path:

```text
wallet connection
→ verified SIWS session
→ project/proposal acceptance
→ deterministic escrow PDA
→ real USDC funding
→ submission
→ release
→ dispute lock
→ admin resolution
→ hosted rehearsal
```

If behind, cut in this order:

1. Attachments.
2. Ratings and reviews.
3. Saved items and notifications page.
4. Multiple milestones.
5. Native SOL project payments.
6. Realtime transport; retain polling.
7. Profile editing; retain seeded profiles.

Never cut wallet signature verification, real USDC escrow, proposal acceptance, contract state authorization, chat persistence, admin resolution, responsive core screens, or deploy/rehearsal time.

## 16. Demo preparation

Prepare and prefund:

- client wallet with devnet SOL and at least 150 devnet USDC;
- freelancer wallet with devnet SOL and empty/known USDC starting balance;
- admin wallet with devnet SOL;
- treasury ATA;
- one open project;
- one disputed funded escrow for the admin chapter.

### Three-minute demo

1. Connect client wallet and sign in.
2. Show published “Build a Solana analytics dashboard” project.
3. Switch to freelancer browser; submit a 90 USDC proposal.
4. Client accepts the proposal and funds escrow; show Solana Explorer.
5. Freelancer accepts the funded contract.
6. Freelancer chats and submits a prepared deliverable URL.
7. Client releases; show freelancer USDC balance increase.
8. Admin opens the seeded dispute, reviews the case, and refunds/awards on-chain.
9. End on architecture view: marketplace data off-chain, authorization and funds on-chain.

## 17. Acceptance criteria

### Identity

- WalletConnect-compatible wallet connects through Reown AppKit.
- Backend verifies a single-use SIWS signature.
- Refresh restores a secure session.
- Wallet/network changes invalidate or rotate the session.

### Marketplace

- Client publishes a project.
- Freelancer submits one proposal.
- Client accepts exactly one proposal under concurrent attempts.
- Both parties see identical contract terms.

### Escrow

- Funding moves configured USDC into the canonical escrow vault.
- UI marks funded only after confirmation.
- Freelancer cannot submit before acceptance/funding.
- Client cannot release before submission.
- Release sends net funds and fee atomically.
- Refresh/retry cannot duplicate payment.

### Disputes

- Either party can dispute a nonterminal funded contract.
- Normal release is blocked while disputed.
- Only configured admin resolves.
- Admin cannot redirect funds.
- Both parties see resolution note and Explorer transaction.

### Chat and responsive UI

- Only participants read/send messages.
- New messages appear through realtime or within three seconds.
- Core journeys work at 360, 768, and 1440 px.
- No payment/wallet page overflows horizontally.

### Submission

- Public URL loads in a clean browser.
- Program ID, USDC mint, admin wallet, and network are visible in documentation.
- Tests, typecheck, lint, build, migrations, seed, and E2E smoke pass.
- Known limitations explicitly say devnet prototype and unaudited program.

## 18. Environment variables

```dotenv
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_REOWN_PROJECT_ID=
NEXT_PUBLIC_SOLANA_CLUSTER=devnet
NEXT_PUBLIC_SOLANA_RPC_URL=
NEXT_PUBLIC_SOLANA_WS_URL=
NEXT_PUBLIC_SOLANA_EXPLORER_URL=https://explorer.solana.com
NEXT_PUBLIC_ESCROW_PROGRAM_ID=
NEXT_PUBLIC_USDC_MINT=

DATABASE_URL=
SESSION_SECRET=
SIWS_DOMAIN=localhost:3000
SIWS_URI=http://localhost:3000
ADMIN_WALLET_ADDRESSES=
PLATFORM_TREASURY_OWNER=
PLATFORM_FEE_BPS=0

SUPABASE_URL=
SUPABASE_ANON_KEY=
SUPABASE_SERVICE_ROLE_KEY=
DEMO_MODE=true
```

Public wallet/program configuration is not a secret, but must be validated against the deployment manifest. Server, database, and service-role secrets must never enter client bundles.

## 19. Main risks

| Risk | Response |
|---|---|
| Escrow program consumes the schedule | Prove fund/release by H6; keep one milestone and USDC only |
| Devnet USDC or RPC instability | Prefund early, configure backup RPC, keep local-validator test, record backup demo |
| Off-chain and on-chain state diverge | Reconcile signature plus program account; chain is authority |
| Wallet rejects or switches account | Preserve form state, clear session, provide retry |
| Duplicate transactions | Deterministic PDA, terminal states, action IDs, reconcile before rebuild |
| Realtime chat delays core work | Use three-second polling fallback |
| Admin power appears unsafe | Store admin on-chain, restrict resolution recipients, disclose centralization |
| Evidence leaks private data | Store URLs/private objects off-chain; only canonical hash on-chain |
| “Escrow” implies production trust | Label prototype unaudited and devnet-only |
| Copying Upwork creates brand/IP risk | Copy interaction patterns, not branding, text, assets, or exact visual design |

## 20. Final recommendation

The strongest hackathon submission is not the broadest clone. It is a polished freelance workflow with one technically credible differentiator:

> Marketplace negotiation and chat happen off-chain, while a Solana program—not the app server—controls USDC release and dispute resolution.

Commit to one fixed-price USDC milestone, three roles, persistent chat, and a transparent admin dispute transaction. Get funding and release working in the first six hours. Every additional feature is secondary to proving that flow safely and explaining it clearly.
