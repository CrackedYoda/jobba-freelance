import type { Metadata } from "next";
import Link from "next/link";

import { NamedIcon } from "@/components/brand/icons";
import { ContentPage } from "@/components/marketplace/content-page";
import { docs } from "@/features/help/content";

export const metadata: Metadata = { title: "Help center | Jobba" };

export default function HelpPage() {
  return (
    <ContentPage eyebrow="Help center" title={<>Everything you need to <em>get paid.</em></>} lead="Guides for clients and freelancers, from your first sign-in to dispute resolution.">
      <div className="docGrid">
        {docs.map((doc) => (
          <Link className="docCard" href={`/help/${doc.slug}`} key={doc.slug}>
            <span className="iconTile"><NamedIcon name={doc.icon} /></span>
            <strong>{doc.title}</strong><span>{doc.summary}</span>
          </Link>
        ))}
      </div>
      <div className="helpCta"><div><strong>Still have a question?</strong><p>Common questions are answered on the FAQ page.</p></div><Link className="button primary" href="/faq">Read the FAQ</Link></div>
    </ContentPage>
  );
}
