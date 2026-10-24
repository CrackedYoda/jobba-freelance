import "server-only";

import { getServerEnv } from "@/lib/config/env";
import { AppError } from "@/server/db/errors";

const LAMPORTS_PER_SOL = 1_000_000_000n;
const PRICE_TTL_MS = 60_000;
let cachedPrice: { cents: bigint; at: number } | undefined;

async function fetchJson(url: string): Promise<unknown> {
  const response = await fetch(url, { signal: AbortSignal.timeout(6_000), headers: { accept: "application/json" }, cache: "no-store" });
  if (!response.ok) throw new Error(`${url} responded ${response.status}`);
  return response.json();
}

function toCents(usd: unknown): bigint {
  const value = Number(usd);
  if (!Number.isFinite(value) || value <= 0) throw new Error("Invalid price");
  return BigInt(Math.round(value * 100));
}

/** SOL/USD in cents from public market data, cached for a minute. Two independent sources. */
export async function solUsdCents(): Promise<bigint> {
  if (cachedPrice && Date.now() - cachedPrice.at < PRICE_TTL_MS) return cachedPrice.cents;
  const sources: (() => Promise<bigint>)[] = [
    async () => toCents(((await fetchJson("https://api.coingecko.com/api/v3/simple/price?ids=solana&vs_currencies=usd")) as { solana?: { usd?: number } }).solana?.usd),
    async () => toCents(((await fetchJson("https://lite-api.jup.ag/price/v3?ids=So11111111111111111111111111111111111111112")) as Record<string, { usdPrice?: number }>).So11111111111111111111111111111111111111112?.usdPrice),
  ];
  for (const source of sources) {
    try {
      const cents = await source();
      cachedPrice = { cents, at: Date.now() };
      return cents;
    } catch { /* try the next source */ }
  }
  throw new AppError("PRICE_UNAVAILABLE", "We could not fetch the SOL price to verify your balance. Please try again in a minute.", 503);
}

export async function walletLamports(wallet: string): Promise<bigint> {
  const rpc = getServerEnv().SOLANA_RPC_URL;
  if (!rpc) throw new AppError("RPC_NOT_CONFIGURED", "Balance verification is not available right now.", 503);
  try {
    const response = await fetch(rpc, {
      method: "POST", headers: { "content-type": "application/json" }, signal: AbortSignal.timeout(8_000), cache: "no-store",
      body: JSON.stringify({ jsonrpc: "2.0", id: 1, method: "getBalance", params: [wallet, { commitment: "confirmed" }] }),
    });
    const body = await response.json() as { result?: { value?: number }; error?: unknown };
    if (typeof body.result?.value !== "number") throw new Error("No balance in RPC response");
    return BigInt(body.result.value);
  } catch {
    throw new AppError("BALANCE_UNAVAILABLE", "We could not read your wallet balance. Please try again.", 503);
  }
}

export function minimumSignupCents(): bigint {
  const usd = Number(process.env.SIGNUP_MIN_USD ?? "1");
  return BigInt(Math.round((Number.isFinite(usd) && usd >= 0 ? usd : 1) * 100));
}

export type BalanceCheck = { lamports: string; sol: string; usdCents: string; minimumUsdCents: string; priceUsdCents: string; sufficient: boolean; cluster: string };

export function assessSignupBalance(lamports: bigint, priceUsdCents: bigint, minimumUsdCents = minimumSignupCents()) {
  const usdCents = (lamports * priceUsdCents) / LAMPORTS_PER_SOL;
  return { usdCents, sufficient: usdCents >= minimumUsdCents };
}

export async function checkSignupBalance(wallet: string): Promise<BalanceCheck> {
  const [lamports, price] = await Promise.all([walletLamports(wallet), solUsdCents()]);
  const minimum = minimumSignupCents();
  const { usdCents, sufficient } = assessSignupBalance(lamports, price, minimum);
  const whole = lamports / LAMPORTS_PER_SOL;
  const fraction = (lamports % LAMPORTS_PER_SOL).toString().padStart(9, "0").slice(0, 4);
  return {
    lamports: lamports.toString(), sol: `${whole}.${fraction}`, usdCents: usdCents.toString(), minimumUsdCents: minimum.toString(),
    priceUsdCents: price.toString(), sufficient, cluster: process.env.NEXT_PUBLIC_SOLANA_CLUSTER ?? "devnet",
  };
}
