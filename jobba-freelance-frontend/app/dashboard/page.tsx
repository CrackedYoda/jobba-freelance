import type { Metadata } from "next";
import Link from "next/link";

import { DashboardShell } from "@/components/marketplace/dashboard-shell";
import { SiteHeader } from "@/components/marketplace/site-header";

export const metadata: Metadata = { title: "Choose workspace | Jobba" };

export default function DashboardPage() {
  return <><SiteHeader /><main><DashboardShell><header className="dashboardHeader"><div><p className="eyebrow">Your workspace</p><h1>What are you here to <em>do?</em></h1><p className="lead">Switch between hiring and earning without changing wallets.</p></div></header><div className="roleGrid"><Link href="/dashboard/client"><span>Hire talent</span><h2>Client workspace</h2><p>Post projects, review proposals, and track funded contracts.</p><b>Open client dashboard →</b></Link><Link href="/dashboard/freelancer"><span>Find work</span><h2>Freelancer workspace</h2><p>Track proposals, active work, submissions, and payments.</p><b>Open freelancer dashboard →</b></Link></div></DashboardShell></main></>;
}
