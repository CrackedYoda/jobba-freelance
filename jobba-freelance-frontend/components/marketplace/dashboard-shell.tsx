import Link from "next/link";
import type { ReactNode } from "react";

const links = [["Overview", "/dashboard"], ["My projects", "/projects"], ["Proposals", "/dashboard/proposals"], ["Profile", "/onboarding"]] as const;

export function DashboardShell({ children, active = "Overview" }: { children: ReactNode; active?: string }) {
  return (
    <div className="dashboardLayout shell">
      <aside className="dashboardRail">
        <p className="eyebrow">Workspace</p>
        <nav aria-label="Dashboard navigation">{links.map(([label, href]) => <Link className={active === label ? "active" : ""} href={href} key={label}>{label}</Link>)}</nav>
        <div className="networkNote"><i aria-hidden="true" /> Solana devnet</div>
      </aside>
      <div className="dashboardMain">{children}</div>
    </div>
  );
}
