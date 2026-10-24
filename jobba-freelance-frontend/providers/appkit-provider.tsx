"use client";

import { createAppKit } from "@reown/appkit/react";
import { SolanaAdapter } from "@reown/appkit-adapter-solana/react";
import type { ReactNode } from "react";

import { publicEnv } from "@/lib/config/public-env";
import { appKitNetwork } from "@/lib/solana/network";

export const isWalletConfigured = Boolean(publicEnv.reownProjectId);

if (publicEnv.reownProjectId) {
  const appKit = createAppKit({
    adapters: [new SolanaAdapter()],
    networks: [appKitNetwork],
    defaultNetwork: appKitNetwork,
    projectId: publicEnv.reownProjectId,
    metadata: {
      name: "Jobba Freelance",
      description: "Freelance work settled through USDC escrow on Solana.",
      url: publicEnv.appUrl,
      icons: [`${publicEnv.appUrl}/icon.svg`],
    },
    features: {
      analytics: false,
      email: false,
      socials: [],
    },
    themeMode: "light",
    // Match the Jobba palette: ink accent, warm paper surfaces, Geist type, rounded corners.
    themeVariables: {
      "--w3m-accent": "#0E1512",
      "--w3m-color-mix": "#F4F3EC",
      "--w3m-color-mix-strength": 40,
      "--w3m-font-family": "\"Geist Variable\", system-ui, sans-serif",
      "--w3m-border-radius-master": "4px",
      "--w3m-qr-color": "#0E1512",
      "--w3m-z-index": 1000,
    },
  });

  appKit.subscribeRemoteFeatures((features) => {
    if (features?.reownBranding) {
      appKit.updateRemoteFeatures({ ...features, reownBranding: false });
    }
  });
}

export function AppKitProvider({ children }: { children: ReactNode }) {
  return children;
}
