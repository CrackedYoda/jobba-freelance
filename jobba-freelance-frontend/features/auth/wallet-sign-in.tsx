"use client";

import { useAppKit, useAppKitAccount, useAppKitProvider } from "@reown/appkit/react";
import type { Provider } from "@reown/appkit-adapter-solana/react";
import { useEffect, useRef, useState } from "react";

import type { ChallengeResponse, SessionResponse } from "@/features/auth/schemas";
import { isWalletConfigured } from "@/providers/appkit-provider";

type Status = "idle" | "requesting" | "signing" | "verifying" | "authenticated" | "error";

function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((byte) => { binary += String.fromCharCode(byte); });
  return window.btoa(binary);
}

function ConnectedWalletSignIn() {
  const { open } = useAppKit();
  const { address, isConnected } = useAppKitAccount({ namespace: "solana" });
  const { walletProvider } = useAppKitProvider<Provider>("solana");
  const [session, setSession] = useState<SessionResponse>({ authenticated: false });
  const [status, setStatus] = useState<Status>("idle");
  const [error, setError] = useState<string>();
  const authenticatedWallet = useRef<string | undefined>(undefined);

  useEffect(() => {
    let active = true;
    void fetch("/api/v1/auth/session", { cache: "no-store" })
      .then((response) => response.json() as Promise<SessionResponse>)
      .then((next) => {
        if (!active) return;
        setSession(next);
        authenticatedWallet.current = next.walletAddress;
        setStatus(next.authenticated ? "authenticated" : "idle");
      });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    if (authenticatedWallet.current && address !== authenticatedWallet.current) {
      void fetch("/api/v1/auth/logout", { method: "POST" }).finally(() => {
        setSession({ authenticated: false });
        authenticatedWallet.current = undefined;
        setStatus("idle");
      });
    }
  }, [address]);

  async function signIn() {
    if (!address || !walletProvider) return;
    setError(undefined);
    try {
      setStatus("requesting");
      const challengeResponse = await fetch("/api/v1/auth/challenge", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ address }),
      });
      if (!challengeResponse.ok) throw new Error("Could not create a sign-in challenge.");
      const challenge = await challengeResponse.json() as ChallengeResponse;

      setStatus("signing");
      const signature = await walletProvider.signMessage(new TextEncoder().encode(challenge.message));

      setStatus("verifying");
      const verifyResponse = await fetch("/api/v1/auth/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ address, nonce: challenge.nonce, signature: bytesToBase64(signature) }),
      });
      if (!verifyResponse.ok) throw new Error("The wallet signature was rejected.");
      const nextSession = await verifyResponse.json() as SessionResponse;
      setSession(nextSession);
      authenticatedWallet.current = nextSession.walletAddress;
      setStatus("authenticated");
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Wallet sign-in failed.");
      setStatus("error");
    }
  }

  async function signOut() {
    await fetch("/api/v1/auth/logout", { method: "POST" });
    authenticatedWallet.current = undefined;
    setSession({ authenticated: false });
    setStatus("idle");
  }

  if (!isConnected) {
    return <button type="button" onClick={() => void open({ view: "Connect" })}>Connect wallet</button>;
  }

  if (session.authenticated && session.walletAddress && session.walletAddress === address) {
    const signedInAddress = session.walletAddress;
    return (
      <div>
        <span>Signed in as {`${signedInAddress.slice(0, 4)}…${signedInAddress.slice(-4)}`}</span>{" "}
        <button type="button" onClick={() => void signOut()}>Sign out</button>
      </div>
    );
  }

  return (
    <div>
      <button type="button" onClick={() => void signIn()} disabled={!walletProvider || ["requesting", "signing", "verifying"].includes(status)}>
        {status === "signing" ? "Check your wallet…" : status === "verifying" ? "Verifying…" : "Sign in with Solana"}
      </button>
      {error ? <p role="alert">{error}</p> : null}
      <p>Signing is free and does not submit a transaction.</p>
    </div>
  );
}

export function WalletSignIn() {
  if (!isWalletConfigured) {
    return <p role="status">Wallet connection is unavailable until NEXT_PUBLIC_REOWN_PROJECT_ID is configured.</p>;
  }
  return <ConnectedWalletSignIn />;
}
