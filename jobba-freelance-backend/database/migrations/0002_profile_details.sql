begin;

alter table public.profiles
  add column if not exists headline text,
  add column if not exists location text;

do $$
begin
  if not exists (select 1 from pg_constraint where conname = 'profiles_headline_length') then
    alter table public.profiles add constraint profiles_headline_length
      check (headline is null or char_length(headline) between 1 and 120);
  end if;
  if not exists (select 1 from pg_constraint where conname = 'profiles_location_length') then
    alter table public.profiles add constraint profiles_location_length
      check (location is null or char_length(location) between 1 and 120);
  end if;
end $$;

commit;
