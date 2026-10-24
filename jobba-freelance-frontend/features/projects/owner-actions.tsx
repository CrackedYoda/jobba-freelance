"use client";

import { Ban, PencilLine, Send } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function ProjectOwnerActions({ projectId, status, hasContract }: { projectId: string; status: string; hasContract: boolean }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string>();

  async function change(next: "open" | "cancelled") {
    if (next === "cancelled" && !window.confirm("Cancel this project? Pending proposals will be declined.")) return;
    setBusy(true); setError(undefined);
    const response = await fetch(`/api/v1/projects/${projectId}`, { method: "PATCH", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ status: next }) });
    const body = await response.json().catch(() => ({})) as { error?: { message?: string } };
    setBusy(false);
    if (!response.ok) { setError(body.error?.message ?? "The project could not be updated."); return; }
    router.refresh();
  }

  if (!["draft", "open"].includes(status)) return null;
  return (
    <div className="ownerActions">
      {status === "draft" ? <button className="button lime" type="button" disabled={busy} onClick={() => void change("open")}><Send size={16} aria-hidden="true" />Publish</button> : null}
      <Link className="button ghost" href={`/projects/${projectId}/edit`}><PencilLine size={16} aria-hidden="true" />Edit</Link>
      {!hasContract ? <button className="button dangerButton" type="button" disabled={busy} onClick={() => void change("cancelled")}><Ban size={16} aria-hidden="true" />Cancel project</button> : null}
      {error ? <p className="formNotice error" role="alert">{error}</p> : null}
    </div>
  );
}
