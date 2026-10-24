-- Deterministic reference data only. Demo profiles/projects should be created by the
-- reset script with fresh UUIDs so terminal escrow PDAs never collide.
insert into public.skills (id, slug, name, category) values
  ('10000000-0000-4000-8000-000000000001', 'frontend-development', 'Frontend Development', 'Development'),
  ('10000000-0000-4000-8000-000000000002', 'backend-development', 'Backend Development', 'Development'),
  ('10000000-0000-4000-8000-000000000003', 'solana', 'Solana', 'Blockchain'),
  ('10000000-0000-4000-8000-000000000004', 'smart-contracts', 'Smart Contracts', 'Blockchain'),
  ('10000000-0000-4000-8000-000000000005', 'ui-ux-design', 'UI/UX Design', 'Design'),
  ('10000000-0000-4000-8000-000000000006', 'technical-writing', 'Technical Writing', 'Writing')
on conflict (id) do update set
  slug = excluded.slug,
  name = excluded.name,
  category = excluded.category;

insert into public.reconciliation_cursors (cluster, last_slot)
values ('devnet', 0)
on conflict (cluster) do nothing;

insert into public.skills (slug, name, category) values
  ('product-design', 'Product design', 'Design'),
  ('figma', 'Figma', 'Design'),
  ('design-systems', 'Design systems', 'Design'),
  ('user-research', 'User research', 'Design'),
  ('typescript', 'TypeScript', 'Development'),
  ('postgresql', 'PostgreSQL', 'Development'),
  ('python', 'Python', 'Data & AI'),
  ('sql', 'SQL', 'Data & AI'),
  ('copywriting', 'Copywriting', 'Writing'),
  ('community', 'Community', 'Marketing'),
  ('motion-design', 'Motion design', 'Video & animation')
on conflict (slug) do nothing;
