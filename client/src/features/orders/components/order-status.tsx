import type { OrderStatus as Status } from "@/types/api";

const LOOK: Record<Status, { label: string; className: string }> = {
  paid: { label: "Paid", className: "bg-mint text-ink" },
  pending_payment: { label: "Awaiting payment", className: "bg-yellow text-ink" },
  failed: { label: "Payment declined", className: "bg-red-pink text-white" },
  expired: { label: "Expired", className: "bg-brand-border text-text-secondary" },
};

// Status tag for the order page and the history list.
export function OrderStatus({ status }: { status: Status }) {
  const { label, className } = LOOK[status];
  return <span className={`inline-block px-2 py-0.5 text-xs font-bold ${className}`}>{label}</span>;
}
