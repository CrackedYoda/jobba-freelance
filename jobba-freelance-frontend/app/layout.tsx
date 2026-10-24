import type { Metadata } from "next";
import { AppKitProvider } from "@/providers/appkit-provider";
import "./styles.css";

export const metadata: Metadata = {
  title: "Jobba Freelance",
  description: "Hire and get paid through USDC escrow on Solana.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body><AppKitProvider>{children}</AppKitProvider></body>
    </html>
  );
}
