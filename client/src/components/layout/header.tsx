import Link from "next/link";
import { Heart, KeyRound, Search, ShoppingCart, User } from "lucide-react";

// Server component. The search box is a plain GET form for now, so Enter opens
// /store?q=... even without JS; phase 12 swaps in the autocomplete SearchBox.
// The cart icon gets its count badge (CartButton) in phase 13, and
// "Log in | Register" becomes the AccountMenu in phase 14.
export function Header() {
  return (
    <header className="bg-brand">
      <div className="mx-auto flex max-w-page flex-wrap items-center gap-x-6 gap-y-3 px-4 py-4 md:flex-nowrap md:py-5">
        <Link
          href="/"
          className="flex shrink-0 items-center gap-2"
          aria-label="Eugene home"
        >
          <KeyRound className="size-8 text-yellow" strokeWidth={2.5} />
          <span className="text-3xl font-extrabold tracking-tight">eugene</span>
        </Link>

        <form
          action="/store"
          role="search"
          className="order-last flex h-12 w-full items-center gap-3 border border-white px-2.5 md:order-none md:h-[57px] md:max-w-[550px] md:flex-1"
        >
          <Search className="size-6 shrink-0 text-text-soft" aria-hidden />
          <input
            name="q"
            type="search"
            maxLength={100}
            placeholder="Search for games, gift cards and more"
            aria-label="Search products"
            className="h-full w-full bg-transparent text-base font-semibold text-white outline-none placeholder:text-text-muted"
          />
        </form>

        {/* Static for now (EUR and English only), but hovers like Eneba's picker.
            The text turns yellow; on Eneba only the nav bar gets a box. */}
        <span className="link-hover hidden shrink-0 p-2.5 text-base text-text-soft lg:block">
          English EU | EUR
        </span>

        {/* 10px padding on each item gives a bigger click and tap target; the
            small gaps and -mr-2.5 keep the spacing and right edge as before. */}
        <nav
          className="-mr-2.5 ml-auto flex shrink-0 items-center gap-1"
          aria-label="Account"
        >
          <Link href="/wishlist" aria-label="Wishlist" className="link-hover p-2.5">
            <Heart className="size-7" strokeWidth={1.75} />
          </Link>

          <Link href="/cart" aria-label="Cart" className="link-hover p-2.5">
            <ShoppingCart className="size-7" strokeWidth={1.75} />
          </Link>
          <div className="flex items-center text-base">
            <User className="mx-1 size-7" strokeWidth={1.75} aria-hidden />
            <Link href="/login" className="link-hover p-2.5">
              Log in
            </Link>
            <span className="hidden text-text-muted xs:inline">|</span>
            <Link href="/register" className="link-hover hidden p-2.5 xs:inline">
              Register
            </Link>
          </div>
        </nav>
      </div>
    </header>
  );
}
