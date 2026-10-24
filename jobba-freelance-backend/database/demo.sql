-- Demo data for local development. Wallet keys are placeholders, not real keypairs,
-- so these profiles cannot sign in; they exist so the marketplace has content.
-- The mint is the Mock USDC placeholder: replace it once the devnet mint is frozen.
begin;

insert into public.profiles (id, wallet_public_key, display_name, headline, location, bio, hourly_rate_atomic, availability) values
  ('20000000-0000-4000-8000-000000000001', '7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU', 'Northstar Labs', null, 'Lagos, NG', null, null, 'available'),
  ('20000000-0000-4000-8000-000000000002', 'GkP2TrWjZ8qNf5bYdXvH3mLcA9sUeR4tKxWpQy7r3rTa', 'Mesa Protocol', null, 'Austin, US', null, null, 'available'),
  ('20000000-0000-4000-8000-000000000003', '5Qe9LvZc2HnJm8dTfXw4UaRkYb7NpGs3VtCxEiMoD6hB', 'Plaintext Systems', null, 'Berlin, DE', null, null, 'available'),
  ('20000000-0000-4000-8000-000000000004', '9Rt4WxN6bKd2PqHsLmZv8YcAeUf3GjTn7XoCiVyB5hMk', 'Amara Okafor', 'Product designer for complex financial tools', 'Lagos, NG', 'I turn operational software into calm, legible products. My recent work spans payments, analytics, and internal tools for distributed teams.', 65000000, 'available'),
  ('20000000-0000-4000-8000-000000000005', '3Bn8XqJd5TrLw2KsPzHv9YcMeUf6GaNt4CxEiVoD7hRk', 'Kofi Mensah', 'Solana and TypeScript engineer', 'Accra, GH', 'Anchor programs, indexers and the TypeScript services around them. I ship small, tested increments.', 90000000, 'limited'),
  ('20000000-0000-4000-8000-000000000006', '6Dm3YpHc8WzKt5RsQnLv2XbAeUf9GjTa4NxCiVoB7hEk', 'Ines Duarte', 'Technical writer for developer platforms', 'Lisbon, PT', 'Documentation, launch copy and positioning for technical buyers.', 55000000, 'available')
on conflict (id) do nothing;

insert into public.profile_roles (profile_id, role) values
  ('20000000-0000-4000-8000-000000000001', 'client'), ('20000000-0000-4000-8000-000000000002', 'client'),
  ('20000000-0000-4000-8000-000000000003', 'client'),
  ('20000000-0000-4000-8000-000000000004', 'freelancer'), ('20000000-0000-4000-8000-000000000005', 'freelancer'),
  ('20000000-0000-4000-8000-000000000006', 'freelancer')
on conflict do nothing;

insert into public.profile_skills (profile_id, skill_id)
select p.id, s.id from (values
  ('20000000-0000-4000-8000-000000000004'::uuid, 'product-design'), ('20000000-0000-4000-8000-000000000004', 'figma'),
  ('20000000-0000-4000-8000-000000000004', 'design-systems'), ('20000000-0000-4000-8000-000000000004', 'user-research'),
  ('20000000-0000-4000-8000-000000000005', 'solana'), ('20000000-0000-4000-8000-000000000005', 'typescript'),
  ('20000000-0000-4000-8000-000000000005', 'smart-contracts'), ('20000000-0000-4000-8000-000000000005', 'postgresql'),
  ('20000000-0000-4000-8000-000000000006', 'technical-writing'), ('20000000-0000-4000-8000-000000000006', 'copywriting')
) as p(id, slug) join public.skills s on s.slug = p.slug
on conflict do nothing;

insert into public.portfolio_links (profile_id, title, url, sort_order) values
  ('20000000-0000-4000-8000-000000000004', 'Payment operations case study', 'https://example.com/case-study', 0),
  ('20000000-0000-4000-8000-000000000004', 'Design system overview', 'https://example.com/system', 1),
  ('20000000-0000-4000-8000-000000000005', 'Escrow program write-up', 'https://example.com/escrow', 0)
on conflict do nothing;

insert into public.projects (id, client_id, title, description, category, budget_atomic, delivery_days, currency_mint, status, published_at, created_at) values
  ('30000000-0000-4000-8000-000000000001', '20000000-0000-4000-8000-000000000001', 'Redesign a Solana commerce dashboard',
   'We need a product designer who can audit the current dashboard, establish a compact design system, and deliver responsive Figma screens for onboarding, payments, refunds, and reporting. Experience with financial products is useful; deep blockchain knowledge is optional.',
   'Design & Creative', 2400000000, 21, '4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU', 'open', now() - interval '2 hours', now() - interval '2 hours'),
  ('30000000-0000-4000-8000-000000000002', '20000000-0000-4000-8000-000000000002', 'Build an Anchor event indexer',
   'Build a TypeScript service that reads finalized Anchor events, stores normalized records, and resumes safely after RPC interruptions. Include structured logs and a concise operations guide.',
   'Development', 4800000000, 30, '4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU', 'open', now() - interval '5 hours', now() - interval '5 hours'),
  ('30000000-0000-4000-8000-000000000003', '20000000-0000-4000-8000-000000000003', 'Launch copy for a developer platform',
   'Interview two founders, distill the product narrative, and write homepage, pricing, and launch announcement copy. The voice should be direct, technical, and credible.',
   'Writing', 1350000000, 14, '4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU', 'open', now() - interval '1 day', now() - interval '1 day'),
  ('30000000-0000-4000-8000-000000000004', '20000000-0000-4000-8000-000000000001', '60-second escrow explainer video',
   'Script, storyboard and animate a 60-second explainer with voiceover. Brand assets are supplied; two rounds of revisions included.',
   'Video & animation', 1800000000, 18, '4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU', 'open', now() - interval '4 days', now() - interval '4 days')
on conflict (id) do nothing;

insert into public.project_skills (project_id, skill_id)
select p.id, s.id from (values
  ('30000000-0000-4000-8000-000000000001'::uuid, 'product-design'), ('30000000-0000-4000-8000-000000000001', 'figma'), ('30000000-0000-4000-8000-000000000001', 'design-systems'),
  ('30000000-0000-4000-8000-000000000002', 'solana'), ('30000000-0000-4000-8000-000000000002', 'typescript'), ('30000000-0000-4000-8000-000000000002', 'postgresql'),
  ('30000000-0000-4000-8000-000000000003', 'copywriting'), ('30000000-0000-4000-8000-000000000003', 'technical-writing'),
  ('30000000-0000-4000-8000-000000000004', 'motion-design')
) as p(id, slug) join public.skills s on s.slug = p.slug
on conflict do nothing;

insert into public.proposals (id, project_id, freelancer_id, cover_letter, amount_atomic, estimated_days) values
  ('40000000-0000-4000-8000-000000000001', '30000000-0000-4000-8000-000000000001', '20000000-0000-4000-8000-000000000004',
   'I would begin with a two-day workflow audit, then prototype the highest-risk payment states before expanding the system across the dashboard.', 2350000000, 18),
  ('40000000-0000-4000-8000-000000000002', '30000000-0000-4000-8000-000000000002', '20000000-0000-4000-8000-000000000005',
   'I have shipped two Anchor event indexers. I would start with the backfill path and add reconciliation tests before any live ingestion.', 4600000000, 28)
on conflict (id) do nothing;

commit;
