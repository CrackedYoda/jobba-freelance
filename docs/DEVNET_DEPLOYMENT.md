# Devnet Deployment Runbook

## Fixed local values

- Program: `CwNiWw9TSRDnszphzCHsShxC64VhgbawWxCLnZ9gLhxy`
- Deployer: `BMPZFSJ4oiSgmMvbXKfyYyBCvtvR4vx637ZvooRio7c8`
- Cluster: Solana devnet
- Token program: `TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA`
- Program keypair: `.tooling/keys/freelance-escrow-devnet.json` (Git-ignored, mode 0600)
- Deployer keypair: `.tooling/keys/devnet-deployer.json` (Git-ignored, mode 0600)
- Program artifact: `jobba-freelance-backend/target/deploy/freelance_escrow.so`

Never commit, print, paste, or transmit either keypair file. Back them up in an approved secret store before relying on this program identity.

## Build

```bash
cd /Users/yoda/Desktop/jobba-freelance
export AVM_HOME="$PWD/.tooling/avm"
export RUSTUP_HOME="$PWD/.tooling/rustup"
export CARGO_HOME="$PWD/.tooling/cargo"
export PATH="/Users/yoda/.local/share/solana/install/active_release/bin:$CARGO_HOME/bin:$PATH"
cd jobba-freelance-backend
$AVM_HOME/bin/anchor-0.30.1 build --no-idl
```

The application uses its checked low-level instruction encoder, so an Anchor IDL is not required at runtime. Standard IDL generation invokes a nightly Rust download and can be run later with the same command without `--no-idl`.

## Fund and deploy

The public faucet rejected both 5 SOL and 1 SOL requests on 24 October 2026. Fund the deployer with devnet SOL from an approved faucet or another devnet wallet, then confirm the balance:

```bash
solana balance BMPZFSJ4oiSgmMvbXKfyYyBCvtvR4vx637ZvooRio7c8 --url devnet
```

Deploy the already-built artifact:

```bash
solana program deploy \
  --url devnet \
  --keypair /Users/yoda/Desktop/jobba-freelance/.tooling/keys/devnet-deployer.json \
  --program-id /Users/yoda/Desktop/jobba-freelance/.tooling/keys/freelance-escrow-devnet.json \
  /Users/yoda/Desktop/jobba-freelance/jobba-freelance-backend/target/deploy/freelance_escrow.so
```

Verify that the reported program address exactly matches the fixed program ID above.

## Mint and Config PDA

Create or select a clearly labeled six-decimal devnet Mock USDC mint. Record its mint authority and freeze policy for the demo. Use the same mint, token program, administrator, treasury, and fee basis points when invoking `initialize_config` and when producing `SOLANA_DEPLOYMENT_MANIFEST_JSON`.

The web server validates the complete Config account and mint owner/decimals before constructing any transaction. A mismatch fails closed.

## Application configuration

Set matching public and server values:

```text
NEXT_PUBLIC_SOLANA_CLUSTER=devnet
NEXT_PUBLIC_SOLANA_PROGRAM_ID=CwNiWw9TSRDnszphzCHsShxC64VhgbawWxCLnZ9gLhxy
SOLANA_PROGRAM_ID=CwNiWw9TSRDnszphzCHsShxC64VhgbawWxCLnZ9gLhxy
NEXT_PUBLIC_USDC_MINT=<six-decimal-devnet-mint>
SOLANA_USDC_MINT=<same-mint>
```

Also provide the Reown project ID, RPC URL, PostgreSQL URL, application origin/domain, reconciliation secret, and compact deployment manifest. Run `npm run env:check:server` before deployment.

## Acceptance matrix

Use separate client, freelancer, and administrator wallets and execute:

1. Wallet sign-in and profile onboarding.
2. Project posting, proposal submission, selection, and escrow funding.
3. Freelancer acceptance, chat, submission, and client release.
4. A separate funded contract cancelled before acceptance.
5. A separate contract disputed by a participant and resolved once by refund and once by award.
6. Rejected signatures, wrong wallet/network, refresh/retry, duplicate idempotency keys, and finalized database reconciliation.
