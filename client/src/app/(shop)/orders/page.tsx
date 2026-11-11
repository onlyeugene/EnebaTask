import type { Metadata } from "next";
import { OrdersView } from "@/features/orders/components/orders-view";

export const metadata: Metadata = { title: "My orders | Eugene" };

export default function Page() {
  return <OrdersView />;
}
