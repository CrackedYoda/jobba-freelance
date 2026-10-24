import {
  BadgeCheck, Briefcase, Code2, Database, Film, Lock, Megaphone, Palette, PenLine, Receipt, Rocket, Scale, ShieldCheck, UserCheck, Wallet,
  type LucideIcon,
} from "lucide-react";

const icons: Record<string, LucideIcon> = {
  rocket: Rocket, wallet: Wallet, briefcase: Briefcase, "user-check": UserCheck, lock: Lock, receipt: Receipt, scale: Scale, "shield-check": ShieldCheck,
  Development: Code2, "Design & Creative": Palette, Writing: PenLine, Marketing: Megaphone, "Data & AI": Database, "Video & animation": Film,
};

export function NamedIcon({ name, size = 20 }: { name: string; size?: number }) {
  const Icon = icons[name] ?? BadgeCheck;
  return <Icon size={size} strokeWidth={1.8} aria-hidden="true" />;
}
