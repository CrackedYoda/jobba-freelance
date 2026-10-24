import Link from "next/link";
import { notFound } from "next/navigation";
import { SiteHeader } from "@/components/marketplace/site-header";
import { getFreelancer } from "@/server/marketplace/service";

export default async function FreelancerPage({ params }: { params: Promise<{ profileId: string }> }) {
  const profile = await getFreelancer((await params).profileId);
  if (!profile) notFound();
  return <><SiteHeader /><main className="shell profilePage pageSection"><section className="profileHero"><div className="avatar" aria-hidden="true">{profile.name.split(" ").map((part) => part[0]).join("")}</div><div><p className="eyebrow">Available talent</p><h1>{profile.name}</h1><p className="lead">{profile.headline}</p><div className="profileMeta"><span>{profile.location}</span><span>{profile.availability}</span></div></div><Link className="button primary" href="/projects/new">Discuss a project</Link></section><div className="profileGrid"><article><section className="contentSection"><h2>About</h2><p>{profile.bio}</p></section><section className="contentSection"><h2>Expertise</h2><div className="tagRow">{profile.skills.map((skill) => <span key={skill}>{skill}</span>)}</div></section><section className="contentSection"><h2>Selected work</h2><div className="portfolioGrid">{profile.portfolio.map((item) => <a href={item.url} key={item.label}><strong>{item.label}</strong><i>View project ↗</i></a>)}</div></section></article><aside className="profileStats"><div><span>Hourly rate</span><strong>{profile.rateUsdc} USDC</strong></div><div><span>Completed contracts</span><strong>{profile.completedContracts}</strong></div><div><span>Lifetime earnings</span><strong>{profile.earnedUsdc} USDC</strong></div></aside></div></main></>;
}
