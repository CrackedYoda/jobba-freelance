begin;

alter table public.projects
  add column if not exists delivery_days integer not null default 30 check (delivery_days between 1 and 365);

alter table public.proposals
  add column if not exists portfolio_url text check (portfolio_url is null or char_length(portfolio_url) <= 2048);

commit;
