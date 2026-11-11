import type { Offer, ProductDetail } from "@/types/api";
import { OfferRow } from "./offer-row";

type Props = { product: ProductDetail; offers: Offer[] };

// "Recommended offers" holds the cheapest offer, "N other offers" the rest
// (product_details1–2.png). A product has at most 4 offers, so Eneba's
// "Load more offers" button isn't needed.
export function OfferList({ product, offers }: Props) {
  if (offers.length === 0) return null;
  const [first, ...rest] = offers;

  const row = (o: Offer) => (
    <OfferRow
      key={o.id}
      offer={o}
      originalCents={product.originalPriceCents}
      cashbackPct={product.cashbackPct}
    />
  );

  return (
    <section id="offers" className="flex scroll-mt-6 flex-col gap-10">
      <div>
        <Heading>Recommended offers</Heading>
        <ul className="flex flex-col gap-2">{row(first)}</ul>
      </div>
      {rest.length > 0 && (
        <div>
          <Heading>
            <span className="text-mint">{rest.length}</span> other{" "}
            {rest.length === 1 ? "offer" : "offers"}
          </Heading>
          <ul className="flex flex-col gap-2">{rest.map(row)}</ul>
        </div>
      )}
    </section>
  );
}

function Heading({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="mb-6 text-xl font-extrabold">{children}</h2>
  );
}
