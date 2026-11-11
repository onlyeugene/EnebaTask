"use client";

import type { ProductDetail } from "@/types/api";
import { platformLabel, regionLabel, typeLabel } from "@/lib/format";

export type DetailsTab = "description" | "activation";

type Props = { product: ProductDetail; tab: DetailsTab; onTab: (t: DetailsTab) => void };

const TABS: { id: DetailsTab; label: string }[] = [
  { id: "description", label: "Description" },
  { id: "activation", label: "Activation guide" },
];

// "Product Details" and "Other details" (product_details3.png). Eneba's
// reviews block is left out: reviews aren't built (PLAN.md section 9).
// The tab lives in ProductView, so "Check activation guide" at the top can open it.
export function ProductDetails({ product, tab, onTab }: Props) {
  const rows = [
    ["Type", typeLabel(product.type)],
    ["Platform", platformLabel(product.platform)],
    ["Region", regionLabel(product.region)],
    ["Developer", product.developer],
    ["Publisher", product.publisher],
    ["Release date", formatDate(product.releaseDate)],
  ].filter(([, v]) => v);

  return (
    <>
      <section id="details" className="scroll-mt-6">
        <h2 className="text-xl font-extrabold">Product Details</h2>
        {product.genres.length > 0 && (
          <ul className="mt-6 flex flex-wrap gap-2">
            {product.genres.map((g) => (
              <li key={g} className="bg-brand-dark px-3 py-1.5 text-base capitalize">
                {g}
              </li>
            ))}
          </ul>
        )}

        <div className="mt-4 bg-brand-dark">
          <div role="tablist" className="flex bg-brand-mid/40">
            {TABS.map((t) => (
              <button
                key={t.id}
                type="button"
                role="tab"
                aria-selected={tab === t.id}
                onClick={() => onTab(t.id)}
                className={`px-6 py-5 text-base ${tab === t.id ? "bg-brand-dark font-bold" : "text-text-secondary hover:text-white"}`}
              >
                {t.label}
              </button>
            ))}
          </div>
          <div role="tabpanel" className="p-6 text-base leading-7 text-text-secondary">
            {tab === "description" ? (
              <>
                <h3 className="mb-4 text-xl font-extrabold text-white">{product.name}</h3>
                <p>{product.description || "No description yet"}</p>
              </>
            ) : (
              <p className="whitespace-pre-line">{product.activation}</p>
            )}
          </div>
        </div>
      </section>

      <section>
        <h2 className="text-xl font-extrabold">Other details</h2>
        <dl className="mt-6 grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 bg-brand-dark p-6 text-base">
          {rows.map(([label, value]) => (
            <div key={label} className="contents">
              <dt className="text-text-muted">{label}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
      </section>
    </>
  );
}

// "2026-10-14" → "October 14, 2026". Fixed locale and UTC, so the
// date never shifts by a day with the visitor's time zone.
function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric", timeZone: "UTC" });
}
