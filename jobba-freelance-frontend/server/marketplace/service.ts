import "server-only";

import { freelancerFixtures, projectFixtures, proposalFixtures } from "@/features/marketplace/fixtures";
import type { ProjectSummary } from "@/features/marketplace/types";
import { formatUsdc } from "@/lib/domain/marketplace";
import { hasDatabase } from "@/server/db/client";
import { browseProjects, readProject } from "@/server/services/projects";

function relativeDate(value: string): string {
  const elapsed = Math.max(0, Date.now() - Date.parse(value));
  const hours = Math.floor(elapsed / 3_600_000);
  if (hours < 1) return "Just now";
  if (hours < 24) return `${hours} hour${hours === 1 ? "" : "s"} ago`;
  const days = Math.floor(hours / 24);
  return `${days} day${days === 1 ? "" : "s"} ago`;
}

function projectView(project: Record<string, unknown>): ProjectSummary {
  const client = (project.client ?? {}) as Record<string, unknown>;
  const skills = Array.isArray(project.skills) ? project.skills as Array<Record<string, unknown>> : [];
  const description = String(project.description ?? "");
  return {
    id: String(project.id),
    title: String(project.title),
    summary: description.length > 180 ? `${description.slice(0, 177)}…` : description,
    description,
    category: String(project.category),
    skills: skills.map((skill) => String(skill.name ?? skill.slug ?? "")).filter(Boolean),
    budgetUsdc: formatUsdc(BigInt(String(project.budgetAtomic))),
    deliveryDays: 30,
    proposals: 0,
    postedAt: relativeDate(String(project.publishedAt ?? project.createdAt)),
    client: {
      name: String(client.display_name ?? "Wallet client"),
      location: "Remote",
      projectsFunded: 0,
      verified: true,
    },
  };
}

export async function listProjects(query?: string, category?: string) {
  if (hasDatabase()) {
    const result = await browseProjects({
      limit: 50,
      ...(query?.trim() ? { search: query.trim() } : {}),
      ...(category && category !== "All" ? { category } : {}),
    });
    return result.items.map((project) => projectView(project as Record<string, unknown>));
  }
  const normalized = query?.trim().toLowerCase();
  return projectFixtures.filter((project) => {
    const matchesQuery = !normalized || [project.title, project.summary, ...project.skills].join(" ").toLowerCase().includes(normalized);
    const matchesCategory = !category || category === "All" || project.category === category;
    return matchesQuery && matchesCategory;
  });
}

export async function getProject(projectId: string) {
  if (hasDatabase()) {
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(projectId)) return undefined;
    try {
      return projectView(await readProject(projectId) as Record<string, unknown>);
    } catch (error) {
      if (error instanceof Error && "code" in error && error.code === "PROJECT_NOT_FOUND") return undefined;
      throw error;
    }
  }
  return projectFixtures.find((project) => project.id === projectId);
}

export async function getFreelancer(profileId: string) {
  if (hasDatabase()) return undefined;
  return freelancerFixtures.find((profile) => profile.id === profileId);
}

export async function listProjectProposals(projectId: string) {
  if (hasDatabase()) return [];
  return proposalFixtures.filter((proposal) => proposal.projectId === projectId);
}

export async function getDashboardSnapshot() {
  if (hasDatabase()) return { activeContracts: 0, openProjects: 0, pendingProposals: 0, escrowedUsdc: "0", recentProjects: [], proposals: [] };
  return {
    activeContracts: 2,
    openProjects: 3,
    pendingProposals: 5,
    escrowedUsdc: "6,250",
    recentProjects: projectFixtures.slice(0, 3),
    proposals: proposalFixtures,
  };
}
