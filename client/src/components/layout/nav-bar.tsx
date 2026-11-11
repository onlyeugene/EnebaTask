import Link from "next/link";
import { Menu } from "lucide-react";

// Fixed shortcuts into the store, like Eneba's nav bar. They are site
// navigation, not homepage content, so they live here and not in a block.
const LINKS = [
  { label: "Cheap Games", href: "/store?type=game&sort=price_asc" },
  { label: "Software", href: "/store?type=software" },
  { label: "Gift Cards", href: "/store?type=gift_card" },
  { label: "Deals", href: "/store?sort=discount" },
  { label: "New Releases", href: "/store?sort=newest" },
  { label: "Steam", href: "/store?platform=steam" },
  { label: "PSN", href: "/store?platform=psn" },
  { label: "Xbox", href: "/store?platform=xbox" },
];

// Eneba's nav link: 14px, weight 600, line-height 1, 10px padding.
const LINK = "nav-link-hover px-2.5 py-4 text-sm font-semibold leading-none text-text-secondary";

export function NavBar() {
  return (
    <nav className="bg-brand-dark" aria-label="Categories">
      {/* Scrolls sideways on small screens instead of wrapping. px-1.5 plus
          each link's 10px padding lines the first label up with the logo. */}
      <ul className="mx-auto flex h-12.5 max-w-page items-center gap-1 overflow-x-auto whitespace-nowrap px-1.5">
        <li>
          <Link href="/store" className={`flex items-center gap-2 ${LINK}`}>
            <Menu className="size-5" aria-hidden />
            Categories
          </Link>
        </li>
        {LINKS.map((l) => (
          <li key={l.href}>
            <Link href={l.href} className={`block ${LINK}`}>
              {l.label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
}
