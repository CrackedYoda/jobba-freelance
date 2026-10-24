import { WalletSignIn } from "@/features/auth/wallet-sign-in";

const steps = [
  ["Post", "Describe the outcome, skills, budget, and delivery date."],
  ["Choose", "Compare focused proposals and select a freelancer."],
  ["Fund", "Lock the agreed USDC amount in a Solana escrow."],
  ["Deliver", "Chat, submit work, and release payment on completion."],
];

export default function Home() {
  return (
    <main>
      <nav className="nav shell" aria-label="Primary navigation">
        <a className="brand" href="#top">Jobba<span>.</span></a>
        <div className="navLinks">
          <a href="#work">Find work</a>
          <a href="#talent">Find talent</a>
          <div className="walletAction"><WalletSignIn /></div>
        </div>
      </nav>

      <section className="hero shell" id="top">
        <p className="eyebrow">Freelance without borders</p>
        <h1>Great work deserves <em>clear terms</em> and instant settlement.</h1>
        <p className="lead">Hire verified independent talent, protect every contract with USDC escrow, and keep delivery, chat, and disputes in one workspace.</p>
        <div className="actions">
          <button className="button primary" type="button">Post a project</button>
          <button className="button ghost" type="button">Browse projects</button>
        </div>
        <div className="trustRow" aria-label="Product benefits">
          <span>USDC escrow</span><span>Solana speed</span><span>Admin dispute support</span>
        </div>
      </section>

      <section className="process shell" id="work">
        <div>
          <p className="eyebrow">A complete contract loop</p>
          <h2>From brief to paid, without payment guesswork.</h2>
        </div>
        <div className="stepGrid">
          {steps.map(([title, description], index) => (
            <article className="step" key={title}>
              <span>0{index + 1}</span>
              <h3>{title}</h3>
              <p>{description}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="escrow shell" id="talent">
        <div>
          <p className="eyebrow">Built for both sides</p>
          <h2>Funds stay protected until the work moves forward.</h2>
          <p>Clients fund an escrow after choosing a proposal. Freelancers accept the funded contract before work begins. If the result is disputed, an administrator reviews the shared record and resolves the vault.</p>
        </div>
        <aside className="contractCard">
          <div className="cardHead"><span>Website redesign</span><strong>Funded</strong></div>
          <p className="amount">2,400.00 <small>USDC</small></p>
          <div className="meter"><i /></div>
          <div className="cardMeta"><span>Client funded</span><span>Awaiting acceptance</span></div>
        </aside>
      </section>
    </main>
  );
}
