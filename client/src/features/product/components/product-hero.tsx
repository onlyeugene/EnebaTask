import type { ReactNode } from "react";
import { CircleCheck, KeyRound, ReceiptText } from "lucide-react";
import type { ProductDetail, Region } from "@/types/api";
import { API_URL } from "@/lib/api-url";
import { platformLabel } from "@/lib/format";
import { PlatformIcon } from "@/components/ui/platform-icon";

const REGION_TEXT: Record<Region, { title: string; where: string }> = {
  global: { title: "Global", where: "anywhere in the world" },
  europe: { title: "Europe", where: "European countries" },
  united_states: { title: "United States", where: "the United States" },
  united_kingdom: { title: "United Kingdom", where: "the United Kingdom" },
};

// Top of the product page (product_details1.png): cover, title and the four
// facts. Eneba's fourth fact is "Instant refunds", which we don't offer, so
// ours says where the key shows up instead.
type Props = { product: ProductDetail; onShowGuide: () => void };

export function ProductHero({ product, onShowGuide }: Props) {
  const region = REGION_TEXT[product.region];
  const platform = platformLabel(product.platform);

  return (
    <div className="flex flex-col gap-8 sm:flex-row">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={API_URL + product.coverUrl}
        alt={`${product.name} cover`}
        width={222}
        height={250}
        className="h-auto w-[222px] shrink-0 self-start"
      />

      <div className="min-w-0 flex-1">
        {product.cashbackPct > 0 && (
          <span className="inline-block border border-mint px-3 py-1 text-base font-semibold text-mint">
            Cashback deal
          </span>
        )}
        <h1 className="mt-3 text-3xl font-extrabold">{product.name}</h1>

        <ul className="mt-8 grid gap-6 md:grid-cols-2">
          <Fact icon={<CircleCheck className="size-10 text-teal" strokeWidth={1.5} />} title={region.title}>
            Can be activated in {region.where}
          </Fact>
          <Fact icon={<PlatformIcon platform={product.platform} className="size-10" />} title={platform}>
            Activate/redeem on {platform}
            <br />
            <button type="button" onClick={onShowGuide} className="text-yellow hover:underline">
              Check activation guide
            </button>
          </Fact>
          <Fact icon={<KeyRound className="size-10" strokeWidth={1.5} />} title="Digital key">
            This is a digital edition of the product (CD-KEY)
            <br />
            <span className="text-teal">Instant delivery</span>
          </Fact>
          <Fact icon={<ReceiptText className="size-10" strokeWidth={1.5} />} title="Key on your order page">
            Your key shows on the order page as soon as payment goes through.
          </Fact>
        </ul>
      </div>
    </div>
  );
}

function Fact({ icon, title, children }: { icon: ReactNode; title: string; children: ReactNode }) {
  return (
    <li className="flex gap-4">
      <span className="shrink-0" aria-hidden>
        {icon}
      </span>
      <div>
        <p className="text-xl font-bold">{title}</p>
        <p className="mt-1.5 text-sm text-text-secondary">{children}</p>
      </div>
    </li>
  );
}
