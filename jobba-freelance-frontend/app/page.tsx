import Image from "next/image";
import { ArrowRight, Coins, ExternalLink, KeyRound, Lock, Vault } from "lucide-react";
import Link from "next/link";

import { NamedIcon } from "@/components/brand/icons";
import { Logo } from "@/components/brand/logo";
import { ProjectCard } from "@/components/marketplace/project-card";
import { SiteFooter } from "@/components/marketplace/site-footer";
import { SiteHeader } from "@/components/marketplace/site-header";
import { faq, getDoc } from "@/features/help/content";
import { HowItWorks, LandingHero } from "@/features/landing/hero";
import { hasDatabase } from "@/server/db/client";
import { listProjects } from "@/server/marketplace/service";

const categories = [
  ["Development", "Smart contracts, web apps, indexers"],
  ["Design & Creative", "Product, brand, motion"],
  ["Writing", "Docs, launch copy, research"],
  ["Marketing", "Growth, community, ops"],
  ["Data & AI", "Analytics, notebooks, models"],
  ["Video & animation", "Explainers and edits"],
];

export default async function Home() {
  const featured = hasDatabase() ? (await listProjects({})).slice(0, 3) : [];
  return (
    <main id="top">
      <SiteHeader />
      <LandingHero />
      <section className="trustStrip" aria-label="Platform facts">
        <div>
          <span><Coins size={18} aria-hidden="true" /><strong>USDC</strong> settlement, SOL for network fees</span>
          <span><KeyRound size={18} aria-hidden="true" /><strong>Sign-In With Solana</strong> — no passwords</span>
          <span><Vault size={18} aria-hidden="true" /><strong>One vault</strong> per contract</span>
          <span><ExternalLink size={18} aria-hidden="true" /><strong>Explorer links</strong> on every transaction</span>
        </div>
      </section>
      <HowItWorks />
      <section className="shell" style={{ paddingBlock: "40px 24px" }}>
        <div className="darkBand">
          <div className="splitIntro">
            <div><span className="eyebrow" style={{ color: "var(--lime)" }}>The escrow lifecycle</span><h2>Every state is on-chain. Every move needs the right signer.</h2></div>
            <p style={{ maxWidth: 400 }}>The program checks the actor, state, amount, mint and terms hash on every instruction. Release is locked the moment a dispute opens.</p>
          </div>
          <Image className="flowArt" src="/illustrations/escrow-flow.svg" alt="USDC moves from the client into the escrow vault, then to the freelancer on release." width={560} height={200} unoptimized />
          <div className="lifecycle" aria-label="Happy path"><span>Funded</span><ArrowRight size={16} aria-hidden="true" /><span>Active</span><ArrowRight size={16} aria-hidden="true" /><span>Submitted</span><ArrowRight size={16} aria-hidden="true" /><span className="last">Released</span></div>
          <div className="branchGrid">
            <div><small>FUNDED → CANCELLED</small><strong>Changed your mind before work starts?</strong><p>The client can cancel and recover the full amount while the escrow is still Funded and the freelancer has not accepted.</p></div>
            <div><small className="warn">ANY ACTIVE STATE → DISPUTED</small><strong>Something went wrong?</strong><p>Either party opens a dispute with evidence. Normal release locks; an admin reviews the terms, chat and deliverables.</p></div>
            <div><small>DISPUTED → REFUNDED | AWARDED</small><strong>A recorded, signed decision</strong><p>The admin refunds the client or awards the freelancer in full, with a written reason and an Explorer link in the case history.</p></div>
          </div>
          <Link className="button lime" href="/help/escrow" style={{ justifySelf: "start" }}><Lock size={16} aria-hidden="true" />Read how escrow works</Link>
        </div>
      </section>
      <section className="shell" style={{ paddingBlock: "56px 24px" }}>
        <div className="splitIntro"><div><span className="eyebrow">Browse by category</span><h2>Work that ships.</h2></div><Link className="button ghost" href="/projects">All projects</Link></div>
        <div className="catGrid">
          {categories.map(([name, blurb]) => <Link className="catCard" key={name} href={`/projects?category=${encodeURIComponent(name)}`}><span className="iconTile"><NamedIcon name={name} /></span><strong>{name}</strong><span>{blurb}</span></Link>)}
        </div>
      </section>
      {featured.length ? (
        <section className="shell" style={{ paddingBlock: "40px 24px" }}>
          <div className="splitIntro"><div><span className="eyebrow">Open now</span><h2>Fresh projects.</h2></div><Link className="button ghost" href="/projects">See all</Link></div>
          <div className="listingStack" style={{ marginTop: 24 }}>{featured.map((project) => <ProjectCard project={project} key={project.id} />)}</div>
        </section>
      ) : null}
      <section id="faq" className="shell faqSplit" style={{ paddingBlock: "72px 0" }}>
        <div><span className="eyebrow">FAQ</span><h2>Questions, answered plainly.</h2><p className="lead">Short answers here; full guides in the help center.</p><div className="pillRow"><Link className="button primary" href="/faq">All questions</Link><Link className="button ghost" href="/help">Help center</Link></div></div>
        <div>{faq.slice(0, 6).map((item) => <details className="faq" key={item.q}><summary>{item.q}</summary><p>{item.a} <Link className="inlineLink" href={`/help/${item.doc}`}>{getDoc(item.doc)?.title} →</Link></p></details>)}</div>
      </section>
      <section className="shell" style={{ paddingBlock: "72px 0" }}>
        <div className="ctaBand"><Logo onDark size={44} /><h2>Start with a wallet. Finish with a payment.</h2><div className="actions" style={{ margin: 0 }}><Link className="button lime big" href="/connect">Connect wallet</Link><Link className="button wline big" href="/projects">Browse projects</Link></div></div>
      </section>
      <SiteFooter />
    </main>
  );
}
