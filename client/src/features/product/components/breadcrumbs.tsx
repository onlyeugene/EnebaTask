import Link from "next/link";
import { ChevronRight, House } from "lucide-react";
import type { ProductDetail, ProductType } from "@/types/api";
import { platformLabel } from "@/lib/format";

const TYPE_CRUMBS: Record<ProductType, string> = {
  game: "Games",
  gift_card: "Gift Cards",
  software: "Software",
};

// Home > Store > Games > Steam > name, as on Eneba. Each step is a store link
// with one more filter, so the trail is also a way back to similar products.
export function Breadcrumbs({ product }: { product: ProductDetail }) {
  const steps = [
    { label: "Store", href: "/store" },
    { label: TYPE_CRUMBS[product.type], href: `/store?type=${product.type}` },
    {
      label: platformLabel(product.platform),
      href: `/store?type=${product.type}&platform=${product.platform}`,
    },
  ];

  return (
    <nav aria-label="Breadcrumb" className="mb-6 text-base text-text-secondary">
      <ol className="flex flex-wrap items-center gap-x-2 gap-y-1">
        <li>
          <Link href="/" aria-label="Home" className="link-hover">
            <House className="size-4.5" />
          </Link>
        </li>
        {steps.map((s) => (
          <li key={s.href} className="flex items-center gap-2">
            <ChevronRight className="size-4" aria-hidden />
            <Link href={s.href} className="link-hover">
              {s.label}
            </Link>
          </li>
        ))}
        <li className="flex min-w-0 items-center gap-2">
          <ChevronRight className="size-4 shrink-0" aria-hidden />
          <span aria-current="page" className="max-w-50 truncate">
            {product.name}
          </span>
        </li>
      </ol>
    </nav>
  );
}
