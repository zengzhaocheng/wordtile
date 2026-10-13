import type { ContributionDay } from "../types";
import { buildCalendar } from "../utils/heatmap";
import { useI18n, formatDuration } from "../i18n";

const levelClass = ["bg-raised", "hm-l1", "hm-l2", "hm-l3", "hm-l4"];

export default function Heatmap({ days }: { days: ContributionDay[] }) {
  const { t } = useI18n();
  const columns = buildCalendar(days, 26);

  return (
    <div className="overflow-x-auto">
      <div className="min-w-max">
        <div className="flex gap-[3px]">
          {columns.map((col, i) => (
            <div key={i} className="flex flex-col gap-[3px]">
              {col.map((cell) => (
                <div
                  key={cell.date}
                  className={`h-3 w-3 rounded-[3px] ${levelClass[cell.level]} ${
                    cell.inRange ? "" : "opacity-30"
                  }`}
                  title={
                    cell.inRange
                      ? `${cell.date} · ${cell.sessions} ${t("home.sessions")} · ${formatDuration(
                          cell.durationMs,
                        )} · ${Math.round(cell.accuracy * 100)}%`
                      : cell.date
                  }
                  role="img"
                  aria-label={cell.date}
                />
              ))}
            </div>
          ))}
        </div>
        <div className="mt-2 flex items-center gap-1 text-[10px] text-muted">
          <span>{t("stats.contributionHint")}</span>
          <span>少</span>
          {levelClass.map((c, i) => (
            <span key={i} className={`inline-block h-3 w-3 rounded-[3px] ${c}`} />
          ))}
          <span>多</span>
        </div>
      </div>
    </div>
  );
}
