import { useQuery } from "@tanstack/react-query";
import type { ProductDetail } from "@/types/api";
import { listProducts } from "../api";

// Every value of the same gift card (same platform and region), cheapest
// value first. inStock=false lists sold-out values too, so a tile doesn't
// vanish when its keys run out. Off for anything that isn't a gift card.
export function useGiftCardFamily(product: ProductDetail) {
  const isGiftCard = product.type === "gift_card";
  return useQuery({
    queryKey: ["gift-card-family", product.platform, product.region],
    queryFn: ({ signal }) =>
      listProducts(
        {
          type: "gift_card",
          platform: product.platform,
          region: product.region,
          inStock: "false",
          pageSize: "60",
        },
        signal,
      ),
    enabled: isGiftCard,
    // A gift card's MSRP is its face value, so sorting by it sorts by value.
    select: (list) =>
      list.items
        .filter((p) => p.giftCardValue)
        .sort((a, b) => a.originalPriceCents - b.originalPriceCents),
  });
}
