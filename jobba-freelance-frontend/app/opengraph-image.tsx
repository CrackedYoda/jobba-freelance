import { ImageResponse } from "next/og";

export const alt = "Jobba — hire anywhere, pay on delivery";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div style={{ width: "100%", height: "100%", background: "#0E1512", color: "#fff", display: "flex", flexDirection: "column", justifyContent: "space-between", padding: 72 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 20 }}>
          <svg width="72" height="72" viewBox="0 0 40 40"><rect width="40" height="40" rx="11" fill="#C6F24E" /><path d="M26 9.5v13a7.5 7.5 0 0 1-15 0" fill="none" stroke="#0E1512" strokeWidth="4.6" strokeLinecap="round" /><circle cx="14.5" cy="13.5" r="3.6" fill="#0E1512" /></svg>
          <div style={{ fontSize: 56, fontWeight: 800 }}>Jobba.</div>
        </div>
        <div style={{ display: "flex", flexDirection: "column", fontSize: 92, fontWeight: 800, lineHeight: 1 }}>
          <div>Hire anywhere.</div>
          <div style={{ color: "#C6F24E" }}>Pay on delivery.</div>
        </div>
        <div style={{ fontSize: 30, color: "#B8C3BE" }}>Freelance contracts settled in USDC escrow on Solana.</div>
      </div>
    ),
    size,
  );
}
