export type DocBlock =
  | { type: "p"; text: string }
  | { type: "h"; text: string }
  | { type: "list"; items: string[] }
  | { type: "steps"; items: [string, string][] }
  | { type: "note"; text: string };

export type DocPage = { slug: string; title: string; icon: string; summary: string; blocks: DocBlock[]; related: string[] };

export const docs: DocPage[] = [
  {
    slug: "getting-started",
    title: "Getting started",
    icon: "rocket",
    summary: "Connect your wallet and choose how you will use Jobba.",
    related: ["wallets-and-sign-in", "for-clients", "for-freelancers"],
    blocks: [
      { type: "p", text: "Jobba is a decentralized freelance marketplace powered by protected USDC payments. Your wallet is your account." },
      { type: "steps", items: [
        ["Connect your wallet", "Choose Connect wallet to sign in or create your account."],
        ["Confirm your wallet", "Approve the sign-in request in your wallet to continue."],
        ["Pick your roles", "Choose hiring, freelancing, or both. You can change this from your profile at any time."],
        ["Start working", "Clients post a project; freelancers browse open projects and send proposals."],
      ] },
      { type: "note", text: "Keep a small amount of SOL in your wallet. Solana network fees are paid in SOL; contract amounts are always in USDC." },
    ],
  },
  {
    slug: "wallets-and-sign-in",
    title: "Wallet sign-in",
    icon: "wallet",
    summary: "Use your wallet to access Jobba.",
    related: ["getting-started", "security"],
    blocks: [
      { type: "p", text: "Your wallet is your account. Connect it and approve the sign-in request to access Jobba." },
      { type: "list", items: [
        "Signing in does not move funds.",
        "You stay in control of your wallet.",
        "Switching wallets signs you out of Jobba.",
      ] },
      { type: "note", text: "Jobba never asks for your seed phrase or private key. Anyone who does is not Jobba." },
    ],
  },
  {
    slug: "for-clients",
    title: "Hiring on Jobba",
    icon: "briefcase",
    summary: "Post a project, compare proposals, fund escrow and release payment.",
    related: ["escrow", "fees", "disputes"],
    blocks: [
      { type: "steps", items: [
        ["Post a project", "Describe the outcome, required skills, a fixed USDC budget and a delivery target. Save it as a draft or publish it."],
        ["Review proposals", "Each proposal shows the freelancer's price, delivery estimate and approach. Decline the ones that do not fit."],
        ["Select and fund", "Selecting a proposal creates the contract. One wallet signature moves the agreed USDC into the escrow vault."],
        ["Work together", "Once the freelancer accepts, use the contract chat to share files and decisions."],
        ["Release payment", "When the delivery is submitted, review it and release the escrow with one signature."],
      ] },
      { type: "h", text: "Changing your mind" },
      { type: "list", items: [
        "Before funding, you can choose a different proposal from the contract page.",
        "After funding but before the freelancer accepts, you can cancel and recover the full amount.",
        "After acceptance, unresolved problems go through a dispute.",
      ] },
    ],
  },
  {
    slug: "for-freelancers",
    title: "Freelancing on Jobba",
    icon: "user-check",
    summary: "Build a profile, propose, accept funded work and get paid in USDC.",
    related: ["escrow", "fees", "disputes"],
    blocks: [
      { type: "steps", items: [
        ["Complete your profile", "Add a headline, bio, skills, an hourly guide and portfolio links. Clients see this next to every proposal."],
        ["Send proposals", "Set your price and delivery estimate. You see your payout after the platform fee before you submit."],
        ["Accept funded contracts", "You only accept after the client has funded the escrow, so the money is visible before you start."],
        ["Deliver", "Record your delivery with links and a description, then submit it on-chain from the contract page."],
        ["Get paid", "When the client releases, USDC arrives in your wallet minus the platform fee."],
      ] },
      { type: "note", text: "You can withdraw a pending proposal at any time and send a new one while the project is open." },
    ],
  },
  {
    slug: "escrow",
    title: "How escrow works",
    icon: "lock",
    summary: "The states a contract moves through and who can move it.",
    related: ["fees", "disputes", "security"],
    blocks: [
      { type: "p", text: "Every contract has its own escrow account controlled by the Jobba escrow program on Solana. The program, not Jobba and not either party, holds the USDC. It only moves funds when the right person signs a valid instruction for the current state." },
      { type: "steps", items: [
        ["Funded", "The client signed the funding transaction. The full amount sits in the contract's vault."],
        ["Active", "The freelancer accepted the funded contract. Work begins."],
        ["Submitted", "The freelancer submitted the delivery for review."],
        ["Released", "The client approved the work. The freelancer is paid and the vault is closed."],
      ] },
      { type: "h", text: "Other outcomes" },
      { type: "list", items: [
        "Cancelled: the client recovers the full amount before the freelancer accepts.",
        "Disputed: either party opened a dispute. Normal release is locked.",
        "Refunded or awarded: an administrator resolved the dispute in full to one side.",
      ] },
      { type: "p", text: "Every transaction has an Explorer link in the contract workspace so both parties can verify it independently." },
    ],
  },
  {
    slug: "fees",
    title: "Fees and payouts",
    icon: "receipt",
    summary: "What clients pay, what freelancers receive, and network fees.",
    related: ["escrow", "for-freelancers"],
    blocks: [
      { type: "list", items: [
        "Clients fund exactly the agreed price. No fee is added on top.",
        "The platform fee is a percentage of the contract amount, deducted from the freelancer's payout on release.",
        "If a contract is cancelled before acceptance or refunded after a dispute, the client receives the full amount back.",
        "Solana network fees are paid in SOL by whoever signs the transaction, typically a fraction of a cent.",
      ] },
      { type: "note", text: "The proposal composer and contract terms show the exact fee and payout before anyone signs." },
    ],
  },
  {
    slug: "disputes",
    title: "Disputes",
    icon: "scale",
    summary: "What happens when a contract goes wrong.",
    related: ["escrow", "for-clients", "for-freelancers"],
    blocks: [
      { type: "steps", items: [
        ["Open a dispute", "Either party can open a dispute from the contract page while the escrow is funded, active or submitted."],
        ["Release locks", "Opening the dispute is recorded on-chain. Normal release and cancellation are blocked until it is resolved."],
        ["Add evidence", "Both sides can add a statement and links. The administrator also sees the terms, chat and deliveries."],
        ["Decision", "An administrator refunds the client or awards the freelancer in full, with a written reason."],
      ] },
      { type: "p", text: "The decision, the reason and the settlement transaction are recorded permanently in the case history." },
    ],
  },
  {
    slug: "security",
    title: "Security",
    icon: "shield-check",
    summary: "How Jobba protects your account and your funds.",
    related: ["wallets-and-sign-in", "escrow"],
    blocks: [
      { type: "list", items: [
        "Your wallet is your account, and Jobba never asks for your private key or recovery phrase.",
        "Contract amounts and recipients are checked before a contract is updated.",
        "Administrative decisions require a written reason and remain in the case history.",
      ] },
      { type: "note", text: "Always check the amount and recipient in your wallet before approving a transaction." },
    ],
  },
];

