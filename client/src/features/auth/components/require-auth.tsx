"use client";

import { useEffect, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import { LoaderCircle } from "lucide-react";
import { ErrorState } from "@/components/ui/error-state";
import { useMe } from "../hooks/use-me";

// Wraps pages that need a signed-in user (/orders, and /checkout in phase 15).
// Client-side only: tokens live in localStorage, which the Next server can't
// see. The real protection is the API answering 401 (PLAN.md section 8).
export function RequireAuth({ children }: { children: ReactNode }) {
  const me = useMe();
  const router = useRouter();
  const pathname = usePathname();
  const signedOut = me.isSuccess && !me.data;

  useEffect(() => {
    // replace, not push: Back shouldn't return to a page that bounces again.
    if (signedOut) router.replace(`/login?next=${encodeURIComponent(pathname)}`);
  }, [signedOut, router, pathname]);

  if (me.isError) return <ErrorState message="We couldn't check your account." onRetry={() => void me.refetch()} />;
  if (!me.data) {
    return (
      <div aria-busy className="flex justify-center py-16">
        <LoaderCircle className="size-9 animate-spin" strokeWidth={1.5} aria-hidden />
      </div>
    );
  }
  return children;
}
