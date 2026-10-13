import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { deckApi, gameApi, statsApi } from "../api/endpoints";
import { errText } from "../api/client";
import { formatDuration, useI18n } from "../i18n";
import { useGameStore } from "../stores/game";
import { CAMPAIGN_LEVELS, CAMPAIGN_MAX_ERRORS, levelPairs } from "./GamePage";

export default function CampaignPage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const [error, setError] = useState("");

  const { data: decks } = useQuery({ queryKey: ["myDecks"], queryFn: deckApi.listMine });
  const { data: levels } = useQuery({ queryKey: ["campaignStats"], queryFn: statsApi.campaign });
  const startGame = useGameStore((s) => s.startGame);

  // 通关判定：某关最佳成绩的 wrong ≤ 允许错误数 → 解锁下一关
  const unlockedUpTo = (() => {
    let unlocked = 1;
    for (const s of levels ?? []) {
      if (s.best_wrong <= CAMPAIGN_MAX_ERRORS && s.level >= unlocked) {
        unlocked = Math.max(unlocked, s.level + 1);
      }
    }
    return Math.min(unlocked, CAMPAIGN_LEVELS);
  })();

  const play = async (level: number) => {
    const deck = decks?.find((d) => !d.forked_from) ?? decks?.[0];
    if (!deck) {
      setError(t("home.noDeck"));
      return;
    }
    try {
      const session = await gameApi.start({
        deck_id: deck.id,
        mode: "campaign",
        level,
        block_count: levelPairs(level),
      });
      startGame(session);
      nav("/game");
    } catch (e) {
      setError(errText(e));
    }
  };

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <h1 className="text-xl font-bold">🏆 {t("campaign.title")}</h1>
      <p className="text-sm text-muted">
        {t("campaign.maxErrors", { n: CAMPAIGN_MAX_ERRORS })} · {t("play.selectDeck")}:{" "}
        {decks?.[0]?.name ?? "—"}
      </p>
      {error && <p className="text-sm text-danger">{error}</p>}

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {Array.from({ length: CAMPAIGN_LEVELS }, (_, i) => i + 1).map((level) => {
          const unlocked = level <= unlockedUpTo;
          const stat = levels?.find((s) => s.level === level);
          return (
            <button
              key={level}
              disabled={!unlocked}
              onClick={() => play(level)}
              className={`rounded-2xl border-2 p-4 text-left ${
                unlocked ? "border-primary bg-surface active:scale-95" : "border-line bg-raised opacity-60"
              }`}
            >
              <div className="font-bold">{t("campaign.level", { n: level })}</div>
              <div className="text-xs text-muted">{t("campaign.pairs", { n: levelPairs(level) })}</div>
              {stat ? (
                <div className="mt-1 text-xs text-success">
                  {t("campaign.best", {
                    time: formatDuration(stat.best_ms),
                    wrong: stat.best_wrong,
                  })}
                </div>
              ) : (
                <div className="mt-1 text-xs text-muted">
                  {unlocked ? t("campaign.unlocked") : `🔒 ${t("campaign.locked")}`}
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
