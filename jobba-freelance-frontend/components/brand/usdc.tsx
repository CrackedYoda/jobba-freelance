export function UsdcMark({ size = 16 }: { size?: number }) {
  return (
    <svg className="usdcMark" width={size} height={size} viewBox="0 0 32 32" role="img" aria-label="USDC" focusable="false">
      <circle cx="16" cy="16" r="16" fill="#2775CA" />
      <path d="M20.4 18.3c0-2.3-1.4-3.1-4.2-3.4-2-.3-2.4-.8-2.4-1.7s.7-1.5 2-1.5c1.2 0 1.9.4 2.2 1.4.1.2.3.3.5.3h1.1c.3 0 .5-.2.5-.5-.3-1.4-1.4-2.5-2.9-2.7V8.3c0-.3-.2-.5-.5-.5h-1c-.3 0-.5.2-.5.5v1.6c-2 .3-3.3 1.6-3.3 3.3 0 2.2 1.3 3.1 4.1 3.4 1.9.4 2.5.8 2.5 1.8s-.9 1.7-2.2 1.7c-1.7 0-2.3-.7-2.5-1.6-.1-.3-.3-.4-.5-.4h-1.2c-.3 0-.5.2-.5.5.3 1.6 1.3 2.7 3.5 3v1.6c0 .3.2.5.5.5h1c.3 0 .5-.2.5-.5v-1.6c2-.3 3.4-1.7 3.4-3.5z" fill="#fff" />
    </svg>
  );
}

/** "USDC" with the coin mark beside it. */
export function Usdc({ size = 15 }: { size?: number }) {
  return <span className="usdcLabel"><UsdcMark size={size} />USDC</span>;
}
