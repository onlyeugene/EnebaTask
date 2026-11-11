import { ShoppingCart } from "lucide-react";

// Eneba's small "+cart" button next to "Buy now".
export function AddButton({ onClick, className = "" }: { onClick: () => void; className?: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label="Add to cart"
      className={`flex h-12 w-17 shrink-0 items-center justify-center gap-0.5 border border-white text-white transition-colors hover:border-yellow hover:bg-brand-darkest ${className}`}
    >
      <span className="text-xl font-bold" aria-hidden>
        +
      </span>
      <ShoppingCart className="size-5" aria-hidden />
    </button>
  );
}
