begin;

-- Resumable sign-up: the profile row exists from the first verified wallet signature,
-- and onboarding progress is saved step by step until it is completed.
alter table public.profiles
  add column if not exists onboarding_status text not null default 'in_progress' check (onboarding_status in ('in_progress', 'completed')),
  add column if not exists onboarding_step smallint not null default 0 check (onboarding_step between 0 and 10),
  add column if not exists onboarding_completed_at timestamptz,
  add column if not exists primary_role public.app_role,
  add column if not exists wallet_verified_at timestamptz,
  add column if not exists balance_checked_at timestamptz,
  add column if not exists balance_lamports bigint check (balance_lamports is null or balance_lamports >= 0),
  add column if not exists balance_usd_cents bigint check (balance_usd_cents is null or balance_usd_cents >= 0),
  add column if not exists years_experience smallint check (years_experience is null or years_experience between 0 and 60),
  add column if not exists company_name text check (company_name is null or char_length(company_name) between 1 and 120),
  add column if not exists company_website text check (company_website is null or char_length(company_website) <= 2048),
  add column if not exists company_size text check (company_size is null or company_size in ('solo', '2-10', '11-50', '51-200', '201+')),
  add column if not exists industry text check (industry is null or char_length(industry) between 1 and 80),
  add column if not exists hiring_needs text check (hiring_needs is null or char_length(hiring_needs) <= 2000);

-- Profiles that already had roles before this migration were fully set up.
update public.profiles p set onboarding_status = 'completed', onboarding_completed_at = coalesce(onboarding_completed_at, now()), wallet_verified_at = coalesce(wallet_verified_at, created_at)
 where exists (select 1 from public.profile_roles r where r.profile_id = p.id) and onboarding_status <> 'completed';

create table if not exists public.profile_experience (
  id uuid primary key default gen_random_uuid(),
  profile_id uuid not null references public.profiles(id) on delete cascade,
  title text not null check (char_length(title) between 1 and 120),
  organization text not null check (char_length(organization) between 1 and 120),
  start_year smallint not null check (start_year between 1950 and 2100),
  end_year smallint check (end_year is null or end_year between 1950 and 2100),
  description text check (description is null or char_length(description) <= 2000),
  sort_order smallint not null default 0,
  created_at timestamptz not null default now(),
  check (end_year is null or end_year >= start_year)
);
create index if not exists profile_experience_profile_idx on public.profile_experience(profile_id, sort_order);

-- Uploaded profile photos and company logos, served by the API.
create table if not exists public.profile_avatars (
  profile_id uuid primary key references public.profiles(id) on delete cascade,
  content_type text not null check (content_type in ('image/png', 'image/jpeg', 'image/webp')),
  bytes bytea not null check (octet_length(bytes) between 1 and 2097152),
  updated_at timestamptz not null default now()
);

alter table public.profile_experience enable row level security;
alter table public.profile_avatars enable row level security;

commit;
