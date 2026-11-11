"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronDown, LogOut, Package, User } from "lucide-react";
import { useLogout } from "../hooks/use-logout";
import { useMe } from "../hooks/use-me";
import { nextQuery, safeNext } from "../next-path";

// The header's right end. Guests see "Log in | Register" (with ?next= so they
// come back to this page); signed-in users see their name and a small menu
// with My orders and Log out. The (shop) layout passes it into the server Header.
export function AccountMenu() {
  const me = useMe();
  const pathname = usePathname();

  // Until /me answers, keep the space but show nothing, so a signed-in user
  // doesn't see "Log in" flash on every page load.
  if (me.isPending) return <div className="h-12 w-36" aria-hidden />;

  if (!me.data) {
    const q = nextQuery(safeNext(pathname));
    return (
      <div className="flex items-center text-base">
        <User className="mx-1 size-7" strokeWidth={1.75} aria-hidden />
        <Link href={`/login${q}`} className="link-hover p-2.5">
          Log in
        </Link>
        <span className="hidden text-text-muted xs:inline">|</span>
        <Link href={`/register${q}`} className="link-hover hidden p-2.5 xs:inline">
          Register
        </Link>
      </div>
    );
  }

  return <SignedIn name={me.data.name || me.data.email.split("@")[0]} email={me.data.email} />;
}

function SignedIn({ name, email }: { name: string; email: string }) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const logout = useLogout();

  // Close on a click outside, or on Escape (same as the cart popover).
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => !ref.current?.contains(e.target as Node) && setOpen(false);
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setOpen(false);
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        aria-haspopup="menu"
        className={`link-hover flex items-center gap-1.5 p-2.5 text-base ${open ? "text-yellow" : ""}`}
      >
        <User className="size-7" strokeWidth={1.75} aria-hidden />
        <span className="hidden max-w-32 truncate sm:inline">{name}</span>
        <ChevronDown className="size-4" aria-hidden />
      </button>

      {open && (
        <div
          role="menu"
          className="absolute top-full right-0 z-50 w-60 border border-brand-border bg-brand-mid py-2"
        >
          <p className="truncate border-b border-divider px-4 pt-1 pb-3 text-sm text-text-secondary">{email}</p>
          <Link
            href="/orders"
            role="menuitem"
            onClick={() => setOpen(false)}
            className="link-hover flex items-center gap-3 px-4 py-2.5 text-sm font-semibold"
          >
            <Package className="size-4.5" aria-hidden />
            My orders
          </Link>
          <button
            type="button"
            role="menuitem"
            onClick={logout}
            className="link-hover flex w-full items-center gap-3 px-4 py-2.5 text-sm font-semibold"
          >
            <LogOut className="size-4.5" aria-hidden />
            Log out
          </button>
        </div>
      )}
    </div>
  );
}
