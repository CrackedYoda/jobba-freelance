import Image from "next/image";
import Link from "next/link";

import { ContentPage } from "@/components/marketplace/content-page";

export default function NotFound() {
  return (
    <ContentPage eyebrow="404" title="This page doesn't exist.">
      <div className="emptyState"><Image src="/illustrations/not-found.svg" alt="" width={240} height={170} unoptimized /><p>The link may be old, or the project may have been removed.</p><div className="pillRow" style={{ justifyContent: "center" }}><Link className="button primary" href="/projects">Browse projects</Link><Link className="button ghost" href="/help">Help center</Link></div></div>
    </ContentPage>
  );
}
