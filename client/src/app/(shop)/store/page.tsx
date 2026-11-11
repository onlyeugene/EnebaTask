import type { Metadata } from "next";
import { StoreView } from "@/features/store/components/store-view";

export const metadata: Metadata = { title: "Store | Eugene" };

export default function StorePage() {
  return <StoreView />;
}
