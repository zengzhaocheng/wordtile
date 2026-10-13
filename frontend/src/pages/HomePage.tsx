import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { deckApi, statsApi } from "../api/endpoints";
import { formatDuration, useI18n } from "../i18n";
import { useAuthStore } from "../stores/auth";

export default function HomePage() {
  const { t } = useI18n();
  const user = useAuthStore((s) => s.user);
  const { data: days } = useQuery({ queryKey: ["contribution"], queryFn: statsApi.contribution });

  const today = new Date().toISOString().slice(0, 10);
  const todayStats = days?.find((d) => d.date === today);

  const { data: decks } = useQuery({ queryKey: ["myDecks"], queryFn: deckApi.listMine });
  const firstDeck = decks?.[0];

  return (
    <div className="mx-auto max-w-3xl space-y-6 p-4">
      <section>
        <h1 className="text-xl font-bold">
          {user?.display_name || user?.username} 👋
        </h1>
        <p className="text-sm text-muted">{t("app.tagline")}</p>
      </section>

      {/* 今日概览 */}
      <section className="grid grid-cols-3 gap-3">
        {[
          { label: t("home.sessions"), value: todayStats ? `${todayStats.sessions}` : "0" },
          { label: t("home.duration"), value: formatDuration(todayStats?.duration_ms ?? 0) },
          {
            label: t("home.accuracy"),
            value: `${Math.round((todayStats?.accuracy ?? 0) * 100)}%`,
          },
        ].map((item) => (
          <div key={item.label} className="rounded-2xl border border-line bg-surface p-3 text-center">
            <div className="text-2xl font-bold text-primary">{item.value}</div>
            <div className="text-xs text-muted">{item.label}</div>
          </div>
        ))}
      </section>

      {/* 快速开始 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-3 font-semibold">{t("home.quickStart")}</h2>
        <div className="grid grid-cols-2 gap-3">
          <Link
            to="/play"
            className="rounded-xl bg-primary px-4 py-6 text-center font-semibold text-primary-ink"
          >
            ⚡ {t("nav.play")}
          </Link>
          <Link
            to="/campaign"
            className="rounded-xl border-2 border-primary px-4 py-6 text-center font-semibold text-primary"
          >
            🏆 {t("nav.campaign")}
          </Link>
        </div>
        {firstDeck && (
          <p className="mt-3 text-xs text-muted">
            {t("play.selectDeck")}: {firstDeck.name}
          </p>
        )}
      </section>

      {/* 我的题库速览 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="font-semibold">{t("home.myDecks")}</h2>
          <Link to="/decks" className="text-sm text-primary underline">
            {t("nav.decks")} →
          </Link>
        </div>
        {decks && decks.length === 0 && <p className="text-sm text-muted">{t("home.noDeck")}</p>}
        <ul className="space-y-2">
          {decks?.slice(0, 5).map((d) => (
            <li key={d.id} className="flex items-center justify-between rounded-xl bg-raised px-3 py-2 text-sm">
              <span className="truncate">
                {d.name}
                {d.forked_from && <span className="ml-1 text-xs text-muted">(fork)</span>}
              </span>
              <Link to={`/play?deck=${d.id}`} className="text-primary underline">
                {t("action.start")}
              </Link>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
