import { useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { deckApi, gameApi } from "../api/endpoints";
import { errText } from "../api/client";
import { useGameStore } from "../stores/game";
import { useI18n } from "../i18n";

export default function CustomSetupPage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const [params] = useSearchParams();
  const [deckId, setDeckId] = useState<string>(params.get("deck") ?? "");
  const [blockCount, setBlockCount] = useState(10);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [error, setError] = useState("");

  const { data: decks } = useQuery({ queryKey: ["myDecks"], queryFn: deckApi.listMine });
  const { data: detail } = useQuery({
    queryKey: ["deckDetail", deckId],
    queryFn: () => deckApi.detail(deckId),
    enabled: !!deckId,
  });

  const allTags = useMemo(() => {
    const s = new Set<string>();
    detail?.entries.forEach((e) => e.tags.forEach((tag) => s.add(tag)));
    return [...s].sort();
  }, [detail]);

  const startGame = useGameStore((s) => s.startGame);

  const onStart = async () => {
    if (!deckId) {
      setError(t("play.selectDeck"));
      return;
    }
    try {
      const session = await gameApi.start({
        deck_id: deckId,
        mode: "custom",
        block_count: blockCount,
        tags: selectedTags,
      });
      startGame(session);
      nav("/game");
    } catch (e) {
      setError(errText(e));
    }
  };

  return (
    <div className="mx-auto max-w-2xl space-y-5 p-4">
      <h1 className="text-xl font-bold">{t("play.custom")}</h1>

      {/* 题库选择 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-2 font-semibold">{t("play.selectDeck")}</h2>
        {!decks || decks.length === 0 ? (
          <p className="text-sm text-muted">{t("home.noDeck")}</p>
        ) : (
          <div className="grid gap-2 sm:grid-cols-2">
            {decks.map((d) => (
              <button
                key={d.id}
                onClick={() => {
                  setDeckId(d.id);
                  setSelectedTags([]);
                }}
                className={`rounded-xl border-2 px-3 py-2 text-left text-sm ${
                  deckId === d.id ? "border-primary bg-raised" : "border-line bg-bg"
                }`}
              >
                <div className="font-medium">{d.name}</div>
                <div className="text-xs text-muted">
                  {d.visibility === "public" ? "🌐 " : "🔒 "}
                  {d.forked_from ? "fork" : "original"}
                </div>
              </button>
            ))}
          </div>
        )}
      </section>

      {/* 块数（10 ~ 50） */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-2 font-semibold">
          {t("play.blockCount")}: <span className="text-primary">{blockCount}</span>
        </h2>
        <input
          type="range"
          min={10}
          max={50}
          step={2}
          value={blockCount}
          onChange={(e) => setBlockCount(Number(e.target.value))}
          className="w-full accent-[var(--primary)]"
        />
        <div className="flex justify-between text-xs text-muted">
          <span>10</span>
          <span>50</span>
        </div>
      </section>

      {/* 标签筛选 */}
      {allTags.length > 0 && (
        <section className="rounded-2xl border border-line bg-surface p-4">
          <h2 className="mb-2 font-semibold">{t("play.filterTags")}</h2>
          <div className="flex flex-wrap gap-2">
            {allTags.map((tag) => {
              const on = selectedTags.includes(tag);
              return (
                <button
                  key={tag}
                  onClick={() =>
                    setSelectedTags((prev) => (on ? prev.filter((x) => x !== tag) : [...prev, tag]))
                  }
                  className={`rounded-full border px-3 py-1 text-xs ${
                    on ? "border-primary bg-primary text-primary-ink" : "border-line bg-bg"
                  }`}
                >
                  {tag}
                </button>
              );
            })}
          </div>
        </section>
      )}

      {error && <p className="text-sm text-danger">{error}</p>}

      <button
        onClick={onStart}
        disabled={!deckId}
        className="w-full rounded-xl bg-primary px-4 py-3 font-bold text-primary-ink disabled:opacity-50"
      >
        {t("play.startCustom")}
      </button>
    </div>
  );
}
