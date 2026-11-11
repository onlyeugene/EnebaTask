import Link from "next/link";
import { Gamepad, Gamepad2, Info, Joystick, Monitor, Puzzle, type LucideIcon } from "lucide-react";
import type { Platform, ProductCard as Card } from "@/types/api";
import { API_URL } from "@/lib/api-url";
import {
  cashbackCents,
  discountLabel,
  platformLabel,
  regionColor,
  regionLabel,
} from "@/lib/format";
import { CashbackBadge, GiftCardTag } from "./badge";
import { buttonStyles } from "./button";
import { Price } from "./price";

// Generic icons, not brand logos (see notes.md, Branding).
const PLATFORM_ICONS: Record<Platform, LucideIcon> = {
  steam: Gamepad2,
  psn: Gamepad,
  xbox: Joystick,
  nintendo: Puzzle,
  windows: Monitor,
};

// Hover: the strip and info slide up by the height of the button panel
// (2 buttons of 44px + 8px gap + 20px bottom padding = 116px) while the
// panel slides in from below. The card has a fixed height and hides what
// moves outside it, so nothing around it shifts.
// Mouse users get it on hover, keyboard users when focus enters the card.
const SLIDE = "transition-transform duration-200 ease-out";
const UP = "group-hover:-translate-y-[116px] can-hover:group-focus-within:-translate-y-[116px]";

type Props = {
  product: Card;
  // Wired up in phase 13 (cart). Until then the button does nothing.
  onAddToCart?: (product: Card) => void;
};

export function ProductCard({ product, onAddToCart }: Props) {
  const PlatformIcon = PLATFORM_ICONS[product.platform];
  const price = product.minPriceCents;
  const href = `/product/${product.slug}`;

  return (
    <div className="group relative flex h-[470px] w-[224px] shrink-0 flex-col overflow-hidden border border-mint bg-brand-darkest">
      {/* The link covers the whole card, so a tap on a touch screen (where
          the buttons below don't exist) opens the product page. */}
      <Link href={href} className="flex flex-1 flex-col">
        <div className="relative h-[250px] shrink-0 overflow-hidden">
          {/* Plain <img>: covers are SVGs from our API, and next/image refuses
              SVG unless dangerouslyAllowSVG is turned on. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={API_URL + product.coverUrl}
            alt=""
            width={222}
            height={250}
            loading="lazy"
            className="size-full object-cover transition-transform duration-300 group-hover:scale-105"
          />
          {product.type === "gift_card" && product.giftCardValue && (
            <span className="absolute top-2 right-2">
              <GiftCardTag value={product.giftCardValue} />
            </span>
          )}
        </div>

        {/* `relative` puts the sliding part above the cover it moves over. */}
        <div className={`relative flex flex-1 flex-col bg-brand-darkest ${SLIDE} ${UP}`}>
          {/* On Eneba this strip lies over the bottom of the cover. Our generated
              covers print the title there, so it sits below the cover instead. */}
          <div className="flex h-[30px] items-center justify-between gap-2 border-b border-divider pr-3">
            {product.cashbackPct > 0 ? <CashbackBadge /> : <span />}
            <span className="flex items-center gap-1.5 text-xs font-semibold">
              <PlatformIcon className="size-3.5" aria-hidden />
              {platformLabel(product.platform)}
            </span>
          </div>
          <CardInfo product={product} />
        </div>
      </Link>

      {/* Not rendered at all on touch screens (display: none), so it can't be
          tapped or focused there. */}
      <div
        className={`absolute inset-x-0 bottom-0 hidden translate-y-full flex-col gap-2 bg-brand-darkest px-5 pb-5 can-hover:flex group-hover:translate-y-0 group-focus-within:translate-y-0 ${SLIDE}`}
      >
        <button
          type="button"
          disabled={price === null}
          onClick={() => onAddToCart?.(product)}
          className={`${buttonStyles.primary} h-11 w-full py-0 text-base`}
        >
          {price === null ? "Sold out" : "Add to cart"}
        </button>
        <Link href={href} className={`${buttonStyles.outline} h-11 w-full py-0 text-base`}>
          View offers
        </Link>
      </div>
    </div>
  );
}

function CardInfo({ product }: { product: Card }) {
  const price = product.minPriceCents;
  return (
    <div className="flex flex-1 flex-col gap-1.5 p-5">
      <h3 className="line-clamp-2 min-h-10 text-sm font-bold">{product.name}</h3>
      <p className={`text-sm font-bold ${regionColor(product.region)}`}>
        {regionLabel(product.region)}
      </p>

      {price === null ? (
        <p className="mt-auto text-xl font-extrabold text-text-muted">Sold out</p>
      ) : (
        <div className="mt-auto">
          <p className="text-sm font-semibold text-text-muted">
            From{" "}
            {product.discountPct > 0 && (
              <>
                <Price cents={product.originalPriceCents} className="line-through" />{" "}
                <span className="text-lime">{discountLabel(product.discountPct)}</span>
              </>
            )}
          </p>
          <p className="flex items-center gap-1.5">
            <Price cents={price} className="text-3xl font-extrabold" />
            <Info className="size-4 text-text-muted" aria-hidden />
          </p>
          {product.cashbackPct > 0 && (
            <p className="text-sm font-bold text-lime">
              Cashback: <Price cents={cashbackCents(price, product.cashbackPct)} />
            </p>
          )}
        </div>
      )}
    </div>
  );
}
