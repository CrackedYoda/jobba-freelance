import Link from "next/link";
import { DashboardShell } from "@/components/marketplace/dashboard-shell";
import { SiteHeader } from "@/components/marketplace/site-header";
import { getDashboardSnapshot } from "@/server/marketplace/service";

export default async function FreelancerDashboardPage() {
  const data = await getDashboardSnapshot();
  return <><SiteHeader /><main><DashboardShell active="Proposals"><header className="dashboardHeader"><div><p className="eyebrow">Freelancer workspace</p><h1>Work worth <em>winning.</em></h1></div><Link className="button secondary" href="/projects">Find projects</Link></header><section className="statGrid"><article><span>Active contracts</span><strong>{data.activeContracts}</strong></article><article><span>Pending proposals</span><strong>{data.pendingProposals}</strong></article><article><span>Profile strength</span><strong>80%</strong></article><article><span>Available balance</span><strong>0 USDC</strong></article></section><section className="dashboardSection"><div className="sectionHeading"><div><p className="eyebrow">Pipeline</p><h2>Recent proposals</h2></div><Link href="/projects">Find another project →</Link></div>{data.proposals.map((proposal) => <article className="proposalCard detailedProposal" key={proposal.id}><div className="listingTopline"><span className="statusPill">{proposal.status}</span><span>{proposal.amountUsdc} USDC · {proposal.deliveryDays} days</span></div><h3><Link href={`/projects/${proposal.projectId}`}>{proposal.projectTitle}</Link></h3><p>{proposal.coverLetter}</p></article>)}</section></DashboardShell></main></>;
}
