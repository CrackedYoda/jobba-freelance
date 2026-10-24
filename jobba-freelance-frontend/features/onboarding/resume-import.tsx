"use client";

import { Check, FileText, Loader2, Sparkles, Upload } from "lucide-react";
import { useState } from "react";

import type { ResumeDraft } from "@/features/onboarding/merge-resume";

/** Upload a résumé; the server stores it privately and returns a draft that the wizard merges into the profile. */
export function ResumeImport({ current, onDraft }: { current: { fileName: string; uploadedAt: string } | null; onDraft: (draft: ResumeDraft) => string[] }) {
  const [state, setState] = useState<"idle" | "reading" | "done" | "error">("idle");
  const [message, setMessage] = useState<string>();
  const [filled, setFilled] = useState<string[]>([]);
  const [fileName, setFileName] = useState(current?.fileName);

  async function upload(file: File) {
    setState("reading"); setMessage(undefined); setFilled([]);
    try {
      const form = new FormData(); form.set("file", file);
      const response = await fetch("/api/v1/onboarding/resume", { method: "POST", body: form });
      const body = await response.json().catch(() => ({})) as { data?: { draft: ResumeDraft | null; fileName: string; method: "claude" | "rules" }; error?: { message?: string } };
      if (!response.ok || !body.data) throw new Error(body.error?.message ?? "We couldn't read that résumé.");
      setFileName(body.data.fileName);
      if (!body.data.draft) {
        setState("done");
        setMessage("Résumé saved. Automatic fill-in isn't available right now, so please fill in the details below.");
        return;
      }
      const applied = onDraft(body.data.draft);
      setFilled(applied);
      setState("done");
      setMessage(applied.length ? "Review the details below — you can edit anything." : "Your profile already had these details, so nothing was changed.");
    } catch (error) {
      setState("error");
      setMessage(error instanceof Error ? error.message : "We couldn't read that résumé.");
    }
  }

  return (
    <div className={`resumeImport${state === "done" ? " done" : ""}`}>
      <span className="resumeIcon" aria-hidden="true">{state === "reading" ? <Loader2 size={22} className="spin" /> : state === "done" ? <Check size={22} /> : <Sparkles size={22} />}</span>
      <div className="resumeCopy">
        <strong>{state === "reading" ? "Reading your résumé…" : state === "done" ? (filled.length ? "Profile filled from your résumé" : "Résumé saved") : "Start from your résumé"}</strong>
        <span>
          {state === "reading" ? "This takes a few seconds."
            : state === "done" ? message
            : "Upload a PDF or Word file and we'll fill in your headline, bio, skills and experience. Nothing is published until you finish sign-up."}
        </span>
        {filled.length ? <ul className="resumeFilled">{filled.map((item) => <li key={item}><Check size={13} aria-hidden="true" />{item}</li>)}</ul> : null}
        {fileName ? <a className="resumeFile" href="/api/v1/onboarding/resume"><FileText size={14} aria-hidden="true" />{fileName}</a> : null}
        {state === "error" && message ? <span className="resumeError" role="alert">{message}</span> : null}
      </div>
      <label className={`button ${state === "done" || fileName ? "ghost" : "primary"} compactButton fileButton`} aria-disabled={state === "reading"}>
        <Upload size={15} aria-hidden="true" />{fileName ? "Upload another" : "Upload résumé"}
        <input type="file" accept="application/pdf,.pdf,.docx,application/vnd.openxmlformats-officedocument.wordprocessingml.document" disabled={state === "reading"}
          onChange={(event) => { const file = event.target.files?.[0]; if (file) void upload(file); event.target.value = ""; }} />
      </label>
    </div>
  );
}
