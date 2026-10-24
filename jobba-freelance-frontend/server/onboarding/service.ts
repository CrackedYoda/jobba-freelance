import "server-only";

import { missingRequirements, type SaveOnboardingInput } from "@/features/profiles/schemas";
import { getDatabase } from "@/server/db/client";
import { AppError } from "@/server/db/errors";
import { checkSignupBalance } from "@/server/onboarding/balance";
import type { Actor } from "@/server/repositories/profiles";

type Row = Record<string, unknown>;

export async function readOnboarding(profileId: string) {
  const db = getDatabase();
  const { rows } = await db.query(
    `select p.*, (select updated_at from profile_avatars a where a.profile_id = p.id) as avatar_updated_at,
            coalesce((select array_agg(r.role::text order by r.role) from profile_roles r where r.profile_id = p.id and r.role <> 'admin'), '{}') as roles,
            exists (select 1 from profile_roles r where r.profile_id = p.id and r.role = 'admin') as is_admin,
            coalesce((select array_agg(ps.skill_id::text) from profile_skills ps where ps.profile_id = p.id), '{}') as skill_ids,
            coalesce((select json_agg(json_build_object('title', e.title, 'organization', e.organization, 'startYear', e.start_year, 'endYear', e.end_year, 'description', e.description) order by e.sort_order)
                        from profile_experience e where e.profile_id = p.id), '[]') as experience,
            coalesce((select json_agg(json_build_object('title', l.title, 'url', l.url) order by l.sort_order) from portfolio_links l where l.profile_id = p.id), '[]') as portfolio
       from profiles p where p.id = $1`,
    [profileId],
  );
  const p = rows[0] as Row | undefined;
  if (!p) throw new AppError("PROFILE_NOT_FOUND", "Profile not found.", 404);
  const draft = {
    roles: p.roles as string[], displayName: String(p.display_name), headline: (p.headline as string | null) ?? null, location: (p.location as string | null) ?? null,
    bio: (p.bio as string | null) ?? null, hourlyRateAtomic: p.hourly_rate_atomic == null ? null : String(p.hourly_rate_atomic),
    skillIds: p.skill_ids as string[], experience: p.experience as unknown[], yearsExperience: (p.years_experience as number | null) ?? null,
    hasAvatar: Boolean(p.avatar_updated_at), companyName: (p.company_name as string | null) ?? null, industry: (p.industry as string | null) ?? null,
    companySize: (p.company_size as string | null) ?? null, hiringNeeds: (p.hiring_needs as string | null) ?? null,
  };
  const skills = (await db.query("select id, name::text, category::text from skills order by name")).rows;
  return {
    profileId: String(p.id), walletAddress: String(p.wallet_public_key), status: String(p.onboarding_status), step: Number(p.onboarding_step),
    primaryRole: (p.primary_role as string | null) ?? null, isAdmin: Boolean(p.is_admin), walletVerifiedAt: p.wallet_verified_at, avatarUrl: (p.avatar_url as string | null) ?? null,
    availability: String(p.availability), portfolio: p.portfolio as { title: string; url: string }[], companyWebsite: (p.company_website as string | null) ?? null,
    balance: p.balance_checked_at ? { lamports: String(p.balance_lamports), usdCents: String(p.balance_usd_cents), checkedAt: p.balance_checked_at } : null,
    ...draft, missing: missingRequirements(draft), availableSkills: skills,
  };
}

