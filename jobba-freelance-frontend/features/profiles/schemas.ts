import { z } from "zod";

const atomicAmount = z.string().regex(/^\d+$/, "Rate must be an atomic-unit integer string.").refine((value) => BigInt(value) <= 9_223_372_036_854_775_807n, "Rate exceeds PostgreSQL bigint.");
const role = z.enum(["client", "freelancer"]);
const safeUrl = z.string().trim().url().max(2048).refine((value) => {
  try {
    const protocol = new URL(value).protocol;
    return protocol === "https:" || protocol === "http:";
  } catch {
    return false;
  }
}, "Links must use HTTP or HTTPS.");
const text = (max: number) => z.string().trim().max(max);
const currentYear = new Date().getUTCFullYear();

export const portfolioLinkSchema = z.object({ title: z.string().trim().min(1).max(120), url: safeUrl }).strict();
export const experienceSchema = z.object({
  title: z.string().trim().min(1).max(120),
  organization: z.string().trim().min(1).max(120),
  startYear: z.number().int().min(1950).max(currentYear),
  endYear: z.number().int().min(1950).max(currentYear).nullable(),
  description: text(2000).nullable().optional(),
}).strict().refine((value) => value.endYear === null || value.endYear >= value.startYear, { message: "End year must not be before start year.", path: ["endYear"] });

/** Partial, saved on every step so sign-up can be resumed. Completeness is checked separately. */
export const saveOnboardingSchema = z.object({
  step: z.number().int().min(0).max(10).optional(),
  primaryRole: role.optional(),
  roles: z.array(role).min(1).max(2).transform((values) => [...new Set(values)]).optional(),
  displayName: z.string().trim().min(1).max(80).optional(),
  headline: text(120).nullable().optional(),
  location: text(120).nullable().optional(),
  bio: text(2000).nullable().optional(),
  hourlyRateAtomic: atomicAmount.nullable().optional(),
  availability: z.enum(["available", "limited", "unavailable"]).optional(),
  yearsExperience: z.number().int().min(0).max(60).nullable().optional(),
  skillIds: z.array(z.string().uuid()).max(20).transform((values) => [...new Set(values)]).optional(),
  experience: z.array(experienceSchema).max(10).optional(),
  portfolio: z.array(portfolioLinkSchema).max(12).optional(),
  companyName: text(120).nullable().optional(),
  companyWebsite: z.preprocess((value) => value === "" ? null : value, safeUrl.nullable().optional()),
  companySize: z.enum(["solo", "2-10", "11-50", "51-200", "201+"]).nullable().optional(),
  industry: text(80).nullable().optional(),
  hiringNeeds: text(2000).nullable().optional(),
}).strict();

export type SaveOnboardingInput = z.infer<typeof saveOnboardingSchema>;

export type OnboardingRequirement = { field: string; message: string };

type Draft = {
  roles: string[]; displayName: string; headline: string | null; location: string | null; bio: string | null;
  hourlyRateAtomic: string | null; skillIds: string[]; experience: unknown[]; yearsExperience: number | null; hasAvatar: boolean;
  companyName: string | null; industry: string | null; companySize: string | null; hiringNeeds: string | null;
};

/** What is still missing before a profile may complete sign-up. Shared by the wizard and the server. */
export function missingRequirements(draft: Draft): OnboardingRequirement[] {
  const missing: OnboardingRequirement[] = [];
  const need = (ok: boolean, field: string, message: string) => { if (!ok) missing.push({ field, message }); };
  need(draft.roles.length > 0, "roles", "Choose whether you are hiring, freelancing, or both.");
  need(Boolean(draft.displayName.trim()) && !/^Wallet\s/.test(draft.displayName), "displayName", "Add the name people will see.");
  need(draft.hasAvatar, "avatar", "Upload a profile photo or logo.");
  need(Boolean(draft.location?.trim()), "location", "Add your location.");
  if (draft.roles.includes("freelancer")) {
    need(Boolean(draft.headline?.trim()), "headline", "Add a professional headline.");
    need((draft.bio?.trim().length ?? 0) >= 40, "bio", "Write a bio of at least 40 characters.");
    need(Boolean(draft.hourlyRateAtomic && BigInt(draft.hourlyRateAtomic) > 0n), "hourlyRateAtomic", "Set an hourly guide rate.");
    need(draft.skillIds.length > 0, "skillIds", "Pick at least one skill.");
    need(draft.yearsExperience !== null, "yearsExperience", "Tell clients how many years of experience you have.");
    need(draft.experience.length > 0, "experience", "Add at least one role from your relevant experience.");
  }
  if (draft.roles.includes("client")) {
    need(Boolean(draft.companyName?.trim()), "companyName", "Add your company or team name (or your own name if you hire as an individual).");
    need(Boolean(draft.industry?.trim()), "industry", "Choose your industry.");
    need(Boolean(draft.companySize), "companySize", "Choose your team size.");
    need((draft.hiringNeeds?.trim().length ?? 0) >= 20, "hiringNeeds", "Describe what you plan to hire for (at least 20 characters).");
  }
  return missing;
}
