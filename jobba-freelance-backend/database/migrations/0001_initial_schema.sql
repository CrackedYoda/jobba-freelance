begin;

create extension if not exists pgcrypto;
create extension if not exists citext;

create type public.app_role as enum ('client', 'freelancer', 'admin');
create type public.availability_status as enum ('available', 'limited', 'unavailable');
create type public.project_status as enum ('draft', 'open', 'contracted', 'completed', 'cancelled', 'disputed');
create type public.proposal_status as enum ('pending', 'accepted', 'rejected', 'withdrawn');
create type public.contract_status as enum ('funding_pending', 'funded', 'active', 'submitted', 'released', 'cancelled', 'disputed', 'resolved');
create type public.submission_status as enum ('submitted', 'accepted', 'superseded');
create type public.dispute_status as enum ('open', 'under_review', 'awaiting_chain', 'resolved');
create type public.dispute_outcome as enum ('refund_client', 'award_freelancer');
create type public.chain_action_kind as enum ('fund', 'accept', 'submit', 'release', 'cancel', 'open_dispute', 'refund', 'award');
create type public.chain_action_status as enum ('created', 'submitted', 'confirmed', 'finalized', 'expired', 'failed');
create type public.solana_cluster as enum ('devnet', 'testnet', 'mainnet-beta', 'localnet');
create type public.solana_commitment as enum ('processed', 'confirmed', 'finalized');

create function public.set_updated_at()
returns trigger language plpgsql set search_path = pg_catalog, public as $$
begin
  new.updated_at = clock_timestamp();
  return new;
end;
$$;

create function public.reject_row_mutation()
returns trigger language plpgsql set search_path = pg_catalog, public as $$
begin
  raise exception '% is append-only', tg_table_name using errcode = '55000';
end;
$$;

create function public.request_profile_id()
returns uuid language plpgsql stable security definer set search_path = pg_catalog, public as $$
declare
  claims jsonb;
  raw_id text;
begin
  raw_id := nullif(current_setting('app.profile_id', true), '');
  if raw_id is not null then return raw_id::uuid; end if;
  claims := nullif(current_setting('request.jwt.claims', true), '')::jsonb;
  raw_id := coalesce(nullif(claims ->> 'profile_id', ''), nullif(claims ->> 'sub', ''));
  return raw_id::uuid;
exception when others then
  return null;
end;
$$;

revoke all on function public.request_profile_id() from public;

