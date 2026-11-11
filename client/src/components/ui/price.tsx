import { formatCents } from "@/lib/format";

// Cents → "€15.49". Size and weight come from className.
export function Price({ cents, className = "" }: { cents: number; className?: string }) {
  return <span className={className}>{formatCents(cents)}</span>;
}
