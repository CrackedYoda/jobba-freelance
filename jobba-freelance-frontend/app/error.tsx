"use client";

import Image from "next/image";
import Link from "next/link";

export default function ErrorPage({ reset }: { error: Error; reset: () => void }) {
  return (
    <main className="shell pageSection">
      <div className="emptyState"><Image src="/illustrations/not-found.svg" alt="" width={240} height={170} unoptimized /><h1>Something went wrong.</h1><p>Please try again. If it keeps happening, check the help center.</p>
        <div className="pillRow" style={{ justifyContent: "center" }}><button className="button primary" type="button" onClick={reset}>Try again</button><Link className="button ghost" href="/help">Help center</Link></div></div>
    </main>
  );
}
