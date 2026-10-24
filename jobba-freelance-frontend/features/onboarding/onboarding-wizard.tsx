"use client";

/* eslint-disable @next/next/no-img-element -- avatars are served from our own API with cache-busting query strings */
import { BadgeCheck, Briefcase, Camera, Check, CircleAlert, Loader2, Plus, RefreshCw, Trash2, UserCheck, Wallet } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { missingRequirements } from "@/features/profiles/schemas";
import { hardNavigate } from "@/lib/navigation";

type Role = "client" | "freelancer";
type Experience = { title: string; organization: string; startYear: number; endYear: number | null; description?: string | null };
type Balance = { lamports: string; sol?: string; usdCents: string; minimumUsdCents?: string; sufficient?: boolean; cluster?: string };
type State = {
  status: string; step: number; walletAddress: string; avatarUrl: string | null; roles: Role[]; displayName: string; location: string | null;
  headline: string | null; bio: string | null; hourlyRateAtomic: string | null; availability: string; yearsExperience: number | null;
  skillIds: string[]; experience: Experience[]; portfolio: { title: string; url: string }[]; companyName: string | null; companyWebsite: string | null;
  companySize: string | null; industry: string | null; hiringNeeds: string | null; hasAvatar: boolean; balance: Balance | null;
  availableSkills: { id: string; name: string; category: string }[];
};

const industries = ["Software & SaaS", "Web3 & crypto", "Finance & fintech", "E-commerce & retail", "Media & entertainment", "Education", "Health", "Agency & services", "Non-profit", "Other"];
const sizes = [["solo", "Just me"], ["2-10", "2–10"], ["11-50", "11–50"], ["51-200", "51–200"], ["201+", "201+"]] as const;
const thisYear = new Date().getFullYear();

function rateToUsdc(atomic: string | null): string {
  if (!atomic) return "";
  const n = BigInt(atomic);
  const fraction = (n % 1_000_000n).toString().padStart(6, "0").replace(/0+$/, "");
  return `${n / 1_000_000n}${fraction ? `.${fraction}` : ""}`;
}
function usdcToRate(value: string): string | null | undefined {
  if (!value.trim()) return null;
  if (!/^\d+(\.\d{1,6})?$/.test(value.trim())) return undefined;
  const [whole, fraction = ""] = value.trim().split(".");
  return (BigInt(whole) * 1_000_000n + BigInt(fraction.padEnd(6, "0"))).toString();
}
const dollars = (cents?: string) => `$${(Number(cents ?? 0) / 100).toFixed(2)}`;

async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { cache: "no-store", ...init });
  const body = await response.json().catch(() => ({})) as { data?: T; error?: { message?: string; details?: unknown } };
  if (!response.ok || body.data === undefined) throw Object.assign(new Error(body.error?.message ?? "Something went wrong."), { details: body.error?.details });
  return body.data;
}