export async function saveOnboarding(actor: Actor, input: SaveOnboardingInput) {
  const db = getDatabase();
  const conn = await db.connect();
  try {
    await conn.query("begin");
    const columns: [string, unknown][] = [];
    const set = (column: string, value: unknown) => { if (value !== undefined) columns.push([column, value]); };
    set("onboarding_step", input.step);
    set("primary_role", input.primaryRole);
    set("display_name", input.displayName);
    set("headline", input.headline === "" ? null : input.headline);
    set("location", input.location === "" ? null : input.location);
    set("bio", input.bio === "" ? null : input.bio);
    set("hourly_rate_atomic", input.hourlyRateAtomic);
    set("availability", input.availability);
    set("years_experience", input.yearsExperience);
    set("company_name", input.companyName === "" ? null : input.companyName);
    set("company_website", input.companyWebsite);
    set("company_size", input.companySize);
    set("industry", input.industry === "" ? null : input.industry);
    set("hiring_needs", input.hiringNeeds === "" ? null : input.hiringNeeds);
    if (columns.length) {
      await conn.query(`update profiles set ${columns.map(([column], index) => `${column} = $${index + 2}`).join(", ")} where id = $1`, [actor.profileId, ...columns.map(([, value]) => value)]);
    }
    if (input.roles) {
      await conn.query("delete from profile_roles where profile_id = $1 and role <> 'admin'", [actor.profileId]);
      for (const role of input.roles) await conn.query("insert into profile_roles (profile_id, role) values ($1, $2)", [actor.profileId, role]);
    }
    if (input.skillIds) {
      if (input.skillIds.length) {
        const found = await conn.query("select id from skills where id = any($1::uuid[])", [input.skillIds]);
        if (found.rowCount !== input.skillIds.length) throw new AppError("INVALID_SKILL", "One or more selected skills do not exist.", 400);
      }
      await conn.query("delete from profile_skills where profile_id = $1", [actor.profileId]);
      if (input.skillIds.length) await conn.query("insert into profile_skills (profile_id, skill_id) select $1, unnest($2::uuid[])", [actor.profileId, input.skillIds]);
    }
    if (input.experience) {
      await conn.query("delete from profile_experience where profile_id = $1", [actor.profileId]);
      for (const [index, item] of input.experience.entries()) {
        await conn.query("insert into profile_experience (profile_id, title, organization, start_year, end_year, description, sort_order) values ($1,$2,$3,$4,$5,$6,$7)",
          [actor.profileId, item.title, item.organization, item.startYear, item.endYear, item.description || null, index]);
      }
    }
    if (input.portfolio) {
      await conn.query("delete from portfolio_links where profile_id = $1", [actor.profileId]);
      for (const [index, link] of input.portfolio.entries()) await conn.query("insert into portfolio_links (profile_id, title, url, sort_order) values ($1,$2,$3,$4)", [actor.profileId, link.title, link.url, index]);
    }
    await conn.query("commit");
  } catch (error) {
    await conn.query("rollback");
    throw error;
  } finally {
    conn.release();
  }
  return readOnboarding(actor.profileId);
}

async function recordBalance(actor: Actor) {
  const balance = await checkSignupBalance(actor.walletAddress);
  await getDatabase().query("update profiles set balance_lamports = $2, balance_usd_cents = $3, balance_checked_at = now() where id = $1", [actor.profileId, balance.lamports, balance.usdCents]);
  return balance;
}

export async function verifyBalance(actor: Actor) {
  return recordBalance(actor);
}

export async function completeOnboarding(actor: Actor, balanceVerifier: typeof recordBalance = recordBalance) {
  const current = await readOnboarding(actor.profileId);
  if (current.missing.length) throw new AppError("ONBOARDING_INCOMPLETE", "Some required details are missing.", 422, { missing: current.missing });
  const balance = await balanceVerifier(actor);
  if (!balance.sufficient) {
    const min = (Number(balance.minimumUsdCents) / 100).toFixed(2);
    throw new AppError("INSUFFICIENT_BALANCE", `Your wallet needs at least $${min} worth of SOL to finish signing up. It currently holds ${balance.sol} SOL (about $${(Number(balance.usdCents) / 100).toFixed(2)}).`, 402, { balance });
  }
  await getDatabase().query("update profiles set onboarding_status = 'completed', onboarding_completed_at = coalesce(onboarding_completed_at, now()), onboarding_step = 10 where id = $1", [actor.profileId]);
  return { ...(await readOnboarding(actor.profileId)), balance };
}

const signatures: [string, number[]][] = [["image/png", [0x89, 0x50, 0x4e, 0x47]], ["image/jpeg", [0xff, 0xd8, 0xff]], ["image/webp", [0x52, 0x49, 0x46, 0x46]]];

export async function saveAvatar(actor: Actor, file: File) {
  if (file.size === 0 || file.size > 2 * 1024 * 1024) throw new AppError("INVALID_IMAGE", "Upload an image up to 2 MB.", 400);
  const bytes = Buffer.from(await file.arrayBuffer());
  const type = signatures.find(([, magic]) => magic.every((byte, index) => bytes[index] === byte))?.[0];
  if (!type || (type === "image/webp" && bytes.subarray(8, 12).toString("ascii") !== "WEBP")) throw new AppError("INVALID_IMAGE", "Use a PNG, JPEG or WebP image.", 400);
  const db = getDatabase();
  await db.query(
    `insert into profile_avatars (profile_id, content_type, bytes, updated_at) values ($1,$2,$3,now())
     on conflict (profile_id) do update set content_type = excluded.content_type, bytes = excluded.bytes, updated_at = now()`,
    [actor.profileId, type, bytes],
  );
  const url = `/api/v1/profiles/${actor.profileId}/avatar?v=${Date.now()}`;
  await db.query("update profiles set avatar_url = $2 where id = $1", [actor.profileId, url]);
  return { avatarUrl: url };
}

export async function readAvatar(profileId: string) {
  const { rows } = await getDatabase().query<{ content_type: string; bytes: Buffer }>("select content_type, bytes from profile_avatars where profile_id = $1", [profileId]);
  return rows[0];
}
