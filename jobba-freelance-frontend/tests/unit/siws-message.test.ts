import { describe, expect, it } from "vitest";

import { buildSignInMessage, createNonce, siwsChainId } from "../../server/auth/message";

describe("SIWS message", () => {
  it("binds the wallet, origin, chain, nonce, and expiry in exact order", () => {
    const message = buildSignInMessage({
      domain: "localhost:3000",
      address: "11111111111111111111111111111111",
      statement: "Sign in to Jobba Freelance.",
      uri: "http://localhost:3000",
      chainId: "solana:devnet",
      nonce: "abcdefghijklmnop",
      issuedAt: "2026-10-23T12:00:00.000Z",
      expirationTime: "2026-10-23T12:05:00.000Z",
    });

    expect(message).toContain("localhost:3000 wants you to sign in with your Solana account:");
    expect(message).toContain("Chain ID: solana:devnet");
    expect(message).toContain("Nonce: abcdefghijklmnop");
    expect(message.endsWith("Expiration Time: 2026-10-23T12:05:00.000Z")).toBe(true);
  });

  it("uses alphanumeric nonces that wallets accept", () => {
    for (let i = 0; i < 500; i += 1) expect(createNonce()).toMatch(/^[A-Za-z0-9]{24}$/);
  });

  it("maps cluster names to SIWS chain IDs", () => {
    expect(siwsChainId("devnet")).toBe("solana:devnet");
    expect(siwsChainId("testnet")).toBe("solana:testnet");
    expect(siwsChainId("mainnet-beta")).toBe("solana:mainnet");
  });
});
