// Where to go after login, from ?next=. Only a path on this site is allowed:
// "//evil.com" or "/\evil.com" would be read by the browser as another site,
// so a crafted link could send users there after they log in.
// Pointing back at the auth pages would loop, so those go home too.
export function safeNext(raw: string | string[] | undefined): string {
  const v = typeof raw === "string" ? raw : "";
  if (!v.startsWith("/") || v.startsWith("//") || v.startsWith("/\\")) return "/";
  if (v.startsWith("/login") || v.startsWith("/register")) return "/";
  return v;
}

// "?next=..." to carry the target over to the other auth page. Empty for home.
export function nextQuery(next: string) {
  return next === "/" ? "" : `?next=${encodeURIComponent(next)}`;
}
