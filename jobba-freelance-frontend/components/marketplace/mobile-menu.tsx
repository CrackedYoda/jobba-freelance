"use client";

import { Menu, X } from "lucide-react";
import Link from "next/link";
import { useEffect, useId, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

export type NavItem = { key: string; href: string; label: string };

/** Full-screen mobile navigation: links on top, account actions pinned to the bottom. */
export function MobileMenu({ items, active, logo, footer }: { items: NavItem[]; active?: string; logo: ReactNode; footer: ReactNode }) {
  const [open, setOpen] = useState(false);
  const panelId = useId();

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => { if (event.key === "Escape") setOpen(false); };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => { document.removeEventListener("keydown", onKey); document.body.style.overflow = ""; };
  }, [open]);

  return (
    <div className="mobileMenu">
      <button type="button" className="menuToggle" aria-expanded={open} aria-controls={panelId} aria-label="Open menu" onClick={() => setOpen(true)}>
        <Menu size={22} aria-hidden="true" />
      </button>
      {open ? createPortal(
        <div id={panelId} className="menuOverlay" role="dialog" aria-modal="true" aria-label="Menu">
          <div className="menuTop">
            {logo}
            <button type="button" className="menuToggle" aria-label="Close menu" onClick={() => setOpen(false)}><X size={22} aria-hidden="true" /></button>
          </div>
          <nav className="menuLinks" aria-label="Menu" onClick={() => setOpen(false)}>
            {items.map((item) => <Link key={item.key} href={item.href} className={active === item.key ? "on" : undefined} aria-current={active === item.key ? "page" : undefined}>{item.label}</Link>)}
          </nav>
          <div className="menuFooter" onClick={() => setOpen(false)}>{footer}</div>
        </div>,
        document.body,
      ) : null}
    </div>
  );
}
