"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

// A horizontal row that scrolls with touch, trackpad or the arrow buttons.
// Arrows exist only on devices that can hover (can-hover); touch users swipe.
//
// Two looks, as on Eneba:
// - "categories": thin white chevrons at the row's edges, always shown when
//   the row overflows, dimmed and disabled at either end, with a fade on the
//   right edge while there is more to scroll.
// - "rail": grey see-through squares over the cards, shown only while the
//   mouse is over the rail's section. The section must carry `group/rail`.
type Variant = "categories" | "rail";

export function ScrollRow({
  children,
  label,
  variant,
}: {
  children: ReactNode;
  label: string;
  variant: Variant;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [canLeft, setCanLeft] = useState(false);
  const [canRight, setCanRight] = useState(false);

  // The 1px slack covers rounding on high-DPI screens.
  const update = () => {
    const el = ref.current;
    if (!el) return;
    setCanLeft(el.scrollLeft > 1);
    setCanRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1);
  };

  useEffect(() => {
    update();
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, []);

  // One visible width per click. snap-x then lines the row up on an item.
  const scroll = (dir: 1 | -1) => {
    const el = ref.current;
    if (el) el.scrollBy({ left: dir * el.clientWidth, behavior: "smooth" });
  };

  const overflows = canLeft || canRight;
  const isCategories = variant === "categories";

  return (
    <div className="relative">
      <div
        ref={ref}
        onScroll={update}
        aria-label={label}
        // Categories leave room at both sides so the chevrons don't cover an item.
        // scroll-px matches px, or snapping would line the first item up with
        // the padded edge and scroll the row 32px on load.
        className={`flex snap-x gap-5 overflow-x-auto pb-2 [scrollbar-width:none] [&>*]:snap-start ${
          isCategories && overflows ? "can-hover:scroll-px-8 can-hover:px-8" : ""
        }`}
      >
        {children}
      </div>

      {isCategories && overflows && (
        <>
          {canRight && (
            <div
              aria-hidden
              // --band is set by the homepage band this row sits on, so the
              // fade blends into whatever color that is.
              className="pointer-events-none absolute inset-y-0 right-0 w-20 bg-linear-to-l from-(--band) to-transparent can-hover:right-8"
            />
          )}
          <Chevron side="left" enabled={canLeft} onClick={() => scroll(-1)} />
          <Chevron side="right" enabled={canRight} onClick={() => scroll(1)} />
        </>
      )}

      {!isCategories && canLeft && <Square side="left" onClick={() => scroll(-1)} />}
      {!isCategories && canRight && <Square side="right" onClick={() => scroll(1)} />}
    </div>
  );
}

type ArrowProps = { side: "left" | "right"; onClick: () => void };

// Vertically centred on the category icons: ITEM's 12px top padding + half of the 36px icon.
function Chevron({ side, enabled, onClick }: ArrowProps & { enabled: boolean }) {
  const Icon = side === "left" ? ChevronLeft : ChevronRight;
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={!enabled}
      aria-label={side === "left" ? "Scroll left" : "Scroll right"}
      className={`absolute top-[30px] hidden -translate-y-1/2 text-white transition-opacity hover:opacity-70 disabled:opacity-30 can-hover:block ${
        side === "left" ? "left-0" : "right-0"
      }`}
    >
      <Icon className="size-8" strokeWidth={1.25} />
    </button>
  );
}

// Hidden until the mouse is over the rail's section. Keyboard focus shows
// it too, so a tabbed-to arrow is never invisible.
function Square({ side, onClick }: ArrowProps) {
  const Icon = side === "left" ? ChevronLeft : ChevronRight;
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={side === "left" ? "Scroll left" : "Scroll right"}
      className={`absolute top-1/2 hidden -translate-y-1/2 bg-white/30 p-2 text-white opacity-0 transition-[opacity,background-color] duration-200 group-hover/rail:opacity-100 hover:bg-white/45 focus-visible:opacity-100 can-hover:block ${
        side === "left" ? "left-0" : "right-0"
      }`}
    >
      <Icon className="size-6" />
    </button>
  );
}
