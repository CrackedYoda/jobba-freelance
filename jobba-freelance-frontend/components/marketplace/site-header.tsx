import Link from "next/link";

export function SiteHeader() {
  return (
    <header className="marketHeader">
      <nav className="shell marketNav" aria-label="Marketplace navigation">
        <Link className="brand" href="/">Jobba<span>.</span></Link>
        <div className="marketNavLinks">
          <Link href="/projects">Find work</Link>
          <Link href="/freelancers/amara-okafor">Find talent</Link>
          <Link href="/dashboard">Dashboard</Link>
          <Link href="/admin/disputes">Disputes</Link>
        </div>
        <Link className="button secondary compactButton" href="/projects/new">Post a project</Link>
      </nav>
    </header>
  );
}
