import Link from "next/link";
import { CreditCard, KeyRound, ShieldCheck, Zap } from "lucide-react";

// Eneba's footer opens with a light strip of payment logos. We can't use
// brand logos, so the strip says what the demo offers instead.
const PERKS = [
  { icon: ShieldCheck, label: "Secure checkout" },
  { icon: Zap, label: "Keys delivered at once" },
  { icon: CreditCard, label: "Simulated payments only" },
];

const LINKS = [
  { label: "Store", href: "/store" },
  { label: "Cart", href: "/cart" },
  { label: "My orders", href: "/orders" },
];

export function Footer() {
  return (
    <footer className="mt-10">
      <div className="bg-surface-light text-ink">
        <ul className="mx-auto flex max-w-page flex-wrap justify-center gap-x-10 gap-y-3 px-4 py-5 text-sm font-bold">
          {PERKS.map(({ icon: Icon, label }) => (
            <li key={label} className="flex items-center gap-2">
              <Icon className="size-5" aria-hidden />
              {label}
            </li>
          ))}
        </ul>
      </div>

      <div className="bg-brand-darkest">
        <div className="mx-auto flex max-w-page flex-col gap-6 px-4 py-10 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex flex-col gap-3">
            <Link href="/" className="flex items-center gap-2">
              <KeyRound className="size-6 text-yellow" strokeWidth={2.5} aria-hidden />
              <span className="text-xl font-extrabold">eugene</span>
            </Link>
            <p className="max-w-md text-sm text-text-muted">
              Demo project, not affiliated with Eneba. No real money is involved: products, sellers,
              covers and keys are all made up.
            </p>
          </div>

          <ul className="flex gap-6 text-sm font-semibold">
            {LINKS.map((l) => (
              <li key={l.href}>
                <Link href={l.href} className="text-text-secondary hover:text-white">
                  {l.label}
                </Link>
              </li>
            ))}
          </ul>
        </div>

        <p className="border-t border-divider px-4 py-4 text-center text-xs text-text-faint">
          © {new Date().getFullYear()} Eugene. Built as a homework project.
        </p>
      </div>
    </footer>
  );
}
