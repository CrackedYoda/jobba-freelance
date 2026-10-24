# Local Development

## Prerequisites

- Node.js 24 and npm 11 or later.
- A Reown project configured with `http://localhost:3000` as an allowed origin.
- Docker, for the local Postgres 16 container `jobba-freelance-db` on port 5434 (5432/5433 are often used by other projects).
- Solana CLI and Anchor are required when work begins on the escrow program.

## Local database

```bash
cd jobba-freelance-frontend
npm run db:up      # docker compose up -d --wait db
npm run db:reset   # drop public schema, apply migrations, apply seed
```

Set `DATABASE_URL=postgres://jobba:jobba@localhost:5434/jobba_freelance` in `.env.local` (already in `.env.example`). To use a hosted Postgres later, change only `DATABASE_URL` and apply `jobba-freelance-backend/database/migrations/*.sql` then `seed.sql` with `psql`.

## Web application

```bash
cd jobba-freelance-frontend
cp .env.example .env.local
npm install
npm run env:check
npm run dev
```

Open `http://localhost:3000`. Empty public values are allowed only for the static landing-page build. Wallet, auth, database, and transaction routes must fail closed when their required configuration is absent.

## Verification

```bash
npm run verify      # env check, lint, typecheck, unit tests, build
npm run test:db     # integration tests against the local Postgres (run db:up and db:reset first)
```

`test:db` covers schema constraints, wallet sessions and nonce replay/expiry, project and proposal repositories, keyset pagination, and duplicate-proposal rejection. CI runs the same suite against a Postgres service container.

Do not use production wallets or private keys. The browser and Vercel environment never need a participant private key.

## Environment groups

- `NEXT_PUBLIC_*` values are visible to browsers. They may contain origins, cluster names, public addresses, and publishable project keys.
- RPC credentials, the database URL, session settings, and reconciliation secret stay server-only.
- The public and server copies of program ID and mint must match. Deployment code also verifies them against the Config PDA.

## Resume procedure

Read `../CURRENT.md`, run `git status`, and execute the last recorded verification command before making changes. Update `CURRENT.md` before changing phases and before stopping.
