import Link from "next/link";
import { KeyRound } from "lucide-react";

// The "eugene" wordmark, linking home. Used by both the shop and checkout headers.
export function Logo() {
  return (
    <Link href="/" className="flex shrink-0 items-center gap-2" aria-label="Eugene home">
      <KeyRound className="size-8 text-yellow" strokeWidth={2.5} />
      <span className="text-3xl font-extrabold tracking-tight">eugene</span>
    </Link>
  );
}
