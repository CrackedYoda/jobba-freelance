import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { NamedIcon } from "@/components/brand/icons";
import { SiteFooter } from "@/components/marketplace/site-footer";
import { SiteHeader } from "@/components/marketplace/site-header";
import { docs, getDoc } from "@/features/help/content";

export function generateStaticParams() {
  return docs.map((doc) => ({ slug: doc.slug }));
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const doc = getDoc((await params).slug);
  return { title: doc ? `${doc.title} | Jobba help` : "Help | Jobba" };
}

export default async function DocPage({ params }: { params: Promise<{ slug: string }> }) {
  const doc = getDoc((await params).slug);
  if (!doc) notFound();
  return (
    <>
      <SiteHeader />
      <main className="shell pageSection">
        <nav className="crumbs" aria-label="Breadcrumb"><Link href="/help">Help center</Link> / <span>{doc.title}</span></nav>
        <div className="docLayout">
          <aside className="docNav" aria-label="Help topics">
            {docs.map((item) => <Link key={item.slug} href={`/help/${item.slug}`} className={item.slug === doc.slug ? "on" : ""} aria-current={item.slug === doc.slug ? "page" : undefined}><NamedIcon name={item.icon} size={16} />{item.title}</Link>)}
            <Link href="/faq">FAQ</Link>
          </aside>
          <article className="docBody">
            <span className="iconTile big"><NamedIcon name={doc.icon} size={26} /></span>
            <h1>{doc.title}</h1>
            <p className="lead">{doc.summary}</p>
            {doc.blocks.map((block, index) => {
              if (block.type === "p") return <p key={index}>{block.text}</p>;
              if (block.type === "h") return <h2 key={index}>{block.text}</h2>;
              if (block.type === "note") return <p className="docNote" key={index}>{block.text}</p>;
              if (block.type === "list") return <ul key={index}>{block.items.map((item) => <li key={item}>{item}</li>)}</ul>;
              return <ol className="docSteps" key={index}>{block.items.map(([title, text]) => <li key={title}><strong>{title}</strong><span>{text}</span></li>)}</ol>;
            })}
            <h2>Related</h2>
            <div className="pillRow">{doc.related.map((slug) => { const related = getDoc(slug); return related ? <Link className="pillButton" key={slug} href={`/help/${slug}`}>{related.title}</Link> : null; })}</div>
          </article>
        </div>
      </main>
      <SiteFooter />
    </>
  );
}
