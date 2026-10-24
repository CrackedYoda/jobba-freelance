import type { Metadata } from "next";

import { ContentPage } from "@/components/marketplace/content-page";

export const metadata: Metadata = { title: "Privacy policy | Jobba" };

const sections: [string, string][] = [
  ["What we store", "Your public wallet address, the profile details you enter, projects, proposals, contract terms, messages, deliveries, dispute evidence, and the on-chain transactions linked to your contracts."],
  ["What we never store", "Private keys, seed phrases, or card and bank details. Jobba does not need them."],
  ["Sessions", "A secure, HttpOnly cookie keeps you signed in for up to 24 hours. Only a hash of the session token is stored on our servers."],
  ["Who can see what", "Profiles and open projects are public. Proposals are visible to you and the project client. Contract chat, deliveries and disputes are visible to the two parties and, for a dispute, the assigned administrator."],
  ["On-chain data", "Escrow transactions are recorded on the Solana blockchain and are public by nature. They cannot be deleted."],
  ["Your choices", "You can edit your profile at any time. To delete off-chain data, contact support from your signed-in wallet."],
];

export default function PrivacyPage() {
  return (
    <ContentPage eyebrow="Legal" title="Privacy policy" lead="What Jobba keeps, why, and who can see it.">
      <article className="docBody legal">{sections.map(([title, text]) => <section key={title}><h2>{title}</h2><p>{text}</p></section>)}</article>
    </ContentPage>
  );
}