export function OnboardingWizard() {
  const [state, setState] = useState<State>();
  const [stepIndex, setStepIndex] = useState(0);
  const [rate, setRate] = useState("");
  const [saving, setSaving] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [error, setError] = useState<string>();
  const [busy, setBusy] = useState(false);
  const [balance, setBalance] = useState<Balance | null>(null);
  const timer = useRef<number | undefined>(undefined);
  const pending = useRef<Record<string, unknown>>({});

  useEffect(() => {
    void api<State>("/api/v1/onboarding").then((data) => {
      setState(data); setRate(rateToUsdc(data.hourlyRateAtomic)); setBalance(data.balance); setStepIndex(data.step);
    }).catch((caught: Error) => setError(caught.message));
  }, []);

  const roles = useMemo(() => state?.roles ?? [], [state?.roles]);
  const steps = useMemo(() => [
    { key: "role", label: "Your role" },
    { key: "profile", label: "Profile" },
    ...(roles.includes("freelancer") ? [{ key: "work", label: "Skills & experience" }] : []),
    ...(roles.includes("client") ? [{ key: "hiring", label: "Hiring" }] : []),
    { key: "verify", label: "Verify wallet" },
  ], [roles]);
  const current = steps[Math.min(stepIndex, steps.length - 1)];
  const completed = state?.status === "completed";

  const flush = useCallback(async (extra: Record<string, unknown> = {}) => {
    window.clearTimeout(timer.current);
    const payload = { ...pending.current, ...extra };
    pending.current = {};
    if (!Object.keys(payload).length) return;
    setSaving("saving");
    try {
      const next = await api<State>("/api/v1/onboarding", { method: "PATCH", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) });
      setState((prev) => prev ? { ...next, displayName: prev.displayName === next.displayName ? next.displayName : prev.displayName } : next);
      setSaving("saved");
    } catch (caught) {
      setSaving("error"); setError((caught as Error).message);
    }
  }, []);

  /** Local edit plus a debounced save, so closing the tab never loses more than a second of typing. */
  function update(patch: Partial<State>, server: Record<string, unknown> = patch) {
    setState((prev) => prev ? { ...prev, ...patch } : prev);
    pending.current = { ...pending.current, ...server };
    setSaving("saving");
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => void flush(), 800);
  }

  useEffect(() => {
    const onHide = () => {
      if (!Object.keys(pending.current).length) return;
      void fetch("/api/v1/onboarding", { method: "PATCH", keepalive: true, headers: { "Content-Type": "application/json" }, body: JSON.stringify(pending.current) });
      pending.current = {};
    };
    window.addEventListener("pagehide", onHide);
    return () => window.removeEventListener("pagehide", onHide);
  }, []);

  async function go(index: number) {
    setError(undefined);
    const next = Math.max(0, Math.min(index, steps.length - 1));
    setStepIndex(next);
    await flush({ step: next });
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function upload(file: File) {
    setBusy(true); setError(undefined);
    try {
      const form = new FormData(); form.set("file", file);
      const result = await api<{ avatarUrl: string }>("/api/v1/onboarding/avatar", { method: "POST", body: form });
      setState((prev) => prev ? { ...prev, avatarUrl: result.avatarUrl, hasAvatar: true } : prev);
    } catch (caught) { setError((caught as Error).message); }
    finally { setBusy(false); }
  }

  async function checkBalance() {
    setBusy(true); setError(undefined);
    try { setBalance(await api<Balance>("/api/v1/onboarding/balance", { method: "POST" })); }
    catch (caught) { setError((caught as Error).message); }
    finally { setBusy(false); }
  }

  async function complete() {
    setBusy(true); setError(undefined);
    await flush();
    try {
      await api("/api/v1/onboarding/complete", { method: "POST" });
      hardNavigate("/dashboard");
    } catch (caught) {
      const details = (caught as { details?: { balance?: Balance } }).details;
      if (details?.balance) setBalance(details.balance);
      setError((caught as Error).message);
      setBusy(false);
    }
  }

  if (!state) return <p className={`formNotice${error ? " error" : ""}`} role={error ? "alert" : "status"}>{error ?? "Loading your sign-up…"}</p>;

  const missing = missingRequirements({ ...state, hourlyRateAtomic: state.hourlyRateAtomic });
  const toggleRole = (role: Role) => {
    const next = roles.includes(role) ? roles.filter((r) => r !== role) : [...roles, role];
    update({ roles: next }, { roles: next.length ? next : undefined, primaryRole: next[0] });
  };
  const setExperience = (experience: Experience[]) => update({ experience }, { experience: experience.filter((e) => e.title.trim() && e.organization.trim()) });
  const setPortfolio = (portfolio: { title: string; url: string }[]) => update({ portfolio }, { portfolio: portfolio.filter((p) => p.title.trim() && /^https?:\/\/.+/.test(p.url.trim())) });

  return (
    <div className="wizard">
      <ol className="wizardSteps" aria-label="Sign-up progress">
        {steps.map((step, index) => (
          <li key={step.key} className={index < stepIndex ? "done" : index === stepIndex ? "now" : ""}>
            <button type="button" onClick={() => void go(index)} disabled={!completed && index > stepIndex} aria-current={index === stepIndex ? "step" : undefined}>
              <span className="mono">{index < stepIndex || (completed && index !== stepIndex) ? <Check size={14} aria-hidden="true" /> : index + 1}</span>{step.label}
            </button>
          </li>
        ))}
      </ol>
      <p className="saveState" role="status" aria-live="polite">
        {saving === "saving" ? <><Loader2 size={14} className="spin" aria-hidden="true" /> Saving…</> : saving === "error" ? <><CircleAlert size={14} aria-hidden="true" /> Not saved</> : <><Check size={14} aria-hidden="true" /> {completed ? "All changes saved" : "Progress saved — you can close this page and continue later"}</>}
      </p>

      {current.key === "role" ? (
        <section className="wizardCard">
          <h2>How will you use Jobba?</h2>
          <p className="muted">Choose one or both. You can change this later.</p>
          <div className="roleChoice">
            {([["client", "I’m hiring", "Post projects, review proposals and fund escrow.", Briefcase], ["freelancer", "I’m freelancing", "Build a profile, send proposals and get paid in USDC.", UserCheck]] as const).map(([role, title, copy, Icon]) => (
              <button type="button" key={role} className={`roleCard${roles.includes(role) ? " on" : ""}`} aria-pressed={roles.includes(role)} onClick={() => toggleRole(role)}>
                <Icon size={24} aria-hidden="true" /><strong>{title}</strong><span>{copy}</span>
              </button>
            ))}
          </div>
        </section>
      ) : null}

      {current.key === "profile" ? (
        <section className="wizardCard">
          <h2>{roles.includes("client") && !roles.includes("freelancer") ? "About you and your company" : "Your profile"}</h2>
          <div className="avatarField">
            {state.avatarUrl ? <img src={state.avatarUrl} alt="Your profile photo" className="avatarPreview" /> : <span className="avatarPreview empty" aria-hidden="true"><Camera size={28} /></span>}
            <div>
              <label className="button ghost compactButton fileButton">{busy ? "Uploading…" : state.avatarUrl ? "Change photo" : "Upload photo"}
                <input type="file" accept="image/png,image/jpeg,image/webp" onChange={(e) => { const file = e.target.files?.[0]; if (file) void upload(file); e.target.value = ""; }} disabled={busy} />
              </label>
              <p className="muted small">{roles.includes("client") && !roles.includes("freelancer") ? "A photo or company logo." : "A clear photo of you."} PNG, JPEG or WebP up to 2 MB.</p>
            </div>
          </div>
          <div className="twoCol">
            <label><span>Display name</span><input value={/^Wallet\s/.test(state.displayName) ? "" : state.displayName} placeholder="Your name" maxLength={80} onChange={(e) => update({ displayName: e.target.value }, e.target.value.trim() ? { displayName: e.target.value } : {})} /></label>
            <label><span>Location</span><input value={state.location ?? ""} placeholder="City, country" maxLength={120} onChange={(e) => update({ location: e.target.value })} /></label>
          </div>
          {roles.includes("freelancer") ? (
            <>
              <label><span>Professional headline</span><input value={state.headline ?? ""} placeholder="e.g. Product designer for fintech" maxLength={120} onChange={(e) => update({ headline: e.target.value })} /></label>
              <div className="threeCol">
                <label><span>Hourly guide (USDC)</span><input className="mono" inputMode="decimal" value={rate} placeholder="65" onChange={(e) => { setRate(e.target.value); const atomic = usdcToRate(e.target.value); if (atomic !== undefined) update({ hourlyRateAtomic: atomic }); }} /></label>
                <label><span>Years of experience</span><input type="number" min={0} max={60} value={state.yearsExperience ?? ""} onChange={(e) => update({ yearsExperience: e.target.value === "" ? null : Math.max(0, Math.min(60, Number(e.target.value))) })} /></label>
                <label><span>Availability</span><select value={state.availability} onChange={(e) => update({ availability: e.target.value })}><option value="available">Available now</option><option value="limited">Limited</option><option value="unavailable">Not available</option></select></label>
              </div>
            </>
          ) : null}
          {roles.includes("client") ? (
            <>
              <div className="twoCol">
                <label><span>Company or team name</span><input value={state.companyName ?? ""} placeholder="Use your own name if you hire as an individual" maxLength={120} onChange={(e) => update({ companyName: e.target.value })} /></label>
                <label><span>Website (optional)</span><input type="url" value={state.companyWebsite ?? ""} placeholder="https://" onChange={(e) => update({ companyWebsite: e.target.value }, /^https?:\/\/.+\..+/.test(e.target.value) || !e.target.value ? { companyWebsite: e.target.value } : {})} /></label>
              </div>
              <div className="twoCol">
                <label><span>Industry</span><select value={state.industry ?? ""} onChange={(e) => update({ industry: e.target.value || null })}><option value="">Choose…</option>{industries.map((item) => <option key={item}>{item}</option>)}</select></label>
                <label><span>Team size</span><select value={state.companySize ?? ""} onChange={(e) => update({ companySize: e.target.value || null })}><option value="">Choose…</option>{sizes.map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label>
              </div>
            </>
          ) : null}
        </section>
      ) : null}

      {current.key === "work" ? (
        <section className="wizardCard">
          <h2>Skills and experience</h2>
          <label><span>About you</span><textarea rows={5} value={state.bio ?? ""} maxLength={2000} placeholder="What you do best, who you have done it for, and how you work." onChange={(e) => update({ bio: e.target.value })} /><small className="counter">{(state.bio ?? "").length} / 2,000 · at least 40</small></label>
          <div><span className="fieldLabel">Skills</span>
            <div className="pillRow" role="group" aria-label="Skills">
              {state.availableSkills.map((skill) => { const on = state.skillIds.includes(skill.id); return <button key={skill.id} type="button" className={`pillButton${on ? " on" : ""}`} aria-pressed={on} onClick={() => update({ skillIds: on ? state.skillIds.filter((id) => id !== skill.id) : [...state.skillIds, skill.id] })}>{skill.name}</button>; })}
            </div>
          </div>
          <div><span className="fieldLabel">Relevant experience</span>
            <div className="experienceList">
              {state.experience.map((item, index) => (
                <fieldset className="experienceItem" key={index}>
                  <legend className="srOnly">Experience {index + 1}</legend>
                  <div className="twoCol">
                    <label><span>Role</span><input value={item.title} maxLength={120} onChange={(e) => setExperience(state.experience.map((x, i) => i === index ? { ...x, title: e.target.value } : x))} /></label>
                    <label><span>Company or client</span><input value={item.organization} maxLength={120} onChange={(e) => setExperience(state.experience.map((x, i) => i === index ? { ...x, organization: e.target.value } : x))} /></label>
                  </div>
                  <div className="threeCol">
                    <label><span>From</span><input type="number" min={1950} max={thisYear} value={item.startYear} onChange={(e) => setExperience(state.experience.map((x, i) => i === index ? { ...x, startYear: Number(e.target.value) } : x))} /></label>
                    <label><span>To</span><input type="number" min={1950} max={thisYear} value={item.endYear ?? ""} placeholder="Present" onChange={(e) => setExperience(state.experience.map((x, i) => i === index ? { ...x, endYear: e.target.value ? Number(e.target.value) : null } : x))} /></label>
                    <button type="button" className="button ghost compactButton" onClick={() => setExperience(state.experience.filter((_, i) => i !== index))}><Trash2 size={15} aria-hidden="true" />Remove</button>
                  </div>
                  <label><span>What you did (optional)</span><textarea rows={2} maxLength={2000} value={item.description ?? ""} onChange={(e) => setExperience(state.experience.map((x, i) => i === index ? { ...x, description: e.target.value } : x))} /></label>
                </fieldset>
              ))}
              {state.experience.length < 10 ? <button type="button" className="button ghost compactButton" onClick={() => setState({ ...state, experience: [...state.experience, { title: "", organization: "", startYear: thisYear - 1, endYear: null, description: "" }] })}><Plus size={15} aria-hidden="true" />Add experience</button> : null}
            </div>
          </div>
          <div><span className="fieldLabel">Portfolio links (optional)</span>
            <div className="experienceList">
              {state.portfolio.map((link, index) => (
                <div className="threeCol" key={index}>
                  <label><span>Title</span><input value={link.title} maxLength={120} onChange={(e) => setPortfolio(state.portfolio.map((x, i) => i === index ? { ...x, title: e.target.value } : x))} /></label>
                  <label><span>URL</span><input type="url" value={link.url} placeholder="https://" onChange={(e) => setPortfolio(state.portfolio.map((x, i) => i === index ? { ...x, url: e.target.value } : x))} /></label>
                  <button type="button" className="button ghost compactButton" onClick={() => setPortfolio(state.portfolio.filter((_, i) => i !== index))}><Trash2 size={15} aria-hidden="true" />Remove</button>
                </div>
              ))}
              {state.portfolio.length < 12 ? <button type="button" className="button ghost compactButton" onClick={() => setState({ ...state, portfolio: [...state.portfolio, { title: "", url: "" }] })}><Plus size={15} aria-hidden="true" />Add link</button> : null}
            </div>
          </div>
        </section>
      ) : null}

      {current.key === "hiring" ? (
        <section className="wizardCard">
          <h2>What will you hire for?</h2>
          <p className="muted">Freelancers see this on your projects. It helps the right people propose.</p>
          <label><span>Hiring needs</span><textarea rows={5} maxLength={2000} value={state.hiringNeeds ?? ""} placeholder="e.g. We ship a payments product and regularly need design, smart-contract and writing help." onChange={(e) => update({ hiringNeeds: e.target.value })} /><small className="counter">{(state.hiringNeeds ?? "").length} / 2,000 · at least 20</small></label>
        </section>
      ) : null}

      {current.key === "verify" ? (
        <section className="wizardCard">
          <h2>Verify your wallet</h2>
          <p className="muted">Your wallet ownership was verified when you signed in. To finish, it must hold at least $1 worth of SOL. We read the balance on-chain; nothing is moved or charged.</p>
          <div className="verifyRow"><Wallet size={20} aria-hidden="true" /><span className="mono wrap">{state.walletAddress}</span><span className="chip active"><BadgeCheck size={13} aria-hidden="true" />Signature verified</span></div>
          <div className={`balanceBox${balance ? (balance.sufficient ?? Number(balance.usdCents) >= 100 ? " ok" : " low") : ""}`}>
            {balance ? (
              <div><strong>{balance.sol ? `${balance.sol} SOL` : "Balance recorded"}</strong><span>≈ {dollars(balance.usdCents)} · minimum {dollars(balance.minimumUsdCents ?? "100")}</span></div>
            ) : <div><strong>Balance not checked yet</strong><span>Minimum $1.00 worth of SOL</span></div>}
            <button type="button" className="button ghost compactButton" disabled={busy} onClick={() => void checkBalance()}><RefreshCw size={15} aria-hidden="true" />{balance ? "Check again" : "Check balance"}</button>
          </div>
          {missing.length ? (
            <div className="missingBox"><strong>Still needed</strong><ul>{missing.map((item) => <li key={item.field}>{item.message}</li>)}</ul></div>
          ) : <p className="formNotice">All profile details are complete.</p>}
          {!completed ? <button type="button" className="button primary wide" disabled={busy || missing.length > 0} onClick={() => void complete()}>{busy ? "Checking…" : "Complete sign-up"}</button> : <Link className="button primary wide" href="/dashboard">Go to dashboard</Link>}
        </section>
      ) : null}

      {error ? <p className="formNotice error" role="alert">{error}</p> : null}
      <div className="wizardNav">
        {stepIndex > 0 ? <button type="button" className="button ghost" onClick={() => void go(stepIndex - 1)}>Back</button> : <span />}
        {current.key !== "verify" ? <button type="button" className="button primary" disabled={current.key === "role" && roles.length === 0} onClick={() => void go(stepIndex + 1)}>Continue</button> : null}
      </div>
    </div>
  );
}
