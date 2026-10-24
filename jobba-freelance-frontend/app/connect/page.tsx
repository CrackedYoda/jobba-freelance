import type { Metadata } from "next";
import Image from "next/image";
import { redirect } from "next/navigation";

import { Logo } from "@/components/brand/logo";
import { ConnectFlow } from "@/features/auth/connect-flow";
import { getViewer } from "@/server/marketplace/service";

export const metadata: Metadata = { title: "Sign in | Jobba" };

export default async function ConnectPage() {
  const viewer = await getViewer();
  if (viewer.authenticated && viewer.onboarded) redirect("/dashboard");
  return (
    <div className="connectPage">
      <aside className="connectRail">
        <Logo onDark size={36} />
        <div className="connectCopy">
          <h1>Your wallet is your account.</h1>
        </div>
        <Image className="connectArt" src="/illustrations/wallet.svg" alt="" width={320} height={220} unoptimized />
      </aside>
      <main className="connectMain"><ConnectFlow initiallyAuthenticated={viewer.authenticated} /></main>
    </div>
  );
}

export const dynamic = "force-dynamic";
