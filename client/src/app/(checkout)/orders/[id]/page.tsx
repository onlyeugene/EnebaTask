import type { Metadata } from "next";
import { OrderDetail } from "@/features/orders/components/order-detail";

export const metadata: Metadata = { title: "Your order | Eugene" };

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <OrderDetail id={id} />;
}
