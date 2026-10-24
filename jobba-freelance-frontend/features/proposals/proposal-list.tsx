"use client";

import { Usdc } from "@/components/brand/usdc";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import type { ProposalSummary } from "@/features/marketplace/types";

export function ProposalList({ proposals, canSelect }: { proposals: ProposalSummary[]; canSelect: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState<string>();
  const [error, setError] = useState<string>();

  async function select(proposalId: string) {
    setBusy(proposalId); setError(undefined);
    try {
      const response = await fetch(`/api/v1/proposals/${proposalId}/fund`, { method: "POST", headers: { "Idempotency-Key": `select-${proposalId}` } });
      const body = await response.json().catch(() => ({})) as { data?: { id: string }; error?: { message?: string } };
      if (!response.ok || !body.data) throw new Error(body.error?.message ?? "Could not prepare the contract.");
      router.push(`/contracts/${body.data.id}`);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not prepare the contract.");
      setBusy(undefined);
    }
  }

  async function reject(proposalId: string) {
    if (!window.confirm("Decline this proposal? The freelancer will see it as rejected.")) return;
    setBusy(proposalId); setError(undefined);
    const response = await fetch(`/api/v1/proposals/${proposalId}/reject`, { method: "POST" });
    const body = await response.json().catch(() => ({})) as { error?: { message?: string } };
    setBusy(undefined);
    if (!response.ok) { setError(body.error?.message ?? "Could not decline the proposal."); return; }
    router.refresh();
  }

  if (!proposals.length) return <p className="muted">No proposals yet.</p>;
  return (
    <div className="proposalStack">
      {error ? <p className="formNotice error" role="alert">{error}</p> : null}
      {proposals.map((proposal) => (
        <article className="proposalCard" key={proposal.id}>
          <div className="between">
            <div><strong><Link href={`/freelancers/${proposal.freelancer.id}`}>{proposal.freelancer.name}</Link></strong><span className="muted"> · {proposal.freelancer.completedContracts} completed</span><br /><span className="muted">{proposal.freelancer.headline}</span></div>
            <span className={`chip ${proposal.status}`}>{proposal.status.toUpperCase()}</span>
          </div>
          <p>{proposal.coverLetter}</p>
          {proposal.portfolioUrl ? <a className="verifiedMark" href={proposal.portfolioUrl} target="_blank" rel="noreferrer noopener">Portfolio ↗</a> : null}
          <footer>
            <span><b className="mono">{proposal.amountUsdc} <Usdc /></b> · {proposal.deliveryDays} days</span>
            {proposal.contractId ? <Link className="button ghost compactButton" href={`/contracts/${proposal.contractId}`}>Open contract</Link>
              : canSelect && proposal.status === "pending" ? (
                <span className="pillRow">
                  <button className="button ghost compactButton" type="button" disabled={busy !== undefined} onClick={() => void reject(proposal.id)}>Decline</button>
                  <button className="button lime compactButton" type="button" disabled={busy !== undefined} onClick={() => void select(proposal.id)}>{busy === proposal.id ? "Preparing…" : "Select & fund"}</button>
                </span>
              ) : null}
          </footer>
        </article>
      ))}
    </div>
  );
}
