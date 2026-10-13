import { useNavigate } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { deckApi } from "../api/endpoints";
import { useI18n } from "../i18n";

export default function DecksPage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const qc = useQueryClient();
  const { data: decks, isLoading } = useQuery({ queryKey: ["myDecks"], queryFn: deckApi.listMine });

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">{t("deck.title")}</h1>
        <button
          onClick={() => nav("/decks/new")}
          className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-ink"
        >
          + {t("action.create")}
        </button>
      </div>

      {isLoading && <p className="text-sm text-muted">{t("common.loading")}</p>}
      {decks && decks.length === 0 && <p className="text-sm text-muted">{t("home.noDeck")}</p>}

      <ul className="space-y-3">
        {decks?.map((d) => (
          <li key={d.id}>
            <button
              onClick={() => nav(`/decks/${d.id}`)}
              className="w-full rounded-2xl border border-line bg-surface p-4 text-left active:scale-[0.99]"
            >
              <div className="flex items-center justify-between">
                <span className="font-semibold">{d.name}</span>
                <span className="text-xs text-muted">
                  {d.visibility === "public" ? "🌐" : "🔒"} {d.forked_from ? "· fork" : ""}
                </span>
              </div>
              <p className="mt-1 line-clamp-2 text-sm text-muted">{d.description || "—"}</p>
              <p className="mt-1 text-xs text-muted">
                ❤️ {d.likes_count} · 🍴 {d.forks_count}
              </p>
            </button>
          </li>
        ))}
      </ul>

      {/* 刷新缓存提示：编辑词条后回到本页自动重新拉取 */}
      <button className="hidden" onClick={() => qc.invalidateQueries({ queryKey: ["myDecks"] })} />
    </div>
  );
}
