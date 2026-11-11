"use client";

import { useState } from "react";
import { notFound } from "next/navigation";
import { ApiError } from "@/lib/client";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useProduct } from "../hooks/use-product";
import { AlsoViewed } from "./also-viewed";
import { Breadcrumbs } from "./breadcrumbs";
import { OfferList } from "./offer-list";
import { PriceBox } from "./price-box";
import { ProductDetails, type DetailsTab } from "./product-details";
import { ProductHero } from "./product-hero";
import { ValuePicker } from "./value-picker";

// The product page. Eneba draws it wider than the home page, so it uses
// `max-w-wide` (1600px). From 930px up: main column + a 350px sidebar, 30px
// apart. The sidebar's grid cell spans both rows, so its sticky box can
// follow the whole main column while scrolling. Below 930px it's one column
// and the grid order puts the price box between the hero and the offers.
// pb-10: each page ends with its own 40px before the footer.
export function ProductView({ slug }: { slug: string }) {
  const query = useProduct(slug);
  const [tab, setTab] = useState<DetailsTab>("description");

  const showGuide = () => {
    setTab("activation");
    document.getElementById("details")?.scrollIntoView({ behavior: "smooth" });
  };

  // An unknown or inactive slug is a 404 from the API: show the site's 404
  // page ((shop)/not-found.tsx). notFound() works from a client component, but
  // the data loads in the browser, so the HTTP status stays 200.
  if (query.isError && !query.isFetching && query.error instanceof ApiError && query.error.status === 404) {
    notFound();
  }

  let body;
  if (query.isPending || (query.isError && query.isFetching)) {
    body = <ProductSkeleton />;
  } else if (query.isError) {
    body = <ErrorState message="This product couldn't load." onRetry={() => void query.refetch()} />;
  } else {
    const { product, offers } = query.data;
    body = (
      <>
        <Breadcrumbs product={product} />
        <div className="grid gap-x-7.5 gap-y-12 md:grid-cols-[1fr_350px]">
          <div className="min-w-0 md:col-start-1">
            <ProductHero product={product} onShowGuide={showGuide} />
          </div>
          <div className="md:col-start-2 md:row-span-2 md:row-start-1">
            <div className="md:sticky md:top-4">
              <PriceBox product={product} />
            </div>
          </div>
          <div className="flex min-w-0 flex-col gap-12 md:col-start-1">
            <ValuePicker product={product} />
            <OfferList product={product} offers={offers} />
            <AlsoViewed product={product} />
            <ProductDetails product={product} tab={tab} onTab={setTab} />
          </div>
        </div>
      </>
    );
  }

  return <div className="mx-auto max-w-wide px-4 pt-6 pb-10">{body}</div>;
}

function ProductSkeleton() {
  return (
    <div aria-busy className="flex flex-col gap-8 pt-10 sm:flex-row">
      <Skeleton className="h-[250px] w-[222px] shrink-0 bg-brand-light/40" />
      <div className="flex flex-1 flex-col gap-4">
        <Skeleton className="h-9 w-3/4 bg-brand-light/40" />
        <Skeleton className="h-24 w-full bg-brand-light/20" />
      </div>
    </div>
  );
}
