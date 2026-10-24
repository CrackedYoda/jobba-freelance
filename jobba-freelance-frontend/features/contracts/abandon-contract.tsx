"use client";

import { Undo2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function AbandonContract({ contractId, projectId }: { contractId: string; projectId: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string>();
  async function abandon() {
    if (!window.confirm("Choose a different proposal? This unfunded contract will be removed.")) return;
    setBusy(true); setError(undefined);
    const response = await fetch(`/api/v1/contracts/${contractId}/abandon`, { method: "POST" });
    const body = await response.json().catch(() => ({})) as { error?: { message?: string } };
    if (!response.ok) { setBusy(false); setError(body.error?.message ?? "Could not abandon the contract."); return; }
    router.push(`/projects/${projectId}`);
    router.refresh();
  }
  return (
    <section className="workspacePanel">
      <span className="eyebrow">Not funded yet</span>
      <h2>Changed your mind?</h2>
      <p className="mutedCopy">Nothing has moved on-chain. You can go back and select a different proposal.</p>
      <button className="button ghost" type="button" disabled={busy} onClick={() => void abandon()}><Undo2 size={16} aria-hidden="true" />{busy ? "Removing…" : "Choose a different proposal"}</button>
      {error ? <p className="formNotice error" role="alert">{error}</p> : null}
    </section>
  );
}
