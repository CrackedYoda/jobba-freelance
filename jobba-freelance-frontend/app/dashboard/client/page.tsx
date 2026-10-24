import Link from "next/link";
import { DashboardShell } from "@/components/marketplace/dashboard-shell";
import { SiteHeader } from "@/components/marketplace/site-header";
import { getDashboardSnapshot } from "@/server/marketplace/service";

export default async function ClientDashboardPage() {
  const data = await getDashboardSnapshot();
  return <><SiteHeader /><main><DashboardShell><header className="dashboardHeader"><div><p className="eyebrow">Client workspace</p><h1>Keep projects <em>moving.</em></h1></div><Link className="button primary" href="/projects/new">Post a project</Link></header><section className="statGrid"><article><span>Open projects</span><strong>{data.openProjects}</strong></article><article><span>Active contracts</span><strong>{data.activeContracts}</strong></article><article><span>Awaiting review</span><strong>{data.pendingProposals}</strong></article><article><span>Protected in escrow</span><strong>{data.escrowedUsdc} USDC</strong></article></section><section className="dashboardSection"><div className="sectionHeading"><div><p className="eyebrow">Recent activity</p><h2>Your projects</h2></div><Link href="/projects">Browse all →</Link></div><div className="tableWrap"><table><thead><tr><th>Project</th><th>Budget</th><th>Proposals</th><th>Status</th></tr></thead><tbody>{data.recentProjects.map((project) => <tr key={project.id}><td><Link href={`/projects/${project.id}`}>{project.title}</Link></td><td>{project.budgetUsdc} USDC</td><td>{project.proposals}</td><td><span className="statusPill">Open</span></td></tr>)}</tbody></table></div></section></DashboardShell></main></>;
}
