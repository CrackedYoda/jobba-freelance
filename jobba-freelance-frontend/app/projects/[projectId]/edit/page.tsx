import type { Metadata } from "next";
import { notFound, redirect } from "next/navigation";

import { SiteFooter } from "@/components/marketplace/site-footer";
import { SiteHeader } from "@/components/marketplace/site-header";
import { ProjectForm } from "@/features/projects/project-form";
import { getProject, getViewer, listCategories, listSkillOptions } from "@/server/marketplace/service";

export const metadata: Metadata = { title: "Edit project | Jobba" };

export default async function EditProjectPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = await params;
  const viewer = await getViewer();
  if (!viewer.authenticated) redirect("/connect");
  const project = await getProject(projectId, viewer.profileId);
  if (!project || project.clientId !== viewer.profileId) notFound();
  if (!["draft", "open"].includes(project.status)) redirect(`/projects/${projectId}`);
  const [skills, categories] = await Promise.all([listSkillOptions(), listCategories()]);
  return (
    <>
      <SiteHeader />
      <main className="shell narrowPage pageSection">
        <div className="pageIntro"><p className="eyebrow">{project.status === "draft" ? "Draft" : "Open project"}</p><h1>Edit project</h1><p className="lead">{project.title}</p></div>
        <ProjectForm skills={skills.map(({ id, name }) => ({ id, name }))} categories={categories} initial={{
          id: project.id, title: project.title, category: project.category, description: project.description, budgetUsdc: project.budgetUsdc,
          deliveryDays: project.deliveryDays, skillIds: project.skillIds, status: project.status, locked: project.status === "open" && project.pendingProposals > 0,
        }} />
      </main>
      <SiteFooter />
    </>
  );
}
