import type { Metadata } from "next";
import { RegisterForm } from "@/features/auth/components/register-form";
import { safeNext } from "@/features/auth/next-path";

export const metadata: Metadata = { title: "Create account | Eugene" };

// Same as the login page: ?next= is read here and passed down.
export default async function Page({ searchParams }: { searchParams: Promise<{ next?: string | string[] }> }) {
  const { next } = await searchParams;
  return <RegisterForm next={safeNext(next)} />;
}
