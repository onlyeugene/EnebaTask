import { notFound } from "next/navigation";

// Any URL no other route matches lands here (a catch-all ranks below every
// other route, so /cart, /login and the rest are unaffected). A group's
// not-found.tsx only catches notFound() calls, not unknown URLs, so without
// this route an unknown URL would get Next's plain 404 with no header or nav.
// Called on the server, so the response is a real 404.
export default function Page() {
  notFound();
}
