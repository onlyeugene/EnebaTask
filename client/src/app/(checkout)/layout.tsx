import { CheckoutHeader } from "@/components/layout/checkout-header";
import { CheckoutFooter } from "@/components/layout/checkout-footer";

// Cart, checkout and order pages: stepper header, short footer.
export default function CheckoutLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <CheckoutHeader />
      <main className="flex-1">{children}</main>
      <CheckoutFooter />
    </div>
  );
}