create table public.profiles (
  id uuid primary key default gen_random_uuid(),
  wallet_public_key text not null unique check (wallet_public_key ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  display_name text not null check (char_length(display_name) between 1 and 80),
  bio text check (bio is null or char_length(bio) <= 2000),
  hourly_rate_atomic bigint check (hourly_rate_atomic is null or hourly_rate_atomic >= 0),
  availability public.availability_status not null default 'available',
  avatar_url text check (avatar_url is null or char_length(avatar_url) <= 2048),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table public.profile_roles (
  profile_id uuid not null references public.profiles(id) on delete cascade,
  role public.app_role not null,
  created_at timestamptz not null default now(),
  primary key (profile_id, role)
);

create table public.skills (
  id uuid primary key default gen_random_uuid(),
  slug citext not null unique check (slug::text ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
  name citext not null unique check (char_length(name::text) between 1 and 80),
  category citext not null check (char_length(category::text) between 1 and 80),
  created_at timestamptz not null default now()
);

create table public.profile_skills (
  profile_id uuid not null references public.profiles(id) on delete cascade,
  skill_id uuid not null references public.skills(id) on delete restrict,
  created_at timestamptz not null default now(),
  primary key (profile_id, skill_id)
);

create table public.portfolio_links (
  id uuid primary key default gen_random_uuid(),
  profile_id uuid not null references public.profiles(id) on delete cascade,
  title text not null check (char_length(title) between 1 and 120),
  url text not null check (char_length(url) between 8 and 2048),
  sort_order integer not null default 0 check (sort_order >= 0),
  created_at timestamptz not null default now()
);

create table public.projects (
  id uuid primary key default gen_random_uuid(),
  client_id uuid not null references public.profiles(id) on delete restrict,
  title text not null check (char_length(title) between 3 and 160),
  description text not null check (char_length(description) between 20 and 10000),
  category citext not null check (char_length(category::text) between 1 and 80),
  budget_atomic bigint not null check (budget_atomic > 0),
  currency_mint text not null check (currency_mint ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  status public.project_status not null default 'draft',
  version bigint not null default 1 check (version > 0),
  published_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check ((status = 'draft' and published_at is null) or status <> 'draft')
);

create table public.project_skills (
  project_id uuid not null references public.projects(id) on delete cascade,
  skill_id uuid not null references public.skills(id) on delete restrict,
  primary key (project_id, skill_id)
);

create table public.proposals (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  freelancer_id uuid not null references public.profiles(id) on delete restrict,
  cover_letter text not null check (char_length(cover_letter) between 20 and 5000),
  amount_atomic bigint not null check (amount_atomic > 0),
  estimated_days integer not null check (estimated_days between 1 and 3650),
  status public.proposal_status not null default 'pending',
  version bigint not null default 1 check (version > 0),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (project_id, freelancer_id)
);

create table public.contracts (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null unique references public.projects(id) on delete restrict,
  proposal_id uuid not null unique references public.proposals(id) on delete restrict,
  client_id uuid not null references public.profiles(id) on delete restrict,
  freelancer_id uuid not null references public.profiles(id) on delete restrict,
  amount_atomic bigint not null check (amount_atomic > 0),
  fee_atomic bigint not null default 0 check (fee_atomic >= 0 and fee_atomic <= amount_atomic),
  currency_mint text not null check (currency_mint ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  terms_hash bytea not null check (octet_length(terms_hash) = 32),
  status public.contract_status not null default 'funding_pending',
  version bigint not null default 1 check (version > 0),
  escrow_pda text unique check (escrow_pda is null or escrow_pda ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  funded_at timestamptz,
  accepted_at timestamptz,
  submitted_at timestamptz,
  completed_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (client_id <> freelancer_id)
);

create table public.submissions (
  id uuid primary key default gen_random_uuid(),
  contract_id uuid not null references public.contracts(id) on delete restrict,
  freelancer_id uuid not null references public.profiles(id) on delete restrict,
  description text not null check (char_length(description) between 1 and 10000),
  deliverable_urls jsonb not null default '[]'::jsonb check (jsonb_typeof(deliverable_urls) = 'array'),
  content_hash bytea check (content_hash is null or octet_length(content_hash) = 32),
  status public.submission_status not null default 'submitted',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index one_current_submission_per_contract on public.submissions(contract_id)
  where status = 'submitted';

create table public.conversations (
  id uuid primary key default gen_random_uuid(),
  contract_id uuid not null unique references public.contracts(id) on delete restrict,
  created_at timestamptz not null default now()
);

create table public.conversation_participants (
  conversation_id uuid not null references public.conversations(id) on delete cascade,
  profile_id uuid not null references public.profiles(id) on delete cascade,
  joined_at timestamptz not null default now(),
  primary key (conversation_id, profile_id)
);

create table public.messages (
  id uuid primary key default gen_random_uuid(),
  conversation_id uuid not null references public.conversations(id) on delete cascade,
  sender_id uuid not null references public.profiles(id) on delete restrict,
  client_message_id uuid not null,
  body text not null check (char_length(body) between 1 and 5000),
  created_at timestamptz not null default now(),
  unique (sender_id, client_message_id),
  foreign key (conversation_id, sender_id)
    references public.conversation_participants(conversation_id, profile_id) on delete restrict
);

create table public.conversation_read_cursors (
  conversation_id uuid not null,
  profile_id uuid not null,
  last_read_message_id uuid,
  last_read_at timestamptz not null default now(),
  primary key (conversation_id, profile_id),
  foreign key (conversation_id, profile_id) references public.conversation_participants(conversation_id, profile_id) on delete cascade,
  foreign key (last_read_message_id) references public.messages(id) on delete set null
);

create table public.disputes (
  id uuid primary key default gen_random_uuid(),
  contract_id uuid not null unique references public.contracts(id) on delete restrict,
  opened_by uuid not null references public.profiles(id) on delete restrict,
  assigned_admin_id uuid references public.profiles(id) on delete restrict,
  category text not null check (char_length(category) between 1 and 80),
  statement text not null check (char_length(statement) between 20 and 10000),
  terms_hash bytea not null check (octet_length(terms_hash) = 32),
  submission_snapshot jsonb,
  escrow_pda text not null check (escrow_pda ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  status public.dispute_status not null default 'open',
  version bigint not null default 1 check (version > 0),
  resolved_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table public.dispute_evidence (
  id uuid primary key default gen_random_uuid(),
  dispute_id uuid not null references public.disputes(id) on delete restrict,
  submitted_by uuid not null references public.profiles(id) on delete restrict,
  description text not null check (char_length(description) between 1 and 5000),
  urls jsonb not null default '[]'::jsonb check (jsonb_typeof(urls) = 'array'),
  created_at timestamptz not null default now()
);

create table public.dispute_decisions (
  id uuid primary key default gen_random_uuid(),
  dispute_id uuid not null unique references public.disputes(id) on delete restrict,
  admin_id uuid not null references public.profiles(id) on delete restrict,
  outcome public.dispute_outcome not null,
  reason text not null check (char_length(reason) between 20 and 10000),
  amount_atomic bigint not null check (amount_atomic > 0),
  chain_action_id uuid,
  created_at timestamptz not null default now()
);

create table public.admin_audit_log (
  id bigint generated always as identity primary key,
  admin_id uuid not null references public.profiles(id) on delete restrict,
  action text not null check (char_length(action) between 1 and 120),
  entity_type text not null check (char_length(entity_type) between 1 and 80),
  entity_id uuid,
  dispute_id uuid references public.disputes(id) on delete restrict,
  metadata jsonb not null default '{}'::jsonb check (jsonb_typeof(metadata) = 'object'),
  occurred_at timestamptz not null default now()
);

create table public.auth_nonces (
  id uuid primary key default gen_random_uuid(),
  wallet_public_key text not null check (wallet_public_key ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  nonce_hash bytea not null unique check (octet_length(nonce_hash) = 32),
  message_hash bytea not null check (octet_length(message_hash) = 32),
  message_text text not null check (char_length(message_text) between 1 and 4096),
  domain text not null check (char_length(domain) between 1 and 253),
  uri text not null check (char_length(uri) between 1 and 2048),
  chain_id text not null check (char_length(chain_id) between 1 and 80),
  request_ip_hash bytea check (request_ip_hash is null or octet_length(request_ip_hash) = 32),
  issued_at timestamptz not null,
  expires_at timestamptz not null,
  consumed_at timestamptz,
  created_at timestamptz not null default now(),
  check (expires_at > issued_at),
  check (consumed_at is null or consumed_at >= issued_at)
);

create table public.auth_sessions (
  id uuid primary key default gen_random_uuid(),
  profile_id uuid not null references public.profiles(id) on delete cascade,
  token_hash bytea not null unique check (octet_length(token_hash) = 32),
  wallet_public_key text not null check (wallet_public_key ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  user_agent_hash bytea check (user_agent_hash is null or octet_length(user_agent_hash) = 32),
  ip_hash bytea check (ip_hash is null or octet_length(ip_hash) = 32),
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz not null default now(),
  last_seen_at timestamptz not null default now(),
  check (expires_at > created_at),
  check (revoked_at is null or revoked_at >= created_at)
);

create table public.chain_actions (
  id uuid primary key default gen_random_uuid(),
  action_key text not null unique check (char_length(action_key) between 8 and 200),
  idempotency_key text not null check (char_length(idempotency_key) between 8 and 200),
  kind public.chain_action_kind not null,
  status public.chain_action_status not null default 'created',
  contract_id uuid not null references public.contracts(id) on delete restrict,
  actor_id uuid not null references public.profiles(id) on delete restrict,
  cluster public.solana_cluster not null,
  program_id text not null check (program_id ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  mint text not null check (mint ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  token_program_id text not null check (token_program_id ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  escrow_pda text not null check (escrow_pda ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  amount_atomic bigint not null check (amount_atomic > 0),
  terms_hash bytea not null check (octet_length(terms_hash) = 32),
  expected_accounts jsonb not null check (jsonb_typeof(expected_accounts) = 'object'),
  expires_at timestamptz not null,
  failure_code text,
  failure_detail text,
  version bigint not null default 1 check (version > 0),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (actor_id, idempotency_key)
);

create table public.chain_signatures (
  id uuid primary key default gen_random_uuid(),
  chain_action_id uuid not null references public.chain_actions(id) on delete restrict,
  signature text not null unique check (char_length(signature) between 64 and 128),
  slot bigint check (slot is null or slot >= 0),
  block_time timestamptz,
  commitment public.solana_commitment,
  submitted_at timestamptz not null default now(),
  confirmed_at timestamptz,
  finalized_at timestamptz,
  raw_transaction jsonb,
  check (confirmed_at is null or confirmed_at >= submitted_at),
  check (finalized_at is null or confirmed_at is not null)
);

create table public.chain_events (
  id bigint generated always as identity primary key,
  signature_id uuid not null references public.chain_signatures(id) on delete restrict,
  chain_action_id uuid not null references public.chain_actions(id) on delete restrict,
  program_id text not null check (program_id ~ '^[1-9A-HJ-NP-Za-km-z]{32,44}$'),
  event_name text not null check (char_length(event_name) between 1 and 120),
  event_index integer not null check (event_index >= 0),
  slot bigint not null check (slot >= 0),
  block_time timestamptz,
  payload jsonb not null check (jsonb_typeof(payload) = 'object'),
  observed_at timestamptz not null default now(),
  unique (signature_id, event_index)
);

create table public.reconciliation_cursors (
  cluster public.solana_cluster primary key,
  last_slot bigint not null default 0 check (last_slot >= 0),
  last_signature text,
  updated_at timestamptz not null default now()
);

alter table public.dispute_decisions add constraint dispute_decision_chain_action_fk
  foreign key (chain_action_id) references public.chain_actions(id) on delete restrict;

create function public.is_conversation_participant(target_conversation_id uuid, target_profile_id uuid)
returns boolean language sql stable security definer
set search_path = pg_catalog, public set row_security = off as $$
  select exists (
    select 1 from public.conversation_participants cp
    where cp.conversation_id = target_conversation_id and cp.profile_id = target_profile_id
  );
$$;

create function public.is_contract_participant(target_contract_id uuid, target_profile_id uuid)
returns boolean language sql stable security definer
set search_path = pg_catalog, public set row_security = off as $$
  select exists (
    select 1 from public.contracts c
    where c.id = target_contract_id
      and target_profile_id in (c.client_id, c.freelancer_id)
  );
$$;

create function public.can_read_dispute(target_dispute_id uuid, target_profile_id uuid)
returns boolean language sql stable security definer
set search_path = pg_catalog, public set row_security = off as $$
  select exists (
    select 1
    from public.disputes d
    join public.contracts c on c.id = d.contract_id
    where d.id = target_dispute_id
      and (target_profile_id in (c.client_id, c.freelancer_id) or d.assigned_admin_id = target_profile_id)
  );
$$;

revoke all on function public.is_conversation_participant(uuid, uuid) from public;
revoke all on function public.is_contract_participant(uuid, uuid) from public;
revoke all on function public.can_read_dispute(uuid, uuid) from public;

create index profiles_display_name_idx on public.profiles(display_name);
create index profile_roles_role_idx on public.profile_roles(role, profile_id);
create index skills_category_idx on public.skills(category, name);
create index profile_skills_skill_idx on public.profile_skills(skill_id, profile_id);
create index projects_browse_idx on public.projects(status, created_at desc, id desc);
create index projects_client_idx on public.projects(client_id, created_at desc);
create index projects_category_budget_idx on public.projects(category, budget_atomic) where status = 'open';
create index project_skills_skill_idx on public.project_skills(skill_id, project_id);
create index proposals_project_status_idx on public.proposals(project_id, status, created_at desc);
create index proposals_freelancer_idx on public.proposals(freelancer_id, created_at desc);
create index contracts_client_idx on public.contracts(client_id, updated_at desc);
create index contracts_freelancer_idx on public.contracts(freelancer_id, updated_at desc);
create index contracts_status_idx on public.contracts(status, updated_at desc);
create index submissions_contract_idx on public.submissions(contract_id, created_at desc);
create index conversation_participants_profile_idx on public.conversation_participants(profile_id, conversation_id);
create index messages_cursor_idx on public.messages(conversation_id, created_at desc, id desc);
create index disputes_queue_idx on public.disputes(status, created_at, id);
create index disputes_admin_idx on public.disputes(assigned_admin_id, status, created_at);
create index dispute_evidence_idx on public.dispute_evidence(dispute_id, created_at, id);
create index admin_audit_entity_idx on public.admin_audit_log(entity_type, entity_id, occurred_at desc);
create index admin_audit_dispute_idx on public.admin_audit_log(dispute_id, occurred_at desc);
create index auth_nonces_lookup_idx on public.auth_nonces(wallet_public_key, expires_at desc) where consumed_at is null;
create index auth_sessions_profile_idx on public.auth_sessions(profile_id, expires_at desc) where revoked_at is null;
create index auth_sessions_expiry_idx on public.auth_sessions(expires_at) where revoked_at is null;
create index chain_actions_contract_idx on public.chain_actions(contract_id, created_at desc);
create index chain_actions_reconcile_idx on public.chain_actions(status, expires_at) where status in ('created', 'submitted', 'confirmed');
create index chain_signatures_action_idx on public.chain_signatures(chain_action_id, submitted_at desc);
create index chain_events_action_idx on public.chain_events(chain_action_id, slot, event_index);
create index chain_events_slot_idx on public.chain_events(slot, id);

create trigger profiles_updated_at before update on public.profiles for each row execute function public.set_updated_at();
create trigger projects_updated_at before update on public.projects for each row execute function public.set_updated_at();
create trigger proposals_updated_at before update on public.proposals for each row execute function public.set_updated_at();
create trigger contracts_updated_at before update on public.contracts for each row execute function public.set_updated_at();
create trigger submissions_updated_at before update on public.submissions for each row execute function public.set_updated_at();
create trigger disputes_updated_at before update on public.disputes for each row execute function public.set_updated_at();
create trigger chain_actions_updated_at before update on public.chain_actions for each row execute function public.set_updated_at();
create trigger reconciliation_cursors_updated_at before update on public.reconciliation_cursors for each row execute function public.set_updated_at();

create trigger admin_audit_immutable before update or delete on public.admin_audit_log
  for each row execute function public.reject_row_mutation();
create trigger chain_events_immutable before update or delete on public.chain_events
  for each row execute function public.reject_row_mutation();
create trigger dispute_decisions_immutable before update or delete on public.dispute_decisions
  for each row execute function public.reject_row_mutation();

alter table public.profiles enable row level security;
alter table public.profile_roles enable row level security;
alter table public.skills enable row level security;
alter table public.profile_skills enable row level security;
alter table public.portfolio_links enable row level security;
alter table public.projects enable row level security;
alter table public.project_skills enable row level security;
alter table public.proposals enable row level security;
alter table public.contracts enable row level security;
alter table public.submissions enable row level security;
alter table public.conversations enable row level security;
alter table public.conversation_participants enable row level security;
alter table public.messages enable row level security;
alter table public.conversation_read_cursors enable row level security;
alter table public.disputes enable row level security;
alter table public.dispute_evidence enable row level security;
alter table public.dispute_decisions enable row level security;
alter table public.admin_audit_log enable row level security;
alter table public.auth_nonces enable row level security;
alter table public.auth_sessions enable row level security;
alter table public.chain_actions enable row level security;
alter table public.chain_signatures enable row level security;
alter table public.chain_events enable row level security;
alter table public.reconciliation_cursors enable row level security;

create policy profiles_public_read on public.profiles for select using (true);
create policy skills_public_read on public.skills for select using (true);
create policy profile_skills_public_read on public.profile_skills for select using (true);
create policy portfolio_public_read on public.portfolio_links for select using (true);
create policy open_projects_public_read on public.projects for select using (status = 'open' or client_id = public.request_profile_id());
create policy project_skills_public_read on public.project_skills for select using (
  exists (select 1 from public.projects p where p.id = project_id and (p.status = 'open' or p.client_id = public.request_profile_id()))
);
create policy own_roles_read on public.profile_roles for select using (profile_id = public.request_profile_id());
create policy proposal_participant_read on public.proposals for select using (
  freelancer_id = public.request_profile_id() or exists (
    select 1 from public.projects p where p.id = project_id and p.client_id = public.request_profile_id()
  )
);
create policy contract_participant_read on public.contracts for select using (
  public.is_contract_participant(id, public.request_profile_id()) or exists (
    select 1 from public.disputes d where d.contract_id = id and d.assigned_admin_id = public.request_profile_id()
  )
);
create policy submission_participant_read on public.submissions for select using (
  public.is_contract_participant(contract_id, public.request_profile_id()) or exists (
    select 1 from public.disputes d where d.contract_id = contract_id and d.assigned_admin_id = public.request_profile_id()
  )
);
create policy conversation_participant_read on public.conversations for select using (
  public.is_conversation_participant(id, public.request_profile_id())
);
create policy participant_list_read on public.conversation_participants for select using (
  public.is_conversation_participant(conversation_id, public.request_profile_id())
);
create policy participant_message_read on public.messages for select using (
  public.is_conversation_participant(conversation_id, public.request_profile_id())
);
create policy own_read_cursor_read on public.conversation_read_cursors for select using (profile_id = public.request_profile_id());
create policy dispute_participant_read on public.disputes for select using (
  public.can_read_dispute(id, public.request_profile_id())
);
create policy dispute_evidence_participant_read on public.dispute_evidence for select using (
  public.can_read_dispute(dispute_id, public.request_profile_id())
);
create policy dispute_decision_participant_read on public.dispute_decisions for select using (
  public.can_read_dispute(dispute_id, public.request_profile_id())
);
create policy chain_action_participant_read on public.chain_actions for select using (
  actor_id = public.request_profile_id() or public.is_contract_participant(contract_id, public.request_profile_id())
);
create policy chain_signature_participant_read on public.chain_signatures for select using (
  exists (select 1 from public.chain_actions ca where ca.id = chain_action_id
    and public.is_contract_participant(ca.contract_id, public.request_profile_id()))
);
create policy chain_event_participant_read on public.chain_events for select using (
  exists (select 1 from public.chain_actions ca where ca.id = chain_action_id
    and public.is_contract_participant(ca.contract_id, public.request_profile_id()))
);


comment on column public.projects.budget_atomic is 'USDC atomic units (six decimals); never a floating-point amount.';
comment on column public.contracts.amount_atomic is 'Gross USDC amount in atomic units.';
comment on table public.admin_audit_log is 'Append-only audit trail, including sensitive dispute evidence reads and admin mutations.';
comment on function public.request_profile_id() is 'Custom SIWS RLS identity: UUID from app.profile_id or request JWT profile_id/sub claim.';

commit;
