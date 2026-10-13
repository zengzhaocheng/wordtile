import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { deckApi, statsApi } from "../api/endpoints";
import { useI18n } from "../i18n";
import Heatmap from "../components/Heatmap";
import LineChart from "../components/LineChart";

export default function StatsPage() {
  const { t } = useI18n();
  const [deckId, setDeckId] = useState("");

  const { data: days } = useQuery({ queryKey: ["contribution"], queryFn: statsApi.contribution });
  const { data: decks } = useQuery({ queryKey: ["myDecks"], queryFn: deckApi.listMine });
  const { data: points } = useQuery({
    queryKey: ["deckProgress", deckId],
    queryFn: () => statsApi.deckProgress(deckId),
    enabled: !!deckId,
  });

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-4">
      <h1 className="text-xl font-bold">📊 {t("stats.title")}</h1>

      {/* 活跃度砖块图 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-3 font-semibold">{t("stats.contribution")}</h2>
        {days ? <Heatmap days={days} /> : <p className="text-sm text-muted">{t("common.loading")}</p>}
      </section>

      {/* 题库进步曲线 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-3 font-semibold">{t("stats.deckProgress")}</h2>
        <select
          value={deckId}
          onChange={(e) => setDeckId(e.target.value)}
          className="mb-3 w-full rounded-lg border border-line bg-bg px-3 py-2 text-sm"
        >
          <option value="">{t("stats.selectDeck")}</option>
          {decks?.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </select>

        {deckId && points && points.length === 0 && (
          <p className="text-sm text-muted">{t("stats.noData")}</p>
        )}
        {points && points.length > 0 && (
          <div className="space-y-4">
            <div>
              <h3 className="mb-1 text-sm font-medium">{t("stats.duration")}</h3>
              <LineChart
                points={points.map((p) => ({
                  label: new Date(p.completed_at).toLocaleDateString(),
                  value: p.duration_ms / 1000,
                }))}
              />
            </div>
            <div>
              <h3 className="mb-1 text-sm font-medium">{t("stats.errorRate")}</h3>
              <LineChart
                points={points.map((p) => ({
                  label: new Date(p.completed_at).toLocaleDateString(),
                  value: p.error_rate,
                }))}
                stroke="var(--danger)"
              />
            </div>
          </div>
        )}
      </section>
    </div>
  );
}
