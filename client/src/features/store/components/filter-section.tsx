"use client";

import { useId, useState, type ReactNode } from "react";
import { ChevronUp } from "lucide-react";

type Props = { title: string; active: number; children: ReactNode };

// One fold of the filter panel: title, a yellow count of picked values, and
// a chevron that folds it away (store2.png).
export function FilterSection({ title, active, children }: Props) {
  const [open, setOpen] = useState(true);
  const bodyId = useId();

  return (
    <section className="border-b-2 border-brand px-5 py-5 last:border-b-0">
      <h2>
        <button
          type="button"
          aria-expanded={open}
          aria-controls={bodyId}
          onClick={() => setOpen((o) => !o)}
          className="flex w-full items-center gap-2 text-base font-bold"
        >
          <span className="flex-1 text-left">{title}</span>
          {active > 0 && (
            <span className="flex size-5 items-center justify-center rounded-full bg-yellow text-xs font-bold text-ink">
              {active}
            </span>
          )}
          <ChevronUp
            className={`size-5 transition-transform ${open ? "" : "rotate-180"}`}
            aria-hidden
          />
        </button>
      </h2>
      <div id={bodyId} hidden={!open} className="mt-4">
        {children}
      </div>
    </section>
  );
}
