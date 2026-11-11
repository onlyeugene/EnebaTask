"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight, Menu, X } from "lucide-react";
import { Logo } from "@/components/layout/logo";
import { useFilters } from "../hooks/use-filters";

const TOP = [
  { label: "All offers", href: "/store" },
  { label: "Games", panel: true },
  { label: "Gift cards", href: "/store?type=gift_card" },
  { label: "Software", href: "/store?type=software" },
];
const PLATFORMS = [
  { label: "Steam", href: "/store?platform=steam" },
  { label: "PSN", href: "/store?platform=psn" },
  { label: "Xbox", href: "/store?platform=xbox" },
  { label: "Nintendo", href: "/store?platform=nintendo" },
  { label: "Windows", href: "/store?platform=windows" },
];

const ROW = "flex w-full items-center gap-2 px-6 py-3.5 text-base text-left nav-link-hover";

// The nav bar's "Categories" button and the drawer it opens from the left.
// It is a native <dialog> opened with showModal(), like the Modal: the
// browser dims the page, keeps focus inside and closes it on Escape.
// On phones the nav bar scrolls sideways, so this is the main way around.
export function CategoriesDrawer() {
  const [open, setOpen] = useState(false);
  const [panel, setPanel] = useState<"main" | "games">("main");
  const dialog = useRef<HTMLDialogElement>(null);
  const opener = useRef<HTMLButtonElement>(null);
  const gamesButton = useRef<HTMLButtonElement>(null);
  const backButton = useRef<HTMLButtonElement>(null);
  // Genres only load once the drawer has been opened.
  const filters = useFilters(open);

  useEffect(() => {
    const d = dialog.current;
    if (open && d && !d.open) d.showModal();
    if (!open && d?.open) d.close();
  }, [open]);

  // Moving between the two panels moves focus too, so keyboard users land
  // on the panel they just opened (and back on "Games" when they return).
  const showPanel = (p: "main" | "games") => {
    setPanel(p);
    requestAnimationFrame(() => (p === "games" ? backButton : gamesButton).current?.focus());
  };

  const close = () => setOpen(false);

  return (
    <>
      <button
        ref={opener}
        type="button"
        aria-haspopup="dialog"
        aria-expanded={open}
        onClick={() => {
          setPanel("main");
          setOpen(true);
        }}
        className="nav-link-hover flex items-center gap-2 px-2.5 py-4 text-sm leading-none font-semibold text-text-secondary"
      >
        <Menu className="size-5" aria-hidden />
        Categories
      </button>

      <dialog
        ref={dialog}
        aria-label="Categories"
        // Fires on Escape too. Focus goes back to the button that opened it.
        onClose={() => {
          setOpen(false);
          opener.current?.focus();
        }}
        // A click on the dim area lands on the <dialog> itself.
        onClick={(e) => e.target === e.currentTarget && close()}
        className="fixed inset-y-0 left-0 m-0 h-full max-h-none w-[350px] max-w-[85vw] bg-brand text-white backdrop:bg-ink/70"
      >
        {/* Inner box, so a click on the drawer's own padding isn't read as a click outside. */}
        <div className="flex h-full flex-col overflow-y-auto">
          <div className="flex items-center justify-between border-b border-divider px-6 py-4">
            <span onClick={close}>
              <Logo />
            </span>
            <button type="button" onClick={close} aria-label="Close categories" className="link-hover p-1">
              <X className="size-6" />
            </button>
          </div>

          {panel === "main" ? (
            <nav aria-label="Categories" className="py-2">
              <ul>
                {TOP.map((item) => (
                  <li key={item.label}>
                    {item.panel ? (
                      <button ref={gamesButton} type="button" onClick={() => showPanel("games")} className={`${ROW} justify-between`}>
                        {item.label}
                        <ChevronRight className="size-5" aria-hidden />
                      </button>
                    ) : (
                      <Link href={item.href!} onClick={close} className={ROW}>
                        {item.label}
                      </Link>
                    )}
                  </li>
                ))}
              </ul>
              <ul className="mt-2 border-t border-divider pt-2">
                {PLATFORMS.map((item) => (
                  <li key={item.label}>
                    <Link href={item.href} onClick={close} className={ROW}>
                      {item.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </nav>
          ) : (
            <nav aria-label="Games" className="py-2">
              <button ref={backButton} type="button" onClick={() => showPanel("main")} className={`${ROW} font-bold`}>
                <ChevronLeft className="size-5" aria-hidden />
                Back
              </button>
              <ul className="border-t border-divider pt-2">
                <li>
                  <Link href="/store?type=game" onClick={close} className={ROW}>
                    All games
                  </Link>
                </li>
                {filters.data?.genres.map((g) => (
                  <li key={g.value}>
                    <Link href={`/store?type=game&genre=${g.value}`} onClick={close} className={ROW}>
                      {g.label}
                    </Link>
                  </li>
                ))}
              </ul>
              {filters.isError && (
                <p className="px-6 py-3 text-sm text-text-muted">Genres couldn&apos;t load.</p>
              )}
            </nav>
          )}
        </div>
      </dialog>
    </>
  );
}
