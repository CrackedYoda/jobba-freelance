"use client";

import Link from "next/link";

import { useWalletAuth } from "@/features/auth/use-wallet-auth";
import { OnboardingWizard } from "@/features/onboarding/onboarding-wizard";
import { isWalletConfigured } from "@/providers/appkit-provider";

const steps = [
  ["Connect wallet", "Your wallet is your account"],
  ["Confirm wallet", "Approve in your wallet"],
  ["Set up profile", "Choose how you use Jobba"],
];

function Rail({ current }: { current: number }) {
  return (
    <ol className="connectSteps" aria-label="Sign-in progress">
      {steps.map(([title, sub], index) => (
        <li key={title} aria-current={index === current ? "step" : undefined} className={index < current ? "done" : index === current ? "now" : ""}>
          <span className="mono">{index < current ? "✓" : index + 1}</span>
          <div><strong>{title}</strong><small>{sub}</small></div>
        </li>
      ))}
    </ol>
  );
}

function Flow({ initiallyAuthenticated }: { initiallyAuthenticated: boolean }) {
  const auth = useWalletAuth();
  const busy = ["requesting", "signing", "verifying"].includes(auth.status);
  const authed = auth.matched || (initiallyAuthenticated && auth.session.authenticated);
  const step = authed ? 2 : auth.isConnected ? 1 : 0;

  return (
    <div className="connectPanel">
      <Rail current={step} />
      {step === 0 ? (
        <section aria-labelledby="c0">
          <span className="eyebrow">Step 1 of 3</span>
          <h2 id="c0">Connect your wallet</h2>
          <p className="lead">Use one wallet to hire, freelance, or do both.</p>
          <button className="button primary wide" type="button" onClick={auth.connect}>Connect wallet</button>
          <p className="fine">By connecting you agree to the <Link className="inlineLink" href="/terms">terms of service</Link> and <Link className="inlineLink" href="/privacy">privacy policy</Link>. New to wallets? <Link className="inlineLink" href="/help/wallets-and-sign-in">Read the guide</Link>.</p>
        </section>
      ) : null}
      {step === 1 ? (
        <section aria-labelledby="c1">
          <span className="eyebrow">Step 2 of 3</span>
          <h2 id="c1">Confirm your wallet</h2>
          <p className="lead">Approve the sign-in request in your wallet to continue.</p>
          <div className="messageCard">
            <div><span className="mono">{auth.address}</span><span className="chip open mono">SOLANA</span></div>
            <small>Connected wallet</small>
          </div>
          {busy ? <p className="formNotice" role="status">{auth.status === "signing" ? "Approve the request in your wallet to continue." : "Working…"}</p> : null}
          {auth.error ? <p className="formNotice error" role="alert">{auth.error}</p> : null}
          <div className="formActions row">
            <button className="button ghost" type="button" onClick={auth.connect}>Change wallet</button>
            <button className="button primary grow" type="button" disabled={!auth.ready || busy} onClick={() => void auth.signIn()}>{busy ? "Check your wallet…" : "Continue"}</button>
          </div>
        </section>
      ) : null}
      {step === 2 ? (
        <section aria-labelledby="c2">
          <span className="eyebrow">Step 3 of 3</span>
          <h2 id="c2">How will you use Jobba?</h2>
          <p className="lead">Choose your role and complete your profile. Your progress is saved automatically.</p>
          <OnboardingWizard />
          <p className="fine"><button className="textButton" type="button" onClick={() => void auth.signOut()}>Sign out</button></p>
        </section>
      ) : null}
    </div>
  );
}

export function ConnectFlow({ initiallyAuthenticated }: { initiallyAuthenticated: boolean }) {
  if (!isWalletConfigured) {
    return <div className="connectPanel"><h2>Wallet unavailable</h2><p className="formNotice error" role="alert">Wallet sign-in is temporarily unavailable. Please try again later.</p></div>;
  }
  return <Flow initiallyAuthenticated={initiallyAuthenticated} />;
}
