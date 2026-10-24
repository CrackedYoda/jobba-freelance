import type { Metadata } from "next";
import Link from "next/link";

import { ProjectCard } from "@/components/marketplace/project-card";
import { SiteHeader } from "@/components/marketplace/site-header";
import { listProjects } from "@/server/marketplace/service";

export const metadata: Metadata = { title: "Find projects | Jobba" };

export default async function ProjectsPage({ searchParams }: { searchParams: Promise<{ q?: string; category?: string }> }) {
  const { q = "", category = "All" } = await searchParams;
  const projects = await listProjects(q, category);
  const categories = ["All", "Development", "Design & Creative", "Writing", "Marketing"];
  return (
    <>
      <SiteHeader />
      <main className="shell pageSection">
        <div className="pageIntro splitIntro">
          <div><p className="eyebrow">Open work</p><h1>Find your next <em>great contract.</em></h1><p className="lead">Clear scopes, fixed USDC budgets, and escrow-backed settlement.</p></div>
          <Link className="button primary" href="/onboarding">Create freelancer profile</Link>
        </div>
        <form className="searchPanel" role="search">
          <label><span>Search projects</span><input name="q" defaultValue={q} placeholder="Try Solana, product design, copywriting…" /></label>
          <label><span>Category</span><select name="category" defaultValue={category}>{categories.map((item) => <option key={item}>{item}</option>)}</select></label>
          <button className="button secondary" type="submit">Search</button>
        </form>
        <div className="resultsHeading"><h2>{projects.length} projects</h2><span>Newest first</span></div>
        <div className="listingStack">{projects.map((project) => <ProjectCard project={project} key={project.id} />)}</div>
        {projects.length === 0 ? <div className="emptyState"><h2>No exact matches</h2><p>Try a broader skill or choose all categories.</p><Link href="/projects">Clear filters</Link></div> : null}
      </main>
    </>
  );
}
