export type ProjectSummary = {
  id: string;
  title: string;
  summary: string;
  description: string;
  category: string;
  skills: string[];
  budgetUsdc: string;
  deliveryDays: number;
  proposals: number;
  postedAt: string;
  client: { name: string; location: string; projectsFunded: number; verified: boolean };
};

export type FreelancerProfile = {
  id: string;
  name: string;
  headline: string;
  bio: string;
  location: string;
  rateUsdc: string;
  availability: string;
  skills: string[];
  completedContracts: number;
  earnedUsdc: string;
  portfolio: { label: string; url: string }[];
};

export type ProposalSummary = {
  id: string;
  projectId: string;
  projectTitle: string;
  freelancer: Pick<FreelancerProfile, "id" | "name" | "headline" | "completedContracts">;
  amountUsdc: string;
  deliveryDays: number;
  coverLetter: string;
  status: "pending" | "accepted" | "rejected" | "withdrawn";
};
