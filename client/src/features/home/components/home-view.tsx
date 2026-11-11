import { Suspense } from "react";
import { HomeBlocks, HomeSkeleton } from "./home-blocks";

// Server component: the static frame. All content comes from <HomeBlocks />.
// Suspense is required by Next because HomeBlocks reads the URL
// (useSearchParams), which is only known in the browser on a static page.
export function HomeView() {
  return (
    <div>
      <h1 className="sr-only">Eugene: game keys, gift cards and software</h1>
      <Suspense fallback={<HomeSkeleton />}>
        <HomeBlocks />
      </Suspense>
    </div>
  );
}
