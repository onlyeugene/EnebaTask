import type { InputHTMLAttributes, Ref } from "react";

type Props = InputHTMLAttributes<HTMLInputElement> & {
  id: string;
  error?: string;
  ref?: Ref<HTMLInputElement>; // React 19 passes ref as a prop, so RHF's register() works
};

// Eneba's auth field (login.png, signup.png): a darker box than the card,
// placeholder only, no visible label. The placeholder isn't read as a label
// by every screen reader, so pass aria-label too.
export function Input({ id, error, className = "", ...rest }: Props) {
  return (
    <div className={className}>
      <input
        id={id}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${id}-error` : undefined}
        className={`h-14 w-full border bg-brand-border px-5 text-base text-white outline-none placeholder:text-text-muted focus:border-text-faint ${
          error ? "border-red-pink" : "border-transparent"
        }`}
        {...rest}
      />
      {error && (
        <p id={`${id}-error`} className="mt-1.5 text-xs text-red-pink">
          {error}
        </p>
      )}
    </div>
  );
}
