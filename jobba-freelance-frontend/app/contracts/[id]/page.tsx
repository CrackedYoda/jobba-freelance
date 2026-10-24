import { ChainStateCard } from "@/features/contracts/chain-state-card";
import { ContractChat } from "@/features/contracts/contract-chat";
import { EscrowActionPanel } from "@/features/contracts/escrow-action-panel";
import { SubmissionPanel } from "@/features/contracts/submission-panel";
import { ContractDisputePanel } from "@/features/disputes/contract-dispute-panel";

export default async function ContractWorkspace({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <main className="shell pageSection contractWorkspace"><header className="workspaceHero"><span className="eyebrow">Contract workspace</span><h1>Work, messages, and escrow in one place.</h1><p>Contract <code>{id}</code></p></header><div className="workspaceGrid"><div className="workspaceMain"><SubmissionPanel contractId={id} /><ContractChat contractId={id} /><ContractDisputePanel contractId={id} /></div><aside className="workspaceRail"><ChainStateCard contractId={id} /><EscrowActionPanel contractId={id} /></aside></div></main>;
}
