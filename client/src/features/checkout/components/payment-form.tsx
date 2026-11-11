"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import type { Cart } from "@/types/api";
import { Input } from "@/components/ui/input";
import { useCheckout } from "../hooks/use-checkout";
import { paymentSchema, type PaymentValues } from "../schemas";
import { OrderSummary } from "./order-summary";

type Props = { cart: Cart; email: string; cartBusy: boolean };

// The right column of checkout.png: where the keys go, then the card, then
// the summary with the Pay button. One form, so Enter in any field pays.
export function PaymentForm({ cart, email, cartBusy }: Props) {
  const pay = useCheckout();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<PaymentValues>({ resolver: zodResolver(paymentSchema) });

  return (
    <form onSubmit={handleSubmit((v) => pay(v, setError))} noValidate className="flex flex-col gap-6">
      <section className="bg-brand-dark p-6">
        <h2 className="text-xl font-bold">Where to deliver the order?</h2>
        <p className="mt-4 text-sm">Email</p>
        <p className="mt-2 flex h-14 items-center bg-brand-border px-5 text-base text-text-secondary">{email}</p>
        <p className="mt-3 text-sm text-text-muted">Your keys show on the order page right after payment.</p>
      </section>

      <section className="bg-brand-dark p-6">
        <h2 className="text-xl font-bold">Payment</h2>
        <p className="mt-2 text-sm text-text-muted">
          Demo payment: any valid-looking card pays. 4000 0000 0000 0002 is declined.
        </p>
        <div className="mt-4 flex flex-col gap-4">
          <Field id="cardNumber" label="Card number">
            <Input
              id="cardNumber"
              inputMode="numeric"
              autoComplete="cc-number"
              placeholder="4242 4242 4242 4242"
              error={errors.cardNumber?.message}
              {...register("cardNumber")}
            />
          </Field>
          <div className="grid grid-cols-2 gap-4">
            <Field id="expiry" label="Expiry">
              <Input
                id="expiry"
                autoComplete="cc-exp"
                placeholder="MM/YY"
                error={errors.expiry?.message}
                {...register("expiry")}
              />
            </Field>
            <Field id="cvc" label="CVC">
              <Input
                id="cvc"
                inputMode="numeric"
                autoComplete="cc-csc"
                placeholder="123"
                error={errors.cvc?.message}
                {...register("cvc")}
              />
            </Field>
          </div>
          <Field id="name" label="Name on card">
            <Input id="name" autoComplete="cc-name" error={errors.name?.message} {...register("name")} />
          </Field>
        </div>
      </section>

      <OrderSummary
        cart={cart}
        error={errors.root?.message}
        submitting={isSubmitting}
        // A quantity change still on its way would make the order differ from the screen.
        disabled={cartBusy}
      />
    </form>
  );
}

// checkout.png shows a label above each box ("Email *"), unlike the auth pages.
function Field({ id, label, children }: { id: string; label: string; children: React.ReactNode }) {
  return (
    <div>
      <label htmlFor={id} className="mb-2 block text-sm">
        {label}
      </label>
      {children}
    </div>
  );
}
