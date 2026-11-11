import type { ReactNode } from "react";
import Link from "next/link";

type Props = {
  title: string;
  question: string; // "New user?"
  linkLabel: string; // "Create account"
  href: string;
  children: ReactNode;
};

// The purple box on the right of login.png and signup.png: title, a line
// linking to the other auth page, then the form.
export function AuthCard({ title, question, linkLabel, href, children }: Props) {
  return (
    <section className="w-full bg-brand px-6 py-8 sm:w-[380px] sm:px-10">
      <h1 className="text-xl font-extrabold tracking-wide">{title}</h1>
      <p className="mt-2 text-sm font-semibold text-text-secondary">
        {question}{" "}
        <Link href={href} className="text-yellow hover:underline">
          {linkLabel}
        </Link>
      </p>
      <div className="mt-7">{children}</div>
    </section>
  );
}
