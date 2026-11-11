import { LoaderCircle } from "lucide-react";

// A plain placeholder box. Size and color come from className.
export function Skeleton({ className = "" }: { className?: string }) {
  return <div aria-hidden className={`animate-pulse ${className}`} />;
}

// Dark card with a spinner in the middle, like Eneba's loading product cards.
export function CardSkeleton() {
  return (
    <div
      aria-hidden
      // The faint border keeps the card visible on the darkest homepage band.
      className="flex h-[470px] w-[224px] shrink-0 items-center justify-center border border-white/10 bg-brand-darkest"
    >
      <LoaderCircle className="size-9 animate-spin text-white" strokeWidth={1.5} />
    </div>
  );
}
