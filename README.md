# Jobba Freelance

Jobba Freelance is a responsive, wallet-native freelance marketplace for a 48-hour hackathon. Clients post projects and fund fixed-price contracts in USDC on Solana. Freelancers submit proposals, chat with clients, deliver work, and receive escrowed funds. Administrators can review evidence and resolve disputes.

## Repository map

- `jobba-freelance-frontend/` — Next.js App Router web application and server route handlers.
- `jobba-freelance-backend/` — PostgreSQL migrations, the Anchor escrow program, and backend operational material.
- `docs/` — architecture, implementation, security, and reference documents.
- `IMPLEMENTATION.md` — authoritative delivery checklist and decisions.
- `CURRENT.md` — live handoff state; read this first when resuming work.

## Start here

1. Read `CURRENT.md`.
2. Read `IMPLEMENTATION.md`.
3. Copy `jobba-freelance-frontend/.env.example` to `jobba-freelance-frontend/.env` and fill the required configuration.
4. From `jobba-freelance-backend`, run `docker compose up -d --wait db` and `bash database/reset.sh --demo`.
5. From `jobba-freelance-frontend`, run `npm ci` and `npm run dev`. Run `npm run test:db` for the database integration tests.

The local MVP includes wallet authentication, marketplace workflows, PostgreSQL persistence, USDC escrow transaction construction/reconciliation, chat, submissions, and administrator dispute resolution. See `docs/DEVNET_DEPLOYMENT.md` for the remaining hosted configuration and devnet release steps.

## Product scope

The MVP supports one fixed-price USDC escrow per contract. SOL pays network fees. Fiat cards, bank accounts, naira settlement, hourly contracts, agencies, split awards, and multiple milestones are outside the 48-hour scope.

## Reference code policy

The historical source projects are retained under `docs/reference/source/` for reference. Do not commit dependency folders, secrets, private keys, Paystack code, bank models, or floating-point money representations.
