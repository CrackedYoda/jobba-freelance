# Jobba freelance escrow backend

This directory contains the unaudited Anchor escrow foundation for the Jobba devnet demo. It handles one fixed-price contract per escrow PDA and accepts only the mint and token program recorded in the singleton Config PDA.

## Economic and state rules

- Amounts are `u64` token atomic units. Six-decimal USDC is formatted only at UI boundaries.
- The platform fee is `floor(gross_amount * fee_bps / 10_000)`, checked for overflow and snapshotted when the escrow is funded.
- A normal release pays `gross - fee` to the freelancer and `fee` to the treasury.
- A dispute refund returns the full gross to the client. An award uses the same freelancer/treasury split as normal release.
- `Funded -> Active -> Submitted -> Released` is the happy path. `Funded -> Cancelled` and `Funded|Active|Submitted -> Disputed -> Refunded|Awarded` are the only alternatives.
- Pause prevents new funding. Existing escrows can still exit or be resolved.
- The canonical vault is the associated token account owned by the escrow PDA. Any surplus tokens sent to it are returned to the client at settlement so a donation cannot prevent closure.

The escrow PDA uses `["escrow", config, contract_id_hash]`. Generate `contract_id_hash` and `terms_hash` as canonical 32-byte hashes server-side; never hash ambiguous concatenated text in the browser.

## Local database

`docker compose up -d --wait db` starts Postgres 16 (`jobba-freelance-db`, port 5434, user/password/db `jobba`/`jobba`/`jobba_freelance`). `database/reset.sh` recreates the schema and applies migrations and the seed. See `database/README.md`.

## Local validation

The isolated toolchains and SBF platform tools are stored outside Git under `.tooling/`. From the repository root, export the paths documented in `../docs/DEVNET_DEPLOYMENT.md`, then run:

```sh
anchor build --no-idl
cargo test -p freelance_escrow --lib
```

The Rust unit tests cover checked fee behavior. `tests/freelance_escrow.ts` records the required local-validator matrix and fixture expectations; remove `describe.skip` once the generated IDL and local mint fixture are wired into CI.

The devnet program ID is `CwNiWw9TSRDnszphzCHsShxC64VhgbawWxCLnZ9gLhxy`. It was deployed on 2026-10-24 in transaction `4xv38JLqJgVqaDTCA6gGTTj8ESF4DJX53aQrowfqZsodmrzX1zpFwCMrz3whpPciS5ewm4LyCZRwVMC4cWV4h5fk`. Its permission-restricted keypair is stored under the Git-ignored `.tooling/keys/` directory. Validate the mint, token program, admin, treasury, and fee against the Config account after initialization. Never commit deployer or admin keys.

## Repository configuration

Copy `.env.example` to `.env` for local operations. The example contains public devnet endpoints and the deployed program address; wallet addresses and the six-decimal mock USDC mint remain intentionally blank. This directory owns PostgreSQL migrations, Docker database configuration, and the Anchor program. The Next.js application lives in the sibling `../jobba-freelance-frontend` directory of the monorepo.