export const faq: { q: string; a: string; doc: string }[] = [
  { q: "Is my payment protected?", a: "Yes. The agreed USDC is locked in the contract's escrow vault before work starts. Funds only move when the client releases, the client cancels before acceptance, or an administrator resolves a dispute.", doc: "escrow" },
  { q: "Who holds the funds?", a: "The Jobba escrow program on Solana. Each contract has its own vault owned by the program, not by Jobba or either party.", doc: "escrow" },
  { q: "Do I need SOL?", a: "Yes, a small amount. SOL pays Solana network fees, usually a fraction of a cent per transaction. Contract amounts are always in USDC.", doc: "fees" },
  { q: "What does Jobba charge?", a: "Clients pay exactly the agreed price. A platform fee is deducted from the freelancer's payout on release, and the composer shows the exact figure before you send a proposal.", doc: "fees" },
  { q: "Can I cancel?", a: "Before funding, choose a different proposal. After funding, the client can cancel and recover the full amount until the freelancer accepts.", doc: "for-clients" },
  { q: "What happens if we disagree?", a: "Either party opens a dispute with evidence. Normal release locks, an administrator reviews the terms, chat and deliverables, and refunds the client or awards the freelancer in full.", doc: "disputes" },
  { q: "Which wallets work?", a: "Choose from the available wallets shown when you select Connect wallet.", doc: "wallets-and-sign-in" },
  { q: "Does signing in cost anything?", a: "No. Signing in does not move funds or charge a fee.", doc: "wallets-and-sign-in" },
  { q: "Can I hire and freelance with one wallet?", a: "Yes. Turn on both roles in your profile and switch between the client and freelancer workspaces.", doc: "getting-started" },
];

export function getDoc(slug: string): DocPage | undefined {
  return docs.find((doc) => doc.slug === slug);
}
