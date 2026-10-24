import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { ApiForm } from "@/components/marketplace/api-form";
import { SiteHeader } from "@/components/marketplace/site-header";
import { getProject } from "@/server/marketplace/service";

export const metadata: Metadata = { title: "Send a proposal | Jobba" };

export default async function NewProposalPage({ params }: { params: Promise<{ projectId: string }> }) {
  const project = await getProject((await params).projectId);
  if (!project) notFound();
  return (
    <><SiteHeader /><main className="shell narrowPage pageSection"><div className="pageIntro"><p className="eyebrow">Proposal</p><h1>Make your case <em>clearly.</em></h1><p className="lead">For “{project.title}”</p></div>
      <ApiForm endpoint={`/api/v1/projects/${project.id}/proposals`} kind="proposal">
        <fieldset><legend>Your terms</legend><label><span>Proposal amount</span><div className="inputSuffix"><input name="amount" required inputMode="decimal" pattern="\d+(\.\d{1,6})?" defaultValue={project.budgetUsdc.replace(",", "")} /><b>USDC</b></div></label><label><span>Delivery estimate</span><div className="inputSuffix"><input name="deliveryDays" type="number" required min={1} max={365} defaultValue={project.deliveryDays} /><b>days</b></div></label><label className="fullField"><span>Cover letter</span><textarea name="coverLetter" required minLength={80} maxLength={3000} rows={9} placeholder="Explain your approach, relevant experience, and the first step you would take." /></label><label className="fullField"><span>Relevant portfolio URL <i>(optional)</i></span><input name="portfolioUrl" type="url" placeholder="https://" /></label></fieldset>
        <aside className="escrowNotice"><strong>What happens next</strong><p>If the client chooses you, they fund the agreed USDC amount into escrow. You review the funded contract and accept before work starts.</p></aside>
        <div className="formActions"><button className="button primary" type="submit">Submit proposal</button></div>
      </ApiForm>
    </main></>
  );
}
