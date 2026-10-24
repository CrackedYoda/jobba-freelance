import Link from "next/link";

import { Logo } from "@/components/brand/logo";
import { MobileMenu, type NavItem } from "@/components/marketplace/mobile-menu";
import { SignOutButton } from "@/features/auth/wallet-sign-in";
import { getViewer, shortKey } from "@/server/marketplace/service";

function initials(name: string | undefined, fallback: string): string {
  const source = (name ?? fallback).replace(/^Wallet\s+/, "");
  return source.split(/\s+/).map((part) => part[0]).join("").slice(0, 2).toUpperCase();
}

export async function SiteHeader({ active }: { active?: "work" | "talent" | "dashboard" | "contracts" | "admin" | "help" }) {
  const viewer = await getViewer();
  const canPost = !viewer.authenticated || viewer.roles.includes("client");
  const items: NavItem[] = [
    { key: "work", href: "/projects", label: "Find work" },
    { key: "talent", href: "/freelancers", label: "Find talent" },
    ...(viewer.authenticated ? [{ key: "dashboard", href: "/dashboard", label: "Dashboard" }, { key: "contracts", href: "/contracts", label: "Contracts" }] : []),
    ...(viewer.roles.includes("admin") ? [{ key: "admin", href: "/admin/disputes", label: "Disputes" }] : []),
    { key: "help", href: "/help", label: "Help" },
    ...(canPost ? [{ key: "post", href: "/projects/new", label: "Post a project" }] : []),
  ];
  const desktopLinks = items.filter((item) => item.key !== "post");
  return (
    <header className="marketHeader">
      <nav className="shell marketNav" aria-label="Primary">
        <Logo />
        <div className="marketNavLinks">
          {desktopLinks.map((item) => <Link key={item.key} href={item.href} aria-current={active === item.key ? "page" : undefined} className={active === item.key ? "on" : undefined}>{item.label}</Link>)}
        </div>
        <div className="headerActions">
          {canPost ? <Link className="button ghost compactButton" href="/projects/new">Post a project</Link> : null}
          {viewer.authenticated && viewer.walletAddress ? (
            <>
              <Link className="walletPill" href="/dashboard" aria-label="Open dashboard"><i aria-hidden="true" />{shortKey(viewer.walletAddress)}<b>{initials(viewer.displayName, viewer.walletAddress)}</b></Link>
              <SignOutButton />
            </>
          ) : <Link className="button primary compactButton" href="/connect">Connect wallet</Link>}
          <MobileMenu items={items} active={active} logo={<Logo />} footer={viewer.authenticated && viewer.walletAddress ? (
            <>
              <Link className="walletPill" href="/dashboard"><i aria-hidden="true" />{shortKey(viewer.walletAddress)}<b>{initials(viewer.displayName, viewer.walletAddress)}</b></Link>
              <SignOutButton />
            </>
          ) : <Link className="button primary big" href="/connect">Connect wallet</Link>} />
        </div>
      </nav>
    </header>
  );
}
