import type { ButtonHTMLAttributes } from "react";

// Class strings are exported too, so a <Link> can look like a button
// without wrapping a <button> inside an <a>.
export const buttonStyles = {
  primary:
    "inline-flex items-center justify-center gap-2 bg-yellow px-6 py-3 text-sm font-bold text-ink transition-[filter] hover:brightness-110 disabled:opacity-60",
  outline:
    "inline-flex items-center justify-center gap-2 border hover:border-[#fad318] border-white px-6 py-3 text-sm font-bold text-white transition-colors hover:bg-[#1f0a4d] disabled:opacity-60",
};

type Props = ButtonHTMLAttributes<HTMLButtonElement> & { variant?: keyof typeof buttonStyles };

export function Button({ variant = "primary", className = "", type = "button", ...rest }: Props) {
  return <button type={type} className={`${buttonStyles[variant]} ${className}`} {...rest} />;
}
