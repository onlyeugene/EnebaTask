import { Circle, CircleCheck } from "lucide-react";
import { PASSWORD_RULES } from "../schemas";

// The live checklist under the password fields (signup.png). Each rule ticks
// as soon as the typed password meets it. The rules are the ones the form
// checks on submit, so a ticked list means the password will be accepted.
export function PasswordRules({ password }: { password: string }) {
  return (
    <ul className="mt-4 flex flex-col gap-1.5 text-sm" aria-label="Password rules">
      {PASSWORD_RULES.map((r) => {
        const met = r.test(password);
        const Icon = met ? CircleCheck : Circle;
        return (
          <li key={r.label} className={`flex items-center gap-2.5 ${met ? "text-white" : "text-text-secondary"}`}>
            <Icon className={`size-3.5 shrink-0 ${met ? "text-mint" : ""}`} aria-hidden />
            {r.label}
            <span className="sr-only">{met ? "(done)" : "(not yet)"}</span>
          </li>
        );
      })}
    </ul>
  );
}
