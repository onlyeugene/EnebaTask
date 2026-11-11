import type { FieldValues, Path, UseFormSetError } from "react-hook-form";
import { ApiError } from "@/lib/client";

// Puts an API error (API.md, Auth) on the form: field errors under their
// field, everything else as one message above the button ("root").
export function showServerError<T extends FieldValues>(e: unknown, setError: UseFormSetError<T>) {
  if (!(e instanceof ApiError)) {
    setError("root", { message: "Couldn't reach the server. Check your connection and try again." });
    return;
  }
  if (e.code === "EMAIL_TAKEN") {
    setError("email" as Path<T>, { message: "An account with this email already exists" });
    return;
  }
  if (e.fields) {
    // The server's field messages are short ("is not a valid email"), so the
    // field name goes in front.
    for (const [name, msg] of Object.entries(e.fields)) {
      setError(name as Path<T>, { message: `${name[0].toUpperCase()}${name.slice(1)} ${msg}` });
    }
    return;
  }
  // 401 "Invalid email or password", 429 "Too many attempts…": the server's text fits as is.
  setError("root", { message: e.status >= 500 ? "Something went wrong. Please try again." : e.message });
}
