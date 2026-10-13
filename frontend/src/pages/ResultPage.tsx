import { useNavigate } from "react-router-dom";
import { useGameStore } from "../stores/game";
import { formatDuration, useI18n } from "../i18n";

export default function ResultPage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const { result, session, reset } = useGameStore();

  if (!result) {
    nav("/", { replace: true });
    return null;
  }

  const again = () => {
    const mode = session?.mode ?? "custom";
    reset();
    nav(mode === "campaign" ? "/campaign" : "/play");
  };

  return (
    <div className="mx-auto flex h-full max-w-md flex-col items-center justify-center space-y-6 p-6">
      <h1 className="text-2xl font-bold">🎉 {t("result.title")}</h1>

      <div className="grid w-full grid-cols-3 gap-3">
        <div className="rounded-2xl border border-line bg-surface p-4 text-center">
          <div className="text-2xl font-bold text-primary">{formatDuration(result.duration_ms)}</div>
          <div className="text-xs text-muted">{t("result.time")}</div>
        </div>
        <div className="rounded-2xl border border-line bg-surface p-4 text-center">
          <div className="text-2xl font-bold text-danger">{result.wrong_count}</div>
          <div className="text-xs text-muted">{t("result.wrong")}</div>
        </div>
        <div className="rounded-2xl border border-line bg-surface p-4 text-center">
          <div className="text-2xl font-bold text-success">{Math.round(result.accuracy * 100)}%</div>
          <div className="text-xs text-muted">{t("result.accuracy")}</div>
        </div>
      </div>

      {result.mode === "campaign" && result.level > 0 && (
        <p className="text-sm text-muted">
          {t("campaign.level", { n: result.level })} ·{" "}
          {result.wrong_count <= 3 ? "✅ 已解锁下一关" : "❌ 错误过多，未解锁"}
        </p>
      )}

      <div className="flex w-full gap-3">
        <button
          onClick={again}
          className="flex-1 rounded-xl bg-primary px-4 py-3 font-bold text-primary-ink"
        >
          {t("action.retry")}
        </button>
        <button
          onClick={() => {
            reset();
            nav("/");
          }}
          className="flex-1 rounded-xl border-2 border-line px-4 py-3 font-bold"
        >
          {t("result.backHome")}
        </button>
      </div>
    </div>
  );
}
