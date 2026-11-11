import { clearTokens, getTokens } from "@/lib/tokens";
import { logout } from "../api";

// Ends the session on the server, then here. The local tokens are cleared
// even if the call fails (API.md, Auth), so the user is always logged out on
// this device. A full page load to home then drops every cached query of
// this user (cart, orders), so nothing of theirs stays on screen.
export function useLogout() {
  return async () => {
    const tokens = getTokens();
    if (tokens) await logout(tokens.refreshToken).catch(() => {});
    clearTokens();
    window.location.assign("/");
  };
}
