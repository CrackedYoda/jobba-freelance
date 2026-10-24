import type { ResumeDraft } from "@/features/onboarding/merge-resume";

/**
 * Free, rule-based résumé reader. It needs no external service: it looks for the usual résumé structure
 * (name and title at the top, Summary / Experience / Skills sections, date ranges and links) in the plain text
 * of the document. It is deliberately conservative: when unsure it leaves a field empty for the freelancer.
 */

const SECTION_PATTERNS: [Section, RegExp][] = [
  ["summary", /^(professional\s+)?(summary|profile|about(\s+me)?|objective|overview|bio)$/i],
  ["experience", /^(work\s+|professional\s+|relevant\s+)?(experience|employment(\s+history)?|work\s+history|career(\s+history)?)$/i],
  ["skills", /^(technical\s+|core\s+|key\s+)?(skills|competencies|expertise|tools(\s+&\s+technologies)?|technologies|tech\s+stack)$/i],
  ["other", /^(education|projects|selected\s+projects|certifications?|licenses|languages|awards|honou?rs|publications|references|interests|hobbies|volunteer(ing)?|contact)$/i],
];
type Section = "summary" | "experience" | "skills" | "other" | "top";

const MONTH = "(?:jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:t(?:ember)?)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\\.?";
const DATE = `(?:${MONTH}\\s+|\\d{1,2}\\/)?((?:19|20)\\d{2})`;
const RANGE = new RegExp(`${DATE}\\s*(?:–|—|-|to|until)\\s*(?:${DATE}|(present|current|now|today|date))`, "i");
const ROLE_WORDS = /\b(designer|developer|engineer|writer|manager|marketer|analyst|consultant|specialist|artist|editor|producer|architect|strategist|researcher|scientist|lead|founder|director|freelancer|copywriter|illustrator|animator|administrator|accountant|translator|photographer|videographer|programmer|coordinator|officer|assistant|intern)\b/i;
const URL_RE = /\b(?:https?:\/\/|www\.)[^\s,;|()<>]+|\b(?:linkedin\.com|github\.com|gitlab\.com|behance\.net|dribbble\.com|medium\.com|notion\.site|vercel\.app|substack\.com)\/[^\s,;|()<>]+/gi;
const HAS_URL = new RegExp(URL_RE.source, "i");
const EMAIL_RE = /[\w.+-]+@[\w-]+\.[\w.-]+/;
const PHONE_RE = /(\+?\d[\d\s().-]{7,}\d)/;

/** Extra words that indicate a skill from Jobba's list, keyed by the lowercased skill name. */
const SYNONYMS: Record<string, string[]> = {
  "frontend development": ["frontend", "front-end", "front end", "react", "next.js", "nextjs", "vue", "angular", "svelte", "html", "css", "tailwind"],
  "backend development": ["backend", "back-end", "back end", "node.js", "nodejs", "express", "django", "flask", "fastapi", "rails", "laravel", "rest api", "graphql"],
  "smart contracts": ["smart contract", "solidity", "anchor", "rust programs", "evm"],
  "solana": ["solana", "spl token", "anchor"],
  "postgresql": ["postgres", "postgresql"],
  "sql": ["sql", "mysql", "sqlite", "bigquery"],
  "typescript": ["typescript", "ts"],
  "python": ["python", "pandas", "numpy"],
  "ui/ux design": ["ui/ux", "ux/ui", "ux design", "ui design", "user experience", "user interface", "wireframes", "prototyping"],
  "product design": ["product design", "product designer"],
  "design systems": ["design system", "design systems", "component library"],
  "figma": ["figma"],
  "user research": ["user research", "usability testing", "user interviews", "ux research"],
  "motion design": ["motion design", "motion graphics", "after effects", "animation", "lottie"],
  "copywriting": ["copywriting", "copywriter", "copy writing", "content writing"],
  "technical writing": ["technical writing", "technical writer", "documentation", "api docs"],
  "community": ["community management", "community manager", "community building", "discord moderation"],
};

