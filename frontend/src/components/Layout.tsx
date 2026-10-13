import { NavLink, useNavigate } from "react-router-dom";
import { useI18n, type I18nKey } from "../i18n";
import { useAuthStore } from "../stores/auth";
import { useThemeStore } from "../stores/theme";
import type { ReactNode } from "react";

const navItems: { to: string; key: I18nKey }[] = [
  { to: "/", key: "nav.home" },
  { to: "/play", key: "nav.play" },
  { to: "/campaign", key: "nav.campaign" },
  { to: "/decks", key: "nav.decks" },
  { to: "/market", key: "nav.market" },
  { to: "/stats", key: "nav.stats" },
];

export default function Layout({ children }: { children: ReactNode }) {
  const { t, lang, setLang } = useI18n();
  const { user, logout } = useAuthStore();
  const { mode, setMode } = useThemeStore();
  const navigate = useNavigate();

  const themeIcon = mode === "dark" ? "☀️" : mode === "light" ? "🌙" : "🖥️";
  const nextTheme = mode === "system" ? "light" : mode === "light" ? "dark" : "system";

  return (
    <div className="flex h-full flex-col bg-bg text-ink">
      {/* 顶栏 */}
      <header className="flex items-center justify-between border-b border-line bg-surface px-4 py-2.5">
        <button className="flex items-center gap-2 text-lg font-bold" onClick={() => navigate("/")}>
          <span aria-hidden>🀄</span> WordTile
        </button>
        <div className="flex items-center gap-2">
          <button
            className="rounded-full border border-line px-2 py-1 text-xs"
            onClick={() => setLang(lang === "zh" ? "en" : "zh")}
            aria-label="switch language"
          >
            {t("lang.toggle")}
          </button>
          <button
            className="rounded-full border border-line px-2 py-1 text-xs"
            onClick={() => setMode(nextTheme)}
            aria-label="switch theme"
            title={`${t("theme.light")} / ${t("theme.dark")} / ${t("theme.system")}`}
          >
            {themeIcon}
          </button>
          {user && (
            <button
              className="rounded-full border border-line px-2 py-1 text-xs"
              onClick={() => {
                logout();
                navigate("/login");
              }}
            >
              {t("action.logout")}
            </button>
          )}
        </div>
      </header>

      {/* 主内容 */}
      <main className="flex-1 overflow-y-auto pb-20 md:pb-4">{children}</main>

      {/* 底部导航（移动端优先） */}
      <nav className="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-surface md:static md:border-t-0">
        <ul className="mx-auto flex max-w-3xl">
          {navItems.map((item) => (
            <li key={item.to} className="flex-1">
              <NavLink
                to={item.to}
                end={item.to === "/"}
                className={({ isActive }) =>
                  `flex flex-col items-center py-2 text-[11px] ${
                    isActive ? "font-semibold text-primary" : "text-muted"
                  }`
                }
              >
                {t(item.key)}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
