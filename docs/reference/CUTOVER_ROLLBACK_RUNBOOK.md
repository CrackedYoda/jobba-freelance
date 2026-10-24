# Jobba PostgreSQL cutover and rollback runbook

Updated: 22 October 2026
Status: rehearsal required; no production cutover authorized or performed

## Safety boundary

The importer is dry-run by default and cannot write without both `--apply` and `JOBBA_IMPORT_ACK=I_HAVE_A_BACKUP`. It never deletes source MongoDB data. PostgreSQL imports use legacy identifiers and idempotent conflict handling; invalid JSON and orphan references are written to quarantine.

Before Go accepts production writes, rollback means routing traffic back to Node while MongoDB remains authoritative. After Go accepts writes, routing back without reverse reconciliation can lose or fork data. That point is the irreversible operational boundary and requires a separate approval during the cutover window.

## Required inputs and owners

- Named cutover operator and independent verifier.
- Confirmed inventory of REST, Socket.IO, webhook, scheduled-job, and mobile clients.
- Encrypted MongoDB snapshot with restore test and recorded checksum.
- PostgreSQL backup destination with restore test.
- Final export directory containing JSONL files documented in `Jobba-server-go/import/README.md`.
- Approved maintenance window or a separately designed change-capture system. Informal dual writes are prohibited.
- Paystack webhook routing plan and reconciliation owner.

## Rehearsal

1. Start disposable PostgreSQL and validate every migration in both directions:

   ```sh
   cd Jobba-server-go
   make db-smoke
   ```

   This creates and removes only the `jobba_migration_smoke` database inside the Compose PostgreSQL service. It applies all `*.up.sql`, verifies critical tables, applies every `*.down.sql` in reverse order, and requires zero remaining public tables.

2. Export a protected MongoDB snapshot to JSONL. Preserve password hashes, public IDs, timestamps, missing/null distinctions, monetary values, references, and provider references. Do not recalculate historic fees.

3. Run validation without database credentials:

   ```sh
   go run ./cmd/import-data --source ./import/export --out ./import/rehearsal
   ```

4. Review `report.json` and every line of `quarantine.jsonl`. The expected quarantine count is zero unless each exception has a written disposition and owner. Save the reported source hash.

5. Apply to the disposable migrated database:

   ```sh
   export DATABASE_URL='postgres://jobba:jobba@localhost:5432/jobba_migration_smoke?sslmode=disable'
   export JOBBA_IMPORT_ACK=I_HAVE_A_BACKUP
   go run ./cmd/import-data --source ./import/export --out ./import/rehearsal --apply
   ```

6. Interrupt the import after several batches, restart with the unchanged export, and confirm checkpoint skips. Change one source byte and confirm resume is rejected.

7. Reconcile per-collection source/valid/applied counts, legacy-ID uniqueness, orphan counts, user identities, bcrypt samples, job/application assignments, payment totals by currency/status, payout references, message ordering, unread state, and dispute states.

8. Run Go race tests, vet, build, database integration tests, critical mobile journeys, signed webhook replay tests, and Paystack sandbox verification. Record commands, versions, timestamps, and evidence.

## Production cutover

1. Announce the window and stop new Node writes, background jobs, Socket.IO mutations, and Paystack webhook consumption. Confirm write quiescence from metrics and logs.
2. Take the final MongoDB backup and record its checksum and last authoritative timestamp.
3. Export final JSONL, run dry-run validation, and compare the source hash with the reviewed artifact. Any new quarantine blocks cutover.
4. Back up PostgreSQL, apply migrations, and run the importer with the explicit acknowledgement.
5. Run reconciliation queries and sampled API comparisons. Counts alone are insufficient; verify references, money, statuses, timestamps, and representative response contracts.
6. Route Paystack webhooks to one owner only. Replay retained events by stable event/reference keys and reconcile ambiguous payouts before enabling release operations.
7. Start Go in protected mode, run readiness and authenticated smoke tests, then shift traffic gradually using pre-agreed error, latency, and reconciliation thresholds.
8. Mark the exact timestamp when Go becomes the authoritative writer. From this point, rollback requires reverse reconciliation or forward repair.

## Rollback

### Before Go writes

1. Stop Go and webhook delivery to Go.
2. Route traffic and webhooks back to Node.
3. Verify MongoDB remained unchanged and authoritative.
4. Preserve PostgreSQL, import reports, quarantine, checkpoints, and logs for diagnosis. Do not run down migrations on the only evidence copy.

### After Go writes

1. Stop all writers and webhook consumers.
2. Capture PostgreSQL and MongoDB backups plus provider event state.
3. Reconcile Go-only users, jobs, messages, payments, payout intents, notifications, and disputes by stable legacy/provider identifiers.
4. Choose forward repair in Go or an reviewed reverse-import into MongoDB. Do not simply route Node to stale MongoDB.
5. Resume one authoritative writer only after financial and relationship invariants pass.

## Acceptance record

Record the export hash, migration checksums, source/import/quarantine counts, database backups, test output, provider reconciliation totals, traffic thresholds, authoritative-write timestamp, operator, verifier, and the final go/no-go decision. Production retirement of Node requires a completed observation window and a tested recovery copy.
