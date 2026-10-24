import { describe, expect, it } from "vitest";

import { missingRequirements, saveOnboardingSchema } from "@/features/profiles/schemas";

const freelancerDraft = {
  roles: ["freelancer"], displayName: "Amara Okafor", headline: "Product designer", location: "Lagos, NG",
  bio: "I design clear tools for complex financial workflows and payments teams.", hourlyRateAtomic: "65000000",
  skillIds: ["10000000-0000-4000-8000-000000000005"], experience: [{}], yearsExperience: 6, hasAvatar: true,
  companyName: null, industry: null, companySize: null, hiringNeeds: null,
};

describe("onboarding schema", () => {
  it("accepts partial saves so sign-up can resume", () => {
    expect(saveOnboardingSchema.parse({ step: 2, displayName: "Amara" })).toMatchObject({ step: 2, displayName: "Amara" });
  });

  it("rejects admin role assignment, unsafe URLs and inverted experience years", () => {
    expect(() => saveOnboardingSchema.parse({ roles: ["admin"] })).toThrow();
    expect(() => saveOnboardingSchema.parse({ portfolio: [{ title: "bad", url: "javascript:alert(1)" }] })).toThrow();
    expect(() => saveOnboardingSchema.parse({ experience: [{ title: "Designer", organization: "Acme", startYear: 2022, endYear: 2020 }] })).toThrow();
  });

  it("treats an empty company website as cleared", () => {
    expect(saveOnboardingSchema.parse({ companyWebsite: "" }).companyWebsite).toBeNull();
  });
});

describe("sign-up completeness", () => {
  it("passes a complete freelancer", () => {
    expect(missingRequirements(freelancerDraft)).toEqual([]);
  });

  it("requires photo, skills and experience for freelancers", () => {
    const fields = missingRequirements({ ...freelancerDraft, hasAvatar: false, skillIds: [], experience: [] }).map((item) => item.field);
    expect(fields).toEqual(expect.arrayContaining(["avatar", "skillIds", "experience"]));
  });

  it("requires company details for clients", () => {
    const fields = missingRequirements({ ...freelancerDraft, roles: ["client"] }).map((item) => item.field);
    expect(fields).toEqual(expect.arrayContaining(["companyName", "industry", "companySize", "hiringNeeds"]));
  });

  it("does not accept the placeholder wallet name", () => {
    expect(missingRequirements({ ...freelancerDraft, displayName: "Wallet 7xKX…gAsU" }).map((item) => item.field)).toContain("displayName");
  });
});
