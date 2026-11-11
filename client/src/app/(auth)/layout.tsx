import { Logo } from "@/components/layout/logo";
import { CheckoutFooter } from "@/components/layout/checkout-footer";

// login.png / signup.png: dark page, logo top left, a big greeting on the
// left and the form card on the right. No header, search or nav, like Eneba.
// The greeting hides below 930px, where the card takes the whole width.
// Eneba's footer links (terms, privacy, cookies) point at pages we don't
// have, so the footer is the checkout pages' demo line.
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-brand-darkest">
      <header className="px-4 py-6 sm:px-8">
        <Logo />
      </header>
      <main className="flex flex-1 items-center">
        <div className="mx-auto flex w-full max-w-wide items-center justify-center gap-10 px-4 py-10 md:justify-between lg:px-20">
          <p className="hidden min-w-0 flex-1 text-display font-extrabold md:block">
            Hi!
            <br />
            It&rsquo;s nice to see you!
          </p>
          {children}
        </div>
      </main>
      <CheckoutFooter />
    </div>
  );
}
