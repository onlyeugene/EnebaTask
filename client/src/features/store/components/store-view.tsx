import { Suspense } from "react";
import { CardSkeleton } from "@/components/ui/skeleton";
import { StoreContent } from "./store-content";

// Server component: the page frame. Suspense is required by Next because
// StoreContent reads the URL (useSearchParams), as on the home page.
// pb-10: each page ends with its own 40px before the footer.
export function StoreView() {
  return (
    <div className="mx-auto max-w-page px-4 pt-6 pb-10">
      <Suspense fallback={<StoreSkeleton />}>
        <StoreContent />
      </Suspense>
    </div>
  );
}

function StoreSkeleton() {
  return (
    <div aria-busy className="grid grid-cols-[repeat(auto-fill,224px)] gap-5 pt-20">
      {Array.from({ length: 4 }, (_, i) => (
        <CardSkeleton key={i} />
      ))}
    </div>
  );
}
