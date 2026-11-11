import type { Metadata } from "next";
import { CartPage } from "@/features/cart/components/cart-page";

export const metadata: Metadata = { title: "Cart | Eugene" };

export default function Page() {
  return <CartPage />;
}
