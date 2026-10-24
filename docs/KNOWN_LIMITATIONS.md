# Known Limitations

## Deployment gate

The local MVP includes marketplace writes, wallet authentication, contract workspaces, transaction construction and reconciliation, chat, submissions, participant disputes, administrator resolution UI, and a successful SBF build. It is not a deployed hackathon submission until devnet deployment, Config PDA initialization, the real wallet matrix, and hosted environment checks pass.

## Native toolchain

Rust, Anchor 0.30.1, Agave 4.2.2, and SBF platform-tools v1.58 are installed. The program's four host tests and SBF build pass. Anchor 0.30 emits legacy macro `cfg` warnings, but the corrected build emits no stack overflow diagnostic. The Git-ignored devnet deployer exists but has 0 SOL because the public faucet rejected both attempted airdrops.

## Wallet dependency advisories

The current Reown Solana adapter transitively uses Solana web3.js 1.x and packages with published npm advisories. `npm audit --omit=dev` reports eleven findings in that dependency graph. npm proposes no compatible automatic fix. The application remains devnet-only; recheck upstream releases before any mainnet work.

## Development server in this sandbox

Turbopack development mode cannot start its macOS FSEvent stream in this sandbox, and webpack watch mode reaches the sandbox file-descriptor limit. Production builds and `next start` work. Normal development outside the sandbox can use `npm run dev`.

## Authentication fallback

Development without `DATABASE_URL` uses an in-memory nonce/session adapter for one local process. Production fails closed without PostgreSQL. Persistent sessions store SHA-256 token hashes, and challenges are consumed through conditional updates.

## Devnet configuration

The program identity is fixed locally, but the mint, administrator, treasury, and fee remain deployment inputs. Select or create a clearly labeled six-decimal devnet Mock USDC mint and ensure the frontend values, server manifest, program ID, mint owner, Config PDA, admin, treasury, fee, and token program agree.

## Realtime behavior

Chat is durable and refreshes on a three-second polling interval. It does not currently use WebSockets or a shared pub/sub system. This is sufficient for the hackathon MVP but adds up to several seconds of visible delivery latency.
