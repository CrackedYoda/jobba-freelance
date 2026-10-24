# Database

The SQL files in `migrations/` are authoritative and are applied in lexical order. The initial migration targets plain PostgreSQL 15+ (no Supabase roles or extensions). It creates the full marketplace, custom SIWS authentication, chat, dispute, audit, and Solana reconciliation schema.

## Apply locally

Local Docker Postgres (port 5434; drops and recreates the `public` schema, then applies migrations and seed):

```sh
docker compose up -d db   # from jobba-freelance-backend
./database/reset.sh
```

Against an explicitly selected PostgreSQL database:

```sh
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 \
  -f database/migrations/0001_initial_schema.sql \
  -f database/seed.sql
```

Never point this command at production without reviewing the target URL. The seed contains only idempotent reference rows. A demo reset should generate fresh contract UUIDs because escrow PDAs cannot be reused after terminal settlement.

## Authentication and RLS

Application writes go through server route handlers over `DATABASE_URL` and explicit authorization checks. The browser never connects to the database. RLS is defense in depth for non-owner roles; mutations stay behind the SIWS session boundary.

For direct PostgreSQL server connections, set the authenticated profile inside the same transaction:

```sql
begin;
select set_config('app.profile_id', '<profile uuid>', true);
select * from public.messages where conversation_id = '<conversation uuid>';
commit;
```

`request_profile_id()` also accepts a UUID in `request.jwt.claims` (`profile_id` or `sub`) for any future JWT-based reader. Never expose `DATABASE_URL` to the browser.

## Invariants

- All authoritative USDC values are `bigint` atomic units. USDC has six decimals.
- Hashes are raw 32-byte `bytea`, not display strings.
- Proposal uniqueness is enforced per project and freelancer.
- One contract exists per project and accepted proposal; one conversation exists per contract.
- Admin audit rows, dispute decisions, and observed chain events are immutable.
- `action_key` and actor-scoped `idempotency_key` prevent duplicate chain actions.
- State-transition validation still belongs in transactional server functions using row locks or `version` checks; enum columns prevent invalid state names but do not replace the transition matrix.

The API should consume nonces atomically with a statement equivalent to:

```sql
update public.auth_nonces
set consumed_at = now()
where id = $1 and consumed_at is null and expires_at > now()
returning *;
```
