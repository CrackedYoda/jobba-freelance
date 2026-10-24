import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { SiteHeader } from "@/components/marketplace/site-header";
import { getProject, listProjectProposals } from "@/server/marketplace/service";

export const metadata: Metadata = { title: "Project details | Jobba" };

export default async function ProjectDetailPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = await params;
  const project = await getProject(projectId);
  if (!project) notFound();
  const proposals = await listProjectProposals(projectId);
  return (
    <>
      <SiteHeader />
      <main className="shell detailLayout pageSection">
        <article className="detailContent">
          <Link className="backLink" href="/projects">← All projects</Link>
          <div className="detailTitle"><p className="eyebrow">{project.category}</p><h1>{project.title}</h1><p className="lead">{project.summary}</p></div>
          <section className="contentSection"><h2>Project brief</h2><p>{project.description}</p></section>
          <section className="contentSection"><h2>Skills and expertise</h2><div className="tagRow">{project.skills.map((skill) => <span key={skill}>{skill}</span>)}</div></section>
          <section className="contentSection"><h2>Current proposals</h2>{proposals.length ? proposals.map((proposal) => <article className="proposalCard" key={proposal.id}><div><strong>{proposal.freelancer.name}</strong><span>{proposal.freelancer.headline}</span></div><p>{proposal.coverLetter}</p><footer><span>{proposal.amountUsdc} USDC</span><span>{proposal.deliveryDays} days</span></footer></article>) : <p>No proposals yet.</p>}</section>
        </article>
        <aside className="detailRail">
          <p className="railAmount">{project.budgetUsdc} <small>USDC</small></p><span>Fixed price · {project.deliveryDays} day target</span>
          <Link className="button secondary fullButton" href={`/projects/${project.id}/proposals/new`}>Send a proposal</Link>
          <dl><div><dt>Proposals</dt><dd>{project.proposals}</dd></div><div><dt>Client</dt><dd>{project.client.name}</dd></div><div><dt>Location</dt><dd>{project.client.location}</dd></div><div><dt>Funded projects</dt><dd>{project.client.projectsFunded}</dd></div></dl>
          {project.client.verified ? <p className="verifiedMark">✓ Wallet-verified client</p> : null}
        </aside>
      </main>
    </>
  );
}
