import type { ReactNode } from "react";
import Link from "next/link";
import { Heart, User } from "lucide-react";
import { Logo } from "./logo";

type Props = { search: ReactNode; cart: ReactNode };

// Server component. `search` (SearchBox) and `cart` (CartPopover) come from
// features and are passed in by the (shop) layout: layout components never
// import features. "Log in | Register" becomes the AccountMenu in phase 14.
export function Header({ search, cart }: Props) {
  return (
    <header className="bg-brand">
      {/* `relative`: the cart popover is placed against this row. */}
      <div className="relative mx-auto flex max-w-page flex-wrap items-center gap-x-6 gap-y-3 px-4 py-4 md:flex-nowrap md:py-5">
        <Logo />

        <div className="order-last w-full md:order-none md:max-w-[550px] md:flex-1">{search}</div>

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

          {cart}
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
