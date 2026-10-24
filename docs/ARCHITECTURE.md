# Architecture Decisions

## ADR-001: Separate product repository

Jobba Freelance is a separate product workspace. Existing Jobba applications are read-only references during migration. This prevents fiat and handyman assumptions from leaking into the wallet-native product.

## ADR-002: Next.js full-stack boundary

The responsive web UI and HTTP API use one Next.js App Router deployment. Server modules own authentication, authorization, database writes, and transaction construction. The browser owns wallet approval and signature prompts.

## ADR-003: Plain PostgreSQL without Supabase

Wallet signatures establish identity. PostgreSQL stores profiles, marketplace state, sessions, messages, disputes, and indexed chain actions. Chat uses polling over committed rows. The app connects with `pg` from server-only code; Supabase is not used. RLS policies remain as defense in depth for a non-owner role.

## ADR-004: On-chain authority

The Anchor program and USDC token balances are authoritative for escrow. PostgreSQL keeps a queryable projection. A reconciler verifies transactions before applying state and can repair projection drift.

## ADR-005: Funding precedes freelancer acceptance

The client selects a proposal and funds an escrow in `Funded`. The freelancer then accepts on-chain, changing it to `Active`. This makes the program account available for acceptance and preserves a safe client cancellation window before acceptance.

## ADR-006: One stablecoin and one milestone

The MVP prices contracts in six-decimal USDC atomic units. SOL is used only for fees. Multiple tokens, native SOL contracts, and multiple milestones are deferred to protect the critical path.

## Trust boundaries

```text
wallet ──signs──> browser ──session/API──> Next.js server ──SQL──> PostgreSQL
  └──────────────── signed transaction ────────────────────────> Solana
Next.js server ──verify RPC/program state──> Solana ──project──> PostgreSQL
```

The browser is untrusted for roles, amounts, terms hashes, and confirmation claims. Admin UI authorization and program admin authority must both succeed for a dispute decision.
