import type { ReactNode } from "react";

import { SiteFooter } from "@/components/marketplace/site-footer";
import { SiteHeader } from "@/components/marketplace/site-header";

export function ContentPage({ eyebrow, title, lead, children }: { eyebrow: string; title: ReactNode; lead?: string; children: ReactNode }) {
  return (
    <>
      <SiteHeader />
      <main className="shell pageSection">
        <div className="pageIntro"><p className="eyebrow">{eyebrow}</p><h1>{title}</h1>{lead ? <p className="lead">{lead}</p> : null}</div>
        {children}
      </main>
      <SiteFooter />
    </>
  );
}
