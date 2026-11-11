import { usePathname, useRouter, useSearchParams } from "next/navigation";

export type ListFilter = "type" | "platform" | "region" | "genre";

// Everything that narrows the list. `q` and `sort` are not filters, so
// "Clear all" keeps them.
const FILTER_KEYS = ["type", "platform", "region", "genre", "minPrice", "maxPrice"];

// The URL is the store's only state: filters, sort and page all live in the
// query string, so a link, a reload or the Back button shows the same list.
export function useStoreParams() {
  const params = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  // "steam,psn" → ["steam", "psn"]
  function list(name: ListFilter): string[] {
    return (params.get(name) ?? "").split(",").filter(Boolean);
  }

  // null removes a param. Any change that isn't a page change goes back to
  // page 1, since page 3 of the old list means nothing in the new one.
  // push (not replace), so Back undoes the last filter.
  function update(changes: Record<string, string | null>, scroll = false) {
    const next = new URLSearchParams(params.toString());
    for (const [key, value] of Object.entries(changes)) {
      if (value) next.set(key, value);
      else next.delete(key);
    }
    if (!("page" in changes)) next.delete("page");
    const qs = next.toString();
    router.push(qs ? `${pathname}?${qs}` : pathname, { scroll });
  }

  function toggle(name: ListFilter, value: string) {
    const current = list(name);
    const next = current.includes(value)
      ? current.filter((v) => v !== value)
      : [...current, value];
    update({ [name]: next.join(",") || null });
  }

  function clear() {
    update(Object.fromEntries(FILTER_KEYS.map((k) => [k, null])));
  }

  return { params, q: params.get("q") ?? "", list, update, toggle, clear };
}
