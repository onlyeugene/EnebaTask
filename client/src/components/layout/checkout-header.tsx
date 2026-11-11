import { CheckoutSteps } from "./checkout-steps";
import { Logo } from "./logo";

// Server component. Header for the cart, checkout and order pages, as on
// Eneba: no search or nav bar, so nothing pulls the buyer away mid-purchase.
// Darker than the shop header (brand-dark), so the checkout pages read as a
// separate step. The language text is hidden on phones
// to leave room for the steps.
export function CheckoutHeader() {
  return (
    <header className="bg-brand-dark">
      <div className="mx-auto flex max-w-page items-center gap-8 px-4 py-5">
        <Logo />
        <CheckoutSteps />
        <span className="hidden shrink-0 text-base text-text-soft sm:block">English EU | EUR</span>
      </div>
    </header>
  );
}
