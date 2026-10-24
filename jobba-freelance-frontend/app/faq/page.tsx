import type { Metadata } from "next";
import Link from "next/link";

import { ContentPage } from "@/components/marketplace/content-page";
import { faq, getDoc } from "@/features/help/content";

export const metadata: Metadata = { title: "FAQ | Jobba" };

export default function FaqPage() {
  return (
    <ContentPage eyebrow="FAQ" title="Questions, answered plainly." lead="Short answers, with a link to the full guide for each topic.">
      <div className="narrowPage" style={{ marginInline: 0 }}>
        {faq.map((item) => (
          <details className="faq" key={item.q}>
            <summary>{item.q}</summary>
            <p>{item.a} <Link className="inlineLink" href={`/help/${item.doc}`}>{getDoc(item.doc)?.title} →</Link></p>
          </details>
        ))}
      </div>
    </ContentPage>
  );
}
