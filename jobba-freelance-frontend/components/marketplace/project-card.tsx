import Link from "next/link";

import type { ProjectSummary } from "@/features/marketplace/types";

export function ProjectCard({ project }: { project: ProjectSummary }) {
  return (
    <article className="listingCard">
      <div className="listingTopline"><span>{project.category}</span><span>{project.postedAt}</span></div>
      <h3><Link href={`/projects/${project.id}`}>{project.title}</Link></h3>
      <p>{project.summary}</p>
      <div className="tagRow" aria-label="Required skills">{project.skills.map((skill) => <span key={skill}>{skill}</span>)}</div>
      <div className="listingFooter">
        <strong>{project.budgetUsdc} USDC</strong>
        <span>{project.deliveryDays} days · {project.proposals} proposals</span>
      </div>
    </article>
  );
}
