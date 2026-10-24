import type { FreelancerProfile, ProjectSummary, ProposalSummary } from "./types";

export const projectFixtures: ProjectSummary[] = [
  {
    id: "11111111-1111-4111-8111-111111111111",
    title: "Redesign a Solana commerce dashboard",
    summary: "Turn a dense merchant console into a clear, responsive product experience.",
    description: "We need a product designer who can audit the current dashboard, establish a compact design system, and deliver responsive Figma screens for onboarding, payments, refunds, and reporting. Experience with financial products is useful; deep blockchain knowledge is optional.",
    category: "Design & Creative",
    skills: ["Product design", "Figma", "Design systems"],
    budgetUsdc: "2,400",
    deliveryDays: 21,
    proposals: 8,
    postedAt: "2 hours ago",
    client: { name: "Northstar Labs", location: "Lagos, NG", projectsFunded: 6, verified: true },
  },
  {
    id: "22222222-2222-4222-8222-222222222222",
    title: "Build an Anchor event indexer",
    summary: "Index program events into PostgreSQL with resilient backfill and reconciliation.",
    description: "Build a TypeScript service that reads finalized Anchor events, stores normalized records, and resumes safely after RPC interruptions. Include structured logs and a concise operations guide.",
    category: "Development",
    skills: ["Solana", "TypeScript", "PostgreSQL"],
    budgetUsdc: "4,800",
    deliveryDays: 30,
    proposals: 12,
    postedAt: "5 hours ago",
    client: { name: "Mesa Protocol", location: "Austin, US", projectsFunded: 11, verified: true },
  },
  {
    id: "33333333-3333-4333-8333-333333333333",
    title: "Launch copy for a developer platform",
    summary: "Shape product positioning and write a crisp launch site for technical buyers.",
    description: "Interview two founders, distill the product narrative, and write homepage, pricing, and launch announcement copy. The voice should be direct, technical, and credible.",
    category: "Writing",
    skills: ["Copywriting", "Developer tools", "Positioning"],
    budgetUsdc: "1,350",
    deliveryDays: 14,
    proposals: 5,
    postedAt: "Yesterday",
    client: { name: "Plaintext Systems", location: "Berlin, DE", projectsFunded: 3, verified: true },
  },
  {
    id: "44444444-4444-4444-8444-444444444444",
    title: "Run a four-week community sprint",
    summary: "Plan and operate a focused contributor program for an open-source project.",
    description: "Create the calendar, recruit contributors, host weekly calls, and deliver a final report covering participation and shipped contributions.",
    category: "Marketing",
    skills: ["Community", "Operations", "Web3"],
    budgetUsdc: "3,200",
    deliveryDays: 28,
    proposals: 16,
    postedAt: "2 days ago",
    client: { name: "Common Ground", location: "Remote", projectsFunded: 9, verified: true },
  },
];

export const freelancerFixtures: FreelancerProfile[] = [
  {
    id: "amara-okafor",
    name: "Amara Okafor",
    headline: "Product designer for complex financial tools",
    bio: "I turn operational software into calm, legible products. My recent work spans payments, analytics, and internal tools for distributed teams.",
    location: "Lagos, Nigeria",
    rateUsdc: "65",
    availability: "Available next week",
    skills: ["Product design", "Figma", "User research", "Design systems"],
    completedContracts: 24,
    earnedUsdc: "38,900",
    portfolio: [
      { label: "Payment operations case study", url: "https://example.com/case-study" },
      { label: "Design system overview", url: "https://example.com/system" },
    ],
  },
];

export const proposalFixtures: ProposalSummary[] = [
  {
    id: "proposal-amara",
    projectId: "11111111-1111-4111-8111-111111111111",
    projectTitle: "Redesign a Solana commerce dashboard",
    freelancer: { id: "amara-okafor", name: "Amara Okafor", headline: "Product designer for complex financial tools", completedContracts: 24 },
    amountUsdc: "2,350",
    deliveryDays: 18,
    coverLetter: "I would begin with a two-day workflow audit, then prototype the highest-risk payment states before expanding the system across the dashboard.",
    status: "pending",
  },
];
