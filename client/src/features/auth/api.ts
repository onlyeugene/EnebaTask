import { apiFetch } from "@/lib/client";
import type { AuthResponse, User } from "@/types/api";

export function login(email: string, password: string) {
  return apiFetch<AuthResponse>("/api/auth/login", { method: "POST", body: { email, password } });
}

export function register(email: string, password: string) {
  return apiFetch<AuthResponse>("/api/auth/register", { method: "POST", body: { email, password } });
}

export function logout(refreshToken: string) {
  return apiFetch<void>("/api/auth/logout", { method: "POST", body: { refreshToken } });
}

export function getMe(signal?: AbortSignal) {
  return apiFetch<{ user: User }>("/api/auth/me", { signal });
}
