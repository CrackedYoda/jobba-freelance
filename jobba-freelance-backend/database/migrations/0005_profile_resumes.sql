begin;

-- A freelancer's uploaded résumé. Private to its owner: it is used to pre-fill the profile
-- and is never served publicly.
create table if not exists public.profile_resumes (
  profile_id uuid primary key references public.profiles(id) on delete cascade,
  file_name text not null check (char_length(file_name) between 1 and 200),
  content_type text not null check (content_type in ('application/pdf', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document')),
  bytes bytea not null check (octet_length(bytes) between 1 and 5242880),
  extracted_at timestamptz,
  updated_at timestamptz not null default now()
);

alter table public.profile_resumes enable row level security;

commit;
