import { describe, expect, it } from "vitest";

import { mergeResume, type ProfileFields, type ResumeDraft } from "../../features/onboarding/merge-resume";

const profile: ProfileFields = {
  displayName: "Wallet 9xQe…VFin", headline: null, location: "Lagos, NG", bio: "", yearsExperience: null,
  skillIds: ["s-figma"], availableSkills: [{ id: "s-figma", name: "Figma" }, { id: "s-design", name: "Product design" }, { id: "s-ts", name: "TypeScript" }],
  experience: [{ title: "Designer", organization: "Acme", startYear: 2019, endYear: 2021, description: null }, { title: "", organization: "", startYear: 2024, endYear: null }],
  portfolio: [{ title: "Site", url: "https://amara.example" }],
};
const draft: ResumeDraft = {
  displayName: "Amara Okafor", headline: "Product designer for fintech", location: "Abuja, NG", bio: "I design calm payment tools for teams across Africa.",
  yearsExperience: 7, skills: ["Figma", "product design", "Rust"],
  experience: [{ title: "Designer", organization: "Acme", startYear: 2019, endYear: 2021, description: null }, { title: "Lead designer", organization: "Paystack", startYear: 2021, endYear: null, description: "Led the dashboard." }],
  portfolio: [{ title: "Site", url: "https://amara.example" }, { title: "Dribbble", url: "https://dribbble.com/amara" }],
};

describe("résumé merge", () => {
  it("fills only empty fields and never overwrites typed values", () => {
    const { patch } = mergeResume(profile, draft);
    expect(patch.displayName).toBe("Amara Okafor");
    expect(patch.headline).toBe("Product designer for fintech");
    expect(patch.bio).toBe(draft.bio);
    expect(patch.yearsExperience).toBe(7);
    expect(patch.location).toBeUndefined();
  });

  it("adds new skills from the allowed list, case-insensitively, and ignores unknown ones", () => {
    expect(mergeResume(profile, draft).patch.skillIds).toEqual(["s-figma", "s-design"]);
  });

  it("appends only new roles and links, dropping blank rows", () => {
    const { patch, changed } = mergeResume(profile, draft);
    expect(patch.experience?.map((item) => item.organization)).toEqual(["Acme", "Paystack"]);
    expect(patch.portfolio?.map((link) => link.url)).toEqual(["https://amara.example", "https://dribbble.com/amara"]);
    expect(changed).toEqual(["Name", "Headline", "About you", "Years of experience", "1 skill", "1 role of experience", "1 portfolio link"]);
  });

  it("changes nothing when the profile already has everything", () => {
    const full: ProfileFields = { ...profile, displayName: "Amara O.", headline: "Designer", bio: "Already written bio text.", yearsExperience: 5, skillIds: ["s-figma", "s-design"], experience: draft.experience, portfolio: draft.portfolio };
    expect(mergeResume(full, draft)).toEqual({ patch: {}, changed: [] });
  });
});
