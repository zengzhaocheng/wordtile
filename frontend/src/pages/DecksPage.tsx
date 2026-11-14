import { useRef, useState, type ChangeEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { deckApi } from "../api/endpoints";
import { errText } from "../api/client";
import { useI18n } from "../i18n";
import {
  downloadJson,
  parseDeckJson,
  sampleDeckJsonText,
  toDeckJson,
  type DeckJsonFile,
} from "../utils/deckJson";

export default function DecksPage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const qc = useQueryClient();
  const fileRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState("");
  const { data: decks, isLoading } = useQuery({ queryKey: ["myDecks"], queryFn: deckApi.listMine });

  const importNew = useMutation({
    mutationFn: (body: DeckJsonFile) => deckApi.importNewDeck(body),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ["myDecks"] });
      nav(`/decks/${res.deck.id}`);
    },
    onError: (e) => setError(`${t("deck.importFailed")}: ${errText(e)}`),
  });

  const onPickFile = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = ""; // 允许连续选择同一个文件
    if (!file) return;
    setError("");
    try {
      const parsed = parseDeckJson(await file.text());
      importNew.mutate(toDeckJson(parsed, file.name.replace(/\.json$/i, "")));
    } catch (err) {
      setError(`${t("deck.importFailed")}: ${(err as Error).message}`);
    }
  };

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-xl font-bold">{t("deck.title")}</h1>
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => fileRef.current?.click()}
            disabled={importNew.isPending}
            className="rounded-lg border-2 border-primary px-3 py-2 text-sm font-semibold text-primary disabled:opacity-50"
          >
            ⬆️ {t("deck.importNew")}
          </button>
          <button
            onClick={() => nav("/decks/new")}
            className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-ink"
          >
            + {t("action.create")}
          </button>
        </div>
      </div>

      {/* 题库市场入口：功能一直存在，这里补一个显眼的入口 */}
      <button
        onClick={() => nav("/market")}
        className="flex w-full items-center justify-between gap-3 rounded-2xl border-2 border-primary bg-surface p-4 text-left active:scale-[0.99]"
      >
        <div className="min-w-0">
          <p className="font-semibold">🌐 {t("deck.marketEntry")}</p>
          <p className="mt-1 text-xs text-muted">{t("deck.marketEntryDesc")}</p>
        </div>
        <span className="shrink-0 rounded-lg bg-primary px-3 py-1.5 text-sm font-semibold text-primary-ink">
          {t("deck.marketGo")} →
        </span>
      </button>

      {/* 导入说明 */}
      <section className="rounded-2xl border border-line bg-surface p-4 text-sm">
        <p className="font-semibold">📄 {t("deck.importNewHint")}</p>
        <p className="mt-1 text-xs text-muted">{t("deck.importTip")}</p>
        <button
          onClick={() => downloadJson(JSON.parse(sampleDeckJsonText()), "wordtile-deck-template.json")}
          className="mt-2 rounded-lg border border-line px-3 py-1.5 text-xs font-semibold text-primary"
        >
          ⬇️ {t("deck.importTemplate")}
        </button>
        <input
          ref={fileRef}
          type="file"
          accept=".json,application/json"
          onChange={onPickFile}
          className="hidden"
        />
      </section>

      {error && <p className="text-sm text-danger">{error}</p>}
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
    </div>
  );
}
