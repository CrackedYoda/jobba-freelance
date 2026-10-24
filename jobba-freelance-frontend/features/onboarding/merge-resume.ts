export type ResumeDraft = {
  displayName: string | null; headline: string | null; location: string | null; bio: string | null; yearsExperience: number | null;
  skills: string[];
  experience: { title: string; organization: string; startYear: number; endYear: number | null; description: string | null }[];
  portfolio: { title: string; url: string }[];
};

export type ProfileFields = {
  displayName: string; headline: string | null; location: string | null; bio: string | null; yearsExperience: number | null;
  skillIds: string[]; availableSkills: { id: string; name: string }[];
  experience: { title: string; organization: string; startYear: number; endYear: number | null; description?: string | null }[];
  portfolio: { title: string; url: string }[];
};

const empty = (value: string | null) => !value || !value.trim();
const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? "" : "s"}`;

/**
 * Merges a résumé draft into the profile without overwriting anything the freelancer typed:
 * empty fields are filled, and new skills, roles and links are added. Returns the patch and a summary.
 */
export function mergeResume(profile: ProfileFields, draft: ResumeDraft) {
  const patch: Partial<Omit<ProfileFields, "availableSkills">> = {};
  const changed: string[] = [];
  if (draft.displayName && /^Wallet\s/.test(profile.displayName)) { patch.displayName = draft.displayName; changed.push("Name"); }
  if (draft.headline && empty(profile.headline)) { patch.headline = draft.headline; changed.push("Headline"); }
  if (draft.location && empty(profile.location)) { patch.location = draft.location; changed.push("Location"); }
  if (draft.bio && empty(profile.bio)) { patch.bio = draft.bio; changed.push("About you"); }
  if (draft.yearsExperience !== null && profile.yearsExperience === null) { patch.yearsExperience = draft.yearsExperience; changed.push("Years of experience"); }

  const byName = new Map(profile.availableSkills.map((skill) => [skill.name.toLowerCase(), skill.id]));
  const newSkills = [...new Set(draft.skills.map((name) => byName.get(name.toLowerCase())).filter((id): id is string => Boolean(id) && !profile.skillIds.includes(id as string)))];
  if (newSkills.length) { patch.skillIds = [...profile.skillIds, ...newSkills].slice(0, 20); changed.push(plural(newSkills.length, "skill")); }

  const keep = profile.experience.filter((item) => item.title.trim() && item.organization.trim());
  const seen = new Set(keep.map((item) => `${item.title}|${item.organization}`.toLowerCase()));
  const newRoles = draft.experience.filter((item) => !seen.has(`${item.title}|${item.organization}`.toLowerCase()));
  if (newRoles.length) { patch.experience = [...keep, ...newRoles].slice(0, 10); changed.push(`${plural(newRoles.length, "role")} of experience`); }

  const links = profile.portfolio.filter((link) => link.title.trim() && link.url.trim());
  const urls = new Set(links.map((link) => link.url));
  const newLinks = draft.portfolio.filter((link) => !urls.has(link.url));
  if (newLinks.length) { patch.portfolio = [...links, ...newLinks].slice(0, 12); changed.push(plural(newLinks.length, "portfolio link")); }

  return { patch, changed };
}
