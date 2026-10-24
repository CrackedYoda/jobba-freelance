import { AdminDisputeDetail } from "@/features/disputes/admin-dispute-detail";

export default async function AdminDisputePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <main className="shell pageSection"><header className="workspaceHero"><span className="eyebrow">Administrator review</span><h1>Evidence and escrow decision.</h1></header><AdminDisputeDetail disputeId={id} /></main>;
}
