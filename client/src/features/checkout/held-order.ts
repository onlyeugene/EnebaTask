// The order Pay should finish: created, but its payment not known to be
// final. Kept in sessionStorage so it survives a refresh mid-payment (one tab
// only, gone when the tab closes). useResumeOrder reads it when checkout or
// the cart page opens; PLAN.md "Held order".
//
// Every access is wrapped: storage can throw (private mode, blocked site
// data), and losing the id only means the next Pay reserves a fresh order.
const KEY = "eugene.held-order";

export function getHeldOrder(): string | null {
  try {
    return sessionStorage.getItem(KEY);
  } catch {
    return null;
  }
}

export function setHeldOrder(id: string) {
  try {
    sessionStorage.setItem(KEY, id);
  } catch {}
}

// Called when the order is final (paid, declined, expired, gone) and whenever
// the server cart changes: the held order was reserved from the old cart, so
// paying it would charge for something no longer on screen.
export function clearHeldOrder() {
  try {
    sessionStorage.removeItem(KEY);
  } catch {}
}
