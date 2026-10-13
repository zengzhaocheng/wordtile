import { create } from "zustand";

export type ThemeMode = "light" | "dark" | "system";

interface ThemeState {
  mode: ThemeMode;
  setMode: (m: ThemeMode) => void;
}

function persistedMode(): ThemeMode {
  const v = localStorage.getItem("wt.theme");
  return v === "light" || v === "dark" || v === "system" ? v : "system";
}

function systemDark(): boolean {
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function applyTheme(mode: ThemeMode) {
  const dark = mode === "dark" || (mode === "system" && systemDark());
  document.documentElement.classList.toggle("dark", dark);
  document.documentElement.style.colorScheme = dark ? "dark" : "light";
}

// 跟随系统变化
window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
  applyTheme(useThemeStore.getState().mode);
});

export const useThemeStore = create<ThemeState>((set) => ({
  mode: persistedMode(),
  setMode: (m) => {
    localStorage.setItem("wt.theme", m);
    applyTheme(m);
    set({ mode: m });
  },
}));
