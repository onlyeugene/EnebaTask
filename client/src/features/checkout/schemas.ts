import { z } from "zod";

// The server's card rules (API.md, Orders; server/internal/orders/payment.go),
// checked here first so most mistakes show without a request. The server
// checks again, and its 422 fields land under the same inputs.

// "4242 4242-4242" → "424242424242". The server strips the same characters.
export function cardDigits(s: string) {
  return s.replace(/[\s-]/g, "");
}

// A card works until the end of its expiry month, as on the server.
function notExpired(mmyy: string) {
  const [mm, yy] = mmyy.split("/").map((p) => Number(p.trim()));
  const now = new Date();
  // Month index `mm` (0-based) is the month after the expiry month.
  return now < new Date(Date.UTC(2000 + yy, mm, 1));
}

export const paymentSchema = z.object({
  cardNumber: z
    .string()
    .refine((v) => /^\d{12,19}$/.test(cardDigits(v)), "Card number must be 12 to 19 digits"),
  expiry: z
    .string()
    .trim()
    .regex(/^(0[1-9]|1[0-2])\s*\/\s*\d{2}$/, "Expiry must be MM/YY")
    .refine(notExpired, "This card has expired"),
  cvc: z.string().trim().regex(/^\d{3,4}$/, "CVC must be 3 or 4 digits"),
  name: z.string().trim().min(1, "Name on card is required").max(100, "At most 100 characters"),
});

export type PaymentValues = z.infer<typeof paymentSchema>;
