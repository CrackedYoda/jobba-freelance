import Image from "next/image";
import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { ContentPage } from "@/components/marketplace/content-page";
import { StatusChip } from "@/components/marketplace/status-chip";
import { getViewer, listMyContracts, relativeDate } from "@/server/marketplace/service";

export const metadata: Metadata = { title: "Contracts | Jobba" };

export default async function ContractsPage() {
  const viewer = await getViewer();
  if (!viewer.authenticated) redirect("/connect");
  const contracts = await listMyContracts(viewer.profileId!);
  return (
    <ContentPage eyebrow="Contracts" title="Your contracts" lead="Every contract you are part of, as client or freelancer.">
      {contracts.length ? (
        <div className="tableWrap"><table><thead><tr><th>Project</th><th>Your role</th><th>With</th><th>Amount</th><th>Status</th><th>Updated</th></tr></thead><tbody>
          {contracts.map((c) => <tr key={c.id}><td><Link href={`/contracts/${c.id}`}>{c.projectTitle}</Link></td><td>{c.role === "client" ? "Client" : "Freelancer"}</td><td>{c.counterpart}</td><td className="mono">{c.amountUsdc} USDC</td><td><StatusChip status={c.status} /></td><td className="muted">{relativeDate(c.updatedAt)}</td></tr>)}
        </tbody></table></div>
      ) : (
        <div className="emptyState"><Image src="/illustrations/empty-contracts.svg" alt="" width={220} height={160} unoptimized /><h2>No contracts yet</h2><p>Contracts appear here once a client selects a proposal.</p><div className="pillRow" style={{ justifyContent: "center" }}><Link className="button primary" href="/projects">Find work</Link><Link className="button ghost" href="/projects/new">Post a project</Link></div></div>
      )}
    </ContentPage>
  );
}
