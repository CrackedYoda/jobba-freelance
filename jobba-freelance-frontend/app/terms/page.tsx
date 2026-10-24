import type { Metadata } from "next";
import Link from "next/link";

import { ContentPage } from "@/components/marketplace/content-page";

export const metadata: Metadata = { title: "Terms of service | Jobba" };

const sections: [string, string][] = [
  ["Your account", "Your Solana wallet is your account. You are responsible for keeping your wallet and its keys secure. Jobba cannot recover a lost wallet or reverse a transaction you sign."],
  ["Contracts", "A contract is formed when a client funds escrow for a selected proposal and the freelancer accepts it. The agreed scope is the project description and the accepted proposal."],
  ["Payments and escrow", "Contract amounts are held by the Jobba escrow program until release, cancellation before acceptance, or dispute resolution. Network fees are paid by the signer of each transaction."],
  ["Fees", "Jobba charges the platform fee shown before you send a proposal or fund a contract. It is deducted from the freelancer's payout on release."],
  ["Disputes", "Either party may open a dispute on a funded contract. An administrator decides it by refunding the client or awarding the freelancer in full, and that decision is final within Jobba."],
  ["Acceptable use", "Do not post unlawful work, misrepresent your identity or skills, share private credentials in projects or chat, or attempt to move contract payments off the platform to avoid escrow."],
  ["Changes", "We may update these terms. Continuing to use Jobba after an update means you accept the new terms."],
];

export default function TermsPage() {
  return (
    <ContentPage eyebrow="Legal" title="Terms of service" lead="The rules for using Jobba as a client or a freelancer.">
      <article className="docBody legal">
        {sections.map(([title, text]) => <section key={title}><h2>{title}</h2><p>{text}</p></section>)}
        <p className="muted">See also the <Link className="inlineLink" href="/privacy">privacy policy</Link> and the <Link className="inlineLink" href="/help/escrow">escrow guide</Link>.</p>
      </article>
    </ContentPage>
  );
}