const clean = (line: string) => line.replace(/^[\s•●▪◦‣∙*–—-]+/, "").replace(/\s+/g, " ").trim();
const isHeading = (line: string): Section | undefined => {
  const bare = line.replace(/[:：]\s*$/, "").trim();
  if (bare.length > 40) return undefined;
  return SECTION_PATTERNS.find(([, pattern]) => pattern.test(bare))?.[0];
};

function splitSections(lines: string[]) {
  const sections: Record<Section, string[]> = { top: [], summary: [], experience: [], skills: [], other: [] };
  let current: Section = "top";
  for (const line of lines) {
    const heading = isHeading(line);
    if (heading) { current = heading; continue; }
    sections[current].push(line);
  }
  return sections;
}

function findName(top: string[]): string | null {
  for (const line of top.slice(0, 4)) {
    if (EMAIL_RE.test(line) || PHONE_RE.test(line) || HAS_URL.test(line) || /\d/.test(line)) continue;
    const words = line.split(" ");
    if (words.length >= 2 && words.length <= 4 && words.every((word) => /^[A-ZÀ-Ý][A-Za-zÀ-ÿ'’.-]*$/.test(word)) && !ROLE_WORDS.test(line)) {
      return line.split(" ").map((word) => (word === word.toUpperCase() && word.length > 2 ? word[0] + word.slice(1).toLowerCase() : word)).join(" ");
    }
  }
  return null;
}

function findHeadline(top: string[], name: string | null): string | null {
  for (const raw of top.slice(0, 6)) {
    for (const part of raw.split(/\s+[|•·]\s+/)) {
      const line = part.trim();
      if (!line || line === name || EMAIL_RE.test(line) || PHONE_RE.test(line) || HAS_URL.test(line)) continue;
      if (ROLE_WORDS.test(line) && line.length <= 120 && line.split(" ").length <= 14) return line;
    }
  }
  return null;
}

function findLocation(top: string[]): string | null {
  for (const raw of top.slice(0, 8)) {
    const labelled = raw.match(/\b(?:location|based in|address)\s*[:\-]?\s*([A-Za-zÀ-ÿ .'-]+,\s*[A-Za-zÀ-ÿ .'-]+)/i);
    if (labelled) return labelled[1].trim();
    for (const part of raw.split(/\s*[|•·]\s*/)) {
      const line = part.trim();
      if (/^[A-ZÀ-Ý][A-Za-zÀ-ÿ .'-]{1,40},\s*[A-ZÀ-Ý][A-Za-zÀ-ÿ .'-]{1,40}$/.test(line) && !ROLE_WORDS.test(line)) return line;
    }
  }
  return null;
}

function findLinks(text: string) {
  const seen = new Set<string>();
  const links: { title: string; url: string }[] = [];
  for (const match of text.matchAll(URL_RE)) {
    let url = match[0].replace(/[.)\]]+$/, "");
    if (EMAIL_RE.test(url)) continue;
    if (!/^https?:\/\//i.test(url)) url = `https://${url}`;
    let host: string;
    try { host = new URL(url).hostname.replace(/^www\./, ""); } catch { continue; }
    if (seen.has(url.toLowerCase())) continue;
    seen.add(url.toLowerCase());
    const title = host.includes("github") ? "GitHub" : host.includes("linkedin") ? "LinkedIn" : host.includes("behance") ? "Behance" : host.includes("dribbble") ? "Dribbble" : host.includes("gitlab") ? "GitLab" : host.includes("medium") ? "Medium" : "Portfolio";
    links.push({ title, url });
  }
  return links.slice(0, 12);
}

function findSkills(text: string, skillNames: string[]): string[] {
  const haystack = ` ${text.toLowerCase().replace(/\s+/g, " ")} `;
  const has = (term: string) => new RegExp(`(^|[^a-z0-9])${term.replace(/[.*+?^${}()|[\]\\/]/g, "\\$&")}([^a-z0-9]|$)`).test(haystack);
  return skillNames.filter((name) => [name.toLowerCase(), ...(SYNONYMS[name.toLowerCase()] ?? [])].some(has));
}

function splitRole(text: string): { title: string; organization: string } | null {
  const line = text.replace(/[,|·•–—-]\s*$/, "").trim();
  const parts = line.split(/\s+(?:at|@)\s+|\s+[|·•–—]\s+|\s+-\s+|,\s+/).map((part) => part.trim()).filter(Boolean);
  if (parts.length >= 2) {
    const [a, b] = parts;
    // "Company | Title" when the first part doesn't look like a role but the second does.
    return !ROLE_WORDS.test(a) && ROLE_WORDS.test(b) ? { title: b, organization: a } : { title: a, organization: b };
  }
  return null;
}

function findExperience(lines: string[], now: number) {
  const roles: ResumeDraft["experience"] = [];
  for (let i = 0; i < lines.length; i += 1) {
    const match = lines[i].match(RANGE);
    if (!match) continue;
    const startYear = Number(match[1]);
    const endYear = match[3] ? null : Number(match[2]);
    if (startYear < 1950 || startYear > now || (endYear !== null && (endYear < startYear || endYear > now))) continue;
    const rest = clean(lines[i].replace(match[0], "").replace(/[()]/g, ""));
    let role = rest ? splitRole(rest) : null;
    if (!role && rest && i > 0) role = splitRole(`${lines[i - 1]}, ${rest}`) ?? { title: lines[i - 1], organization: rest };
    if (!role && i > 0) role = splitRole(lines[i - 1]) ?? (i > 1 ? { title: lines[i - 2], organization: lines[i - 1] } : null);
    if (!role || role.title.length > 120 || role.organization.length > 120 || isHeading(role.title) || RANGE.test(role.title)) continue;
    const detail: string[] = [];
    for (let j = i + 1; j < lines.length && detail.join(" ").length < 300; j += 1) {
      if (RANGE.test(lines[j]) || (j + 1 < lines.length && RANGE.test(lines[j + 1]) && !lines[j].endsWith("."))) break;
      detail.push(clean(lines[j]));
    }
    const description = detail.join(" ").replace(/\s+/g, " ").trim();
    roles.push({ title: role.title, organization: role.organization, startYear, endYear, description: !description ? null : description.length <= 300 ? description : `${description.slice(0, 300).replace(/\s+\S*$/, "")}…` });
    if (roles.length === 6) break;
  }
  return roles;
}

function findBio(summary: string[]): string | null {
  const text = summary.map(clean).join(" ").replace(/\s+/g, " ").trim();
  if (text.length < 40) return null;
  if (text.length <= 600) return text;
  const cut = text.slice(0, 600);
  return cut.slice(0, Math.max(cut.lastIndexOf(". ") + 1, 300)).trim();
}

export function parseResumeText(raw: string, skillNames: string[], now = new Date().getUTCFullYear()): ResumeDraft {
  const lines = raw.replace(/\r/g, "").split("\n").map((line) => line.replace(/\s+/g, " ").trim()).filter(Boolean);
  const sections = splitSections(lines);
  const top = sections.top.length ? sections.top : lines.slice(0, 8);
  const displayName = findName(top);
  const experience = findExperience(sections.experience.length ? sections.experience : lines, now);
  const stated = raw.match(/(\d{1,2})\+?\s+years?(?:\s+of)?\s+(?:professional\s+|industry\s+|hands-on\s+)?experience/i);
  const earliest = experience.length ? Math.min(...experience.map((role) => role.startYear)) : null;
  const yearsExperience = stated ? Math.min(60, Number(stated[1])) : earliest !== null ? Math.max(0, Math.min(60, now - earliest)) : null;
  return {
    displayName,
    headline: findHeadline(top, displayName),
    location: findLocation(top),
    bio: findBio(sections.summary),
    yearsExperience,
    skills: findSkills(raw, skillNames),
    experience,
    portfolio: findLinks(raw),
  };
}
