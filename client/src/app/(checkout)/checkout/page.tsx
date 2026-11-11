import type { Metadata } from "next";
import { CheckoutView } from "@/features/checkout/components/checkout-view";

export const metadata: Metadata = { title: "Checkout | Eugene" };

export default function Page() {
  return <CheckoutView />;
}
