import { AdminDisputeList } from "@/features/disputes/admin-dispute-list";

export default function AdminDisputesPage() {
  return <main className="shell pageSection"><header className="dashboardHeader"><div><span className="eyebrow">Escrow operations</span><h1>Resolve disputes with a clear audit trail.</h1><p>Review contract terms and submitted evidence before preparing an irreversible devnet settlement.</p></div></header><AdminDisputeList /></main>;
}
