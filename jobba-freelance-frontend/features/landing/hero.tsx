"use client";

import { BadgeCheck, CircleDollarSign, Handshake, Lock, PenSquare, Scale, Send, ShieldCheck, UserRound, Wallet, type LucideIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

const demoSteps = [
  { label: "Funded", chip: "funded", bar: 25, text: "The client locked the full amount in the program-owned vault. The freelancer can see it before accepting." },
  { label: "Active", chip: "active", bar: 50, text: "The freelancer accepted the funded contract. Work begins and chat opens for both parties." },
  { label: "Submitted", chip: "submitted", bar: 75, text: "Deliverables are submitted. The client reviews them; release is one signature away." },
  { label: "Released", chip: "released", bar: 100, text: "The client released payment. The freelancer received the amount minus the platform fee, with an Explorer link." },
];

const content = {
  hire: {
    h1: ["Hire anywhere.", "Pay on delivery."],
    lead: "Post a fixed-price project, pick a proposal, and lock the agreed USDC in a Solana escrow. Funds move to the freelancer only when you release them — or when an admin resolves a dispute.",
    primary: ["Post a project", "/projects/new"], secondary: ["Browse talent", "/freelancers"],
  },
  work: {
    h1: ["Get paid in USDC.", "Funded before you start."],
    lead: "Every contract is funded on-chain before you accept it. You can see the vault, the amount and the terms hash yourself — no invoices, no chasing, no bank wires.",
    primary: ["Find work", "/projects"], secondary: ["Create a profile", "/connect"],
  },
} as const;

export function LandingHero() {
  const [side, setSide] = useState<"hire" | "work">("hire");
  const [step, setStep] = useState(0);
  const c = content[side];
  const s = demoSteps[step];
  return (
    <section className="shell heroSplit" style={{ paddingBlock: "72px 56px" }}>
      <div className="heroCopy">
        <div className="segmented" role="group" aria-label="Choose your side">
          <button type="button" aria-pressed={side === "hire"} onClick={() => setSide("hire")}>I want to hire</button>
          <button type="button" aria-pressed={side === "work"} onClick={() => setSide("work")}>I want to work</button>
        </div>
        <h1>{c.h1[0]}<br /><em>{c.h1[1]}</em></h1>
        <p className="lead">{c.lead}</p>
        <div className="actions">
          <Link className="button primary big" href={c.primary[1]}>{c.primary[0]}</Link>
          <Link className="button ghost big" href={c.secondary[1]}>{c.secondary[0]}</Link>
        </div>
        <div className="trustRow" aria-label="Product benefits"><span><Lock size={14} aria-hidden="true" />Program-owned USDC vault</span><span><Wallet size={14} aria-hidden="true" />Sign in with your wallet</span><span><Scale size={14} aria-hidden="true" />Admin-reviewed disputes</span></div>
      </div>
      <aside aria-label="Escrow lifecycle preview">
        <div className="demoHead mono"><span><Lock size={12} aria-hidden="true" /> ESCROW · SOLANA</span><span>Program-owned vault</span></div>
        <div className="escrowDemo">
          <div className="demoTitle"><div><small>Example contract</small><strong>Website redesign, fixed price</strong></div><span className={`chip ${s.chip}`}><i />{s.label}</span></div>
          <p className="amount">2,350.00 <small>USDC</small></p>
          <div className="meter"><i style={{ width: `${s.bar}%` }} /></div>
          <p className="demoText">{s.text}</p>
          <div className="cardMeta"><span><UserRound size={13} aria-hidden="true" /> Client</span><span><UserRound size={13} aria-hidden="true" /> Freelancer</span></div>
        </div>
        <div className="escrowTabs" role="tablist" aria-label="Escrow lifecycle">
          {demoSteps.map((d, i) => (
            <button key={d.label} role="tab" type="button" aria-selected={i === step} onClick={() => setStep(i)}><span className="mono">0{i + 1}</span><br />{d.label}</button>
          ))}
        </div>
      </aside>
    </section>
  );
}

const flows: Record<"hire" | "work", [string, string, string, LucideIcon][]> = {
  hire: [
    ["Post", "Describe it", "Outcome, skills, fixed USDC budget and delivery target.", PenSquare],
    ["Choose", "Pick a proposal", "Compare focused proposals and select a freelancer.", Handshake],
    ["Fund", "Lock the USDC", "One signature moves the agreed amount into the vault.", Lock],
    ["Release", "Approve the work", "Review the delivery, then release — or open a dispute.", BadgeCheck],
  ],
  work: [
    ["Profile", "Show your craft", "Skills, rate, availability and portfolio links.", UserRound],
    ["Propose", "Send a proposal", "Your price and approach, no bidding credits.", Send],
    ["Accept", "See it funded", "Accept only after the vault shows the full amount.", ShieldCheck],
    ["Get paid", "Deliver and release", "Submit work; USDC lands in your wallet on release.", CircleDollarSign],
  ],
};

export function HowItWorks() {
  const [side, setSide] = useState<"hire" | "work">("hire");
  return (
    <section id="how" className="shell" style={{ paddingBlock: "96px 40px" }}>
      <div className="splitIntro">
        <div><span className="eyebrow">How it works</span><h2>From brief to paid, with the money visible the whole way.</h2></div>
        <div className="segmented" role="group" aria-label="Show steps for">
          <button type="button" aria-pressed={side === "hire"} onClick={() => setSide("hire")}>For clients</button>
          <button type="button" aria-pressed={side === "work"} onClick={() => setSide("work")}>For freelancers</button>
        </div>
      </div>
      <div className="stepGrid">
        {flows[side].map(([tag, title, body, Icon], i) => (
          <article className="step" key={title}><div className="stepTop"><span className="iconTile soft"><Icon size={20} aria-hidden="true" /></span><b>{tag}</b></div><span className="mono stepNo">0{i + 1}</span><h3>{title}</h3><p>{body}</p></article>
        ))}
      </div>
    </section>
  );
}
