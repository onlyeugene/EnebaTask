// Two small pieces both auth forms use.

// Eneba's yellow "Show password" text button (signup.png). It shows every
// password field of the form at once.
export function ShowPassword({ show, onToggle }: { show: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-pressed={show}
      className="-mt-1 self-start text-sm text-yellow hover:underline"
    >
      {show ? "Hide password" : "Show password"}
    </button>
  );
}

// The error that belongs to no single field: wrong password, rate limit, server down.
export function FormError({ message }: { message?: string }) {
  if (!message) return null;
  return (
    <p role="alert" className="border border-red-dark bg-brand-darkest px-4 py-3 text-sm text-white">
      {message}
    </p>
  );
}
