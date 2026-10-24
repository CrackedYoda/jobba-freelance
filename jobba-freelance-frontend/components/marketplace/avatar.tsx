/* eslint-disable @next/next/no-img-element -- user-supplied remote URLs are not known to next/image config */
export function Avatar({ name, url, size = "large" }: { name: string; url: string | null; size?: "large" | "small" }) {
  const initials = name.split(/\s+/).map((part) => part[0]).join("").slice(0, 2).toUpperCase();
  const className = size === "small" ? "avatar small" : "avatar";
  if (url) return <img className={`${className} photo`} src={url} alt="" referrerPolicy="no-referrer" />;
  return <div className={className} aria-hidden="true">{initials}</div>;
}
