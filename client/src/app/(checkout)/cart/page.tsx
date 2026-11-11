import type { Metadata } from "next";
import { CartPage } from "@/features/cart/components/cart-page";

export const metadata: Metadata = { title: "Cart | Eugene" };

// ?payment=declined is read here and passed down, so the cart page needs no
// useSearchParams (which would need its own Suspense boundary).
export default async function Page({ searchParams }: { searchParams: Promise<{ payment?: string | string[] }> }) {
  const { payment } = await searchParams;
  return <CartPage declined={payment === "declined"} />;
}
