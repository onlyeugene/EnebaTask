import type { Metadata } from "next";
import { LoginForm } from "@/features/auth/components/login-form";
import { safeNext } from "@/features/auth/next-path";

export const metadata: Metadata = { title: "Log in | Eugene" };

// ?next= is read here and passed down, so the form needs no useSearchParams
// (which would need its own Suspense boundary).
export default async function Page({ searchParams }: { searchParams: Promise<{ next?: string | string[] }> }) {
  const { next } = await searchParams;
  return <LoginForm next={safeNext(next)} />;
}
