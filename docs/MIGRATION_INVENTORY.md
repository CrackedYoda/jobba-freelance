# Existing Jobba Migration Inventory

The source applications remain at `/Users/yoda/Desktop/jobba`. This document records what to adapt and what to replace.

## Adapt as specifications

- `Jobba-server-go/internal/job/` — project state and validation concepts.
- `Jobba-server-go/internal/application/` — proposal states and uniqueness behavior.
- `Jobba-server-go/internal/marketplace/` — offer, review, saved-item, and job-context concepts.
- `Jobba-server-go/internal/chat/` and `Jobba-server-go/docs/chat-realtime.md` — participants, message history, unread state, and realtime recovery.
- `Jobba-server-go/internal/dispute/` and `internal/admin/` — evidence, assignment, decision, and audit patterns.
- `Jobba-server-go/db/migrations/000020_jobs.up.sql`, `000021_applications.up.sql`, `000022_marketplace.up.sql`, and both `000040_*` migrations — constraints to translate into a fresh migration sequence.
- `Jobba-App/src/types/marketplace.ts` and `src/services/marketplaceService.ts` — existing payload and endpoint inventory to rebuild as Zod schemas.
- `Jobba-App/components/marketplace/` and `components/dashboard/` — information architecture and user journeys only.

## Build from scratch

- Reown AppKit and WalletConnect integration.
- SIWS challenge verification and cookie sessions.
- SPL USDC amount, mint, ATA, and transfer handling.
- Anchor escrow program and deployment manifest.
- Chain action construction, confirmation, finalization, and reconciliation.
- Responsive DOM components for Next.js.

## Do not copy

- Paystack, debit card, bank account, payout, naira, or webhook implementation.
- JWT/email-password persistence or browser `localStorage` tokens.
- React Native/Expo components as web components.
- `node_modules`, `.expo`, `.git`, lockfiles across stacks, `.env` files, or private keys.
- Floating-point or `numeric(14,2)` representations for USDC.
- Mixed legacy response envelopes and pagination formats.

## Hazards to correct

- The old migration set contains two `000040` prefixes; use a new ordered chain.
- Native `EventSource` cannot set bearer headers; cookie-authenticated Realtime/polling avoids that legacy issue.
- Existing payment state is useful only as lifecycle inspiration and must not become escrow authority.
- Existing worker/handyman terminology must not remain in schema or UI.
