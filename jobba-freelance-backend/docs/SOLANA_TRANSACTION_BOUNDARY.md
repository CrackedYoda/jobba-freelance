# Solana transaction boundary

The web server constructs escrow transactions from database contract rows and the on-chain Config account. It never accepts an amount, mint, terms hash, participant, vault, or recipient token account from the browser. It never loads a private key.

Set `SOLANA_DEPLOYMENT_MANIFEST_JSON` to the compact JSON equivalent of `solana-deployment-manifest.example.json`. The server validates every public key, cluster, fee, mint owner and decimals, and the complete Config account before returning an unsigned transaction. Missing or inconsistent deployment configuration fails closed.

The browser displays the canonical summary and asks the connected wallet to sign and submit. The confirmation endpoint independently fetches the confirmed transaction, verifies the fee payer, program, ordered accounts, discriminator, instruction arguments, and success status, then records the signature. Database contract projection remains the responsibility of a finalized reconciliation worker; `server/solana/reconcile-action.ts` provides its account-reading primitive.

The example contains placeholders and is not a deployable manifest. Keep deployment and admin keypairs outside the repository.
