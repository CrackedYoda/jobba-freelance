import { describe, expect, it } from "vitest";

import { parseResumeText } from "../../features/onboarding/parse-resume";

const SKILLS = ["Backend Development", "Community", "Copywriting", "Design systems", "Figma", "Frontend Development", "Motion design", "PostgreSQL", "Product design", "Python", "SQL", "Smart Contracts", "Solana", "Technical Writing", "TypeScript", "UI/UX Design", "User research"];

const designer = `AMARA OKAFOR
Senior Product Designer
Lagos, Nigeria | amara@example.com | +234 801 234 5678
dribbble.com/amara · https://github.com/amara-o

Summary
Product designer with 7 years of experience turning complex financial workflows into calm, usable products. I lead research, prototyping and design systems work for payments teams.

Experience
Lead Product Designer — Paystack
Jan 2021 – Present
• Redesigned the merchant dashboard used by 200k businesses.
• Built the company design system in Figma.
Product Designer at Andela
2018 - 2020
• Ran user research and usability testing for internal tools.

Skills
Figma, Prototyping, Design Systems, User Research, UX writing

Education
BSc Computer Science, University of Lagos, 2014 – 2018`;

const engineer = `Kofi Mensah
Solana Engineer | Accra, Ghana
kofi@example.com

PROFESSIONAL EXPERIENCE
Mesa Protocol | Smart Contract Engineer | 2022 – Current
Wrote Anchor programs and TypeScript services backed by PostgreSQL.
Freelance, Backend Developer, 2019 – 2022
Node.js and Django APIs for small businesses.

TECHNICAL SKILLS
Rust, Anchor, Solidity, TypeScript, Node.js, Postgres, SQL`;

describe("free résumé parser", () => {
  it("reads the name, title, location, links and summary from a designer CV", () => {
    const draft = parseResumeText(designer, SKILLS, 2026);
    expect(draft.displayName).toBe("Amara Okafor");
    expect(draft.headline).toBe("Senior Product Designer");
    expect(draft.location).toBe("Lagos, Nigeria");
    expect(draft.bio).toMatch(/^Product designer with 7 years/);
    expect(draft.yearsExperience).toBe(7);
    expect(draft.portfolio.map((link) => link.title)).toEqual(["Dribbble", "GitHub"]);
  });

  it("finds roles with their dates, current roles, and descriptions", () => {
    const draft = parseResumeText(designer, SKILLS, 2026);
    expect(draft.experience).toEqual([
      expect.objectContaining({ title: "Lead Product Designer", organization: "Paystack", startYear: 2021, endYear: null }),
      expect.objectContaining({ title: "Product Designer", organization: "Andela", startYear: 2018, endYear: 2020 }),
    ]);
    expect(draft.experience[0].description).toBe("Redesigned the merchant dashboard used by 200k businesses. Built the company design system in Figma.");
  });

  it("matches only skills from Jobba's list, including common synonyms", () => {
    expect(parseResumeText(designer, SKILLS, 2026).skills).toEqual(expect.arrayContaining(["Figma", "Design systems", "User research", "Product design"]));
    const skills = parseResumeText(engineer, SKILLS, 2026).skills;
    expect(skills).toEqual(expect.arrayContaining(["Smart Contracts", "Solana", "TypeScript", "PostgreSQL", "SQL", "Backend Development"]));
    expect(skills).not.toContain("Copywriting");
  });

  it("handles 'Company | Title | dates' lines and derives years from the earliest role", () => {
    const draft = parseResumeText(engineer, SKILLS, 2026);
    expect(draft.displayName).toBe("Kofi Mensah");
    expect(draft.headline).toBe("Solana Engineer");
    expect(draft.location).toBe("Accra, Ghana");
    expect(draft.experience[0]).toEqual(expect.objectContaining({ title: "Smart Contract Engineer", organization: "Mesa Protocol", startYear: 2022, endYear: null }));
    expect(draft.experience[1]).toEqual(expect.objectContaining({ organization: "Freelance", startYear: 2019, endYear: 2022 }));
    expect(draft.yearsExperience).toBe(7);
  });

  it("never copies email or phone into any field and leaves unknowns empty", () => {
    const draft = parseResumeText("Just some text without structure.", SKILLS, 2026);
    expect(draft).toEqual({ displayName: null, headline: null, location: null, bio: null, yearsExperience: null, skills: [], experience: [], portfolio: [] });
    expect(JSON.stringify(parseResumeText(designer, SKILLS, 2026))).not.toMatch(/@example\.com|801 234/);
  });
});
