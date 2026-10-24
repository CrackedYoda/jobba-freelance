import Link from "next/link";

import { Logo } from "@/components/brand/logo";

const columns: [string, [string, string][]][] = [
  ["Marketplace", [["Find work", "/projects"], ["Find talent", "/freelancers"], ["Post a project", "/projects/new"], ["Connect wallet", "/connect"]]],
  ["Learn", [["Help center", "/help"], ["How escrow works", "/help/escrow"], ["Fees and payouts", "/help/fees"], ["Disputes", "/help/disputes"], ["FAQ", "/faq"]]],
  ["Company", [["Security", "/help/security"], ["Terms of service", "/terms"], ["Privacy policy", "/privacy"]]],
];

export function SiteFooter() {
  return (
    <footer className="siteFooter">
      <div className="shell footerGrid">
        <div className="footerBrand"><Logo onDark /><p>Freelance contracts settled in USDC escrow on Solana. Funds are held by the escrow program until release or dispute resolution.</p></div>
        {columns.map(([title, links]) => (
          <nav aria-label={title} key={title}><strong>{title}</strong>{links.map(([label, href]) => <Link key={href} href={href}>{label}</Link>)}</nav>
        ))}
      </div>
      <div className="shell footerBase"><span>© {new Date().getFullYear()} Jobba</span><span>Settlement in USDC · Network fees in SOL</span></div>
    </footer>
  );
}
