import { randomBytes } from "node:crypto";

export type SignInChallenge = {
  domain: string;
  address: string;
  statement: string;
  uri: string;
  chainId: string;
  nonce: string;
  issuedAt: string;
  expirationTime: string;
};

/** SIWS chain IDs: wallets accept "solana:mainnet", not the RPC cluster name "mainnet-beta". */
export function siwsChainId(cluster: string): string {
  return cluster === "mainnet-beta" ? "solana:mainnet" : `solana:${cluster}`;
}

export function buildSignInMessage(challenge: SignInChallenge): string {
  return [
    `${challenge.domain} wants you to sign in with your Solana account:`,
    challenge.address,
    "",
    challenge.statement,
    "",
    `URI: ${challenge.uri}`,
    "Version: 1",
    `Chain ID: ${challenge.chainId}`,
    `Nonce: ${challenge.nonce}`,
    `Issued At: ${challenge.issuedAt}`,
    `Expiration Time: ${challenge.expirationTime}`,
  ].join("\n");
}

const NONCE_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

/** Sign-In With Solana requires an alphanumeric nonce; wallets such as Phantom reject "-" or "_". */
export function createNonce(length = 24): string {
  let nonce = "";
  while (nonce.length < length) {
    for (const byte of randomBytes(length)) {
      if (byte < 248 && nonce.length < length) nonce += NONCE_ALPHABET[byte % 62];
    }
  }
  return nonce;
}
