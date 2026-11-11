// A stand-in origin to resolve ?next= against. Any value works: it only has to
// be one a link can't name, so "same origin as this" means "a path on this site".
const BASE = "http://this-site.invalid";

// Where to go after login, from ?next=. Only a path on this site is allowed,
// so a crafted link can't send users to another site after they log in.
// Checking the text alone isn't enough: "//evil.com" and "/\evil.com" are
// other sites to the browser, and so is "/<tab>/evil.com", because the URL
// parser deletes tabs and newlines first. So the value is parsed the way the
// browser will parse it, and only the parsed path is returned.
// Pointing back at the auth pages would loop, so those go home too.
export function safeNext(raw: string | string[] | undefined): string {
  const v = typeof raw === "string" ? raw : "";
  if (!v.startsWith("/")) return "/";

  let url: URL;
  try {
    url = new URL(v, BASE);
  } catch {
    return "/";
  }
  if (url.origin !== BASE) return "/";

  const path = url.pathname + url.search + url.hash;
  if (path.startsWith("/login") || path.startsWith("/register")) return "/";
  return path;
}

// "?next=..." to carry the target over to the other auth page. Empty for home.
export function nextQuery(next: string) {
  return next === "/" ? "" : `?next=${encodeURIComponent(next)}`;
}
