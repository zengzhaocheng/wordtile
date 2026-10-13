import { create } from "zustand";
import type { User } from "../types";

interface AuthState {
  user: User | null;
  setUser: (u: User | null) => void;
  logout: () => void;
}

function loadUser(): User | null {
  try {
    const raw = localStorage.getItem("wt.user");
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  user: loadUser(),
  setUser: (u) => {
    if (u) localStorage.setItem("wt.user", JSON.stringify(u));
    else localStorage.removeItem("wt.user");
    set({ user: u });
  },
  logout: () => {
    localStorage.removeItem("wt.access");
    localStorage.removeItem("wt.refresh");
    localStorage.removeItem("wt.user");
    set({ user: null });
  },
}));

export function applyAuthResponse(data: { user: User; access_token: string; refresh_token: string }) {
  localStorage.setItem("wt.access", data.access_token);
  localStorage.setItem("wt.refresh", data.refresh_token);
  useAuthStore.getState().setUser(data.user);
}

export function isLoggedIn(): boolean {
  return !!localStorage.getItem("wt.access");
}
