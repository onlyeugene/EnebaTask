import { z } from "zod";

// Same rules as the server (API.md, Auth), so the form catches them before a request.
// Length is counted in bytes, as the server does: bcrypt's 72 limit is in bytes.
const bytes = (s: string) => new TextEncoder().encode(s).length;

// Server: unicode.IsDigit || IsPunct || IsSymbol. These are the same unicode classes.
const DIGIT_OR_SYMBOL = /[\p{Nd}\p{P}\p{S}]/u;

export const PASSWORD_RULES = [
  { label: "At least 8 characters", test: (p: string) => bytes(p) >= 8 },
  { label: "At least one number (0-9) or special symbol", test: (p: string) => DIGIT_OR_SYMBOL.test(p) },
];

const email = z.string().trim().min(1, "Enter your email").email("Enter a valid email");

export const loginSchema = z.object({
  email,
  password: z.string().min(1, "Enter your password"),
});

export const registerSchema = z
  .object({
    email,
    password: z
      .string()
      .refine((p) => PASSWORD_RULES.every((r) => r.test(p)), "Password doesn't meet the rules below")
      .refine((p) => bytes(p) <= 72, "Password is too long (72 characters at most)"),
    repeat: z.string(),
  })
  .refine((v) => v.password === v.repeat, { path: ["repeat"], message: "Passwords don't match" });

export type LoginValues = z.infer<typeof loginSchema>;
export type RegisterValues = z.infer<typeof registerSchema>;
