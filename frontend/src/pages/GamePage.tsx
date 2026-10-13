import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { gameApi } from "../api/endpoints";
import { errText } from "../api/client";
import { useGameStore } from "../stores/game";
import { useI18n } from "../i18n";
import TileBlock, { type TileState } from "../components/TileBlock";

/** 闯关模式关卡规则：第 n 关 6+2(n-1) 对（上限 30），最多允许 3 次错误 */
export const CAMPAIGN_MAX_ERRORS = 3;
export function levelPairs(level: number): number {
  return Math.min(6 + (level - 1) * 2, 30);
}
export const CAMPAIGN_LEVELS = 13;

export default function GamePage() {
  const { t } = useI18n();
  const nav = useNavigate();
  const {
    session,
    startedAt,
    selectedTerm,
    matchedTermTokens,
    matchedDefTokens,
    wrongPair,
    hintPair,
    correctCount,
    wrongCount,
    selectTerm,
    applyMatch,
    setWrong,
    clearWrong,
    setHint,
    setResult,
    reset,
  } = useGameStore();

  const [elapsed, setElapsed] = useState(0);
  const [error, setError] = useState("");
  const finishingRef = useRef(false);
  const [hintTimer, setHintTimer] = useState<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!session) {
      nav("/play", { replace: true });
      return;
    }
    const id = setInterval(() => setElapsed(Date.now() - startedAt), 250);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => () => {
    if (hintTimer) clearTimeout(hintTimer);
  }, [hintTimer]);

  if (!session) return null;

  const matchedTermSet = new Set(matchedTermTokens);
  const matchedDefSet = new Set(matchedDefTokens);

  const finish = async () => {
    if (finishingRef.current) return;
    finishingRef.current = true;
    try {
      const summary = await gameApi.finish(session.session_id, Date.now() - startedAt);
      setResult(summary);
      nav("/result");
    } catch (e) {
      setError(errText(e));
      finishingRef.current = false;
    }
  };

  const onDefinitionClick = async (defToken: string) => {
    if (!selectedTerm || wrongPair || finishingRef.current) return;
    const termToken = selectedTerm;
    try {
      const res = await gameApi.match(session.session_id, termToken, defToken);
      if (res.correct) {
        applyMatch(termToken, defToken, res);
        if (res.finished) await finish();
      } else {
        setWrong(termToken, defToken);
        setTimeout(clearWrong, 1000); // 错误红色反馈 1 秒
      }
    } catch (e) {
      setError(errText(e));
    }
  };

  const onHint = async () => {
    try {
      const pair = await gameApi.hint(session.session_id);
      setHint({ term: pair.term_token, def: pair.definition_token });
      const timer = setTimeout(() => setHint(null), 1500);
      setHintTimer(timer);
    } catch (e) {
      setError(errText(e));
    }
  };

  const termState = (token: string): TileState => {
    if (matchedTermSet.has(token)) return "matched";
    if (wrongPair?.term === token) return "wrong";
    if (hintPair?.term === token) return "hint";
    if (selectedTerm === token) return "selected";
    return "normal";
  };
  const defState = (token: string): TileState => {
    if (matchedDefSet.has(token)) return "matched";
    if (wrongPair?.def === token) return "wrong";
    if (hintPair?.def === token) return "hint";
    return "normal";
  };

  const exit = () => {
    reset();
    nav(session.mode === "campaign" ? "/campaign" : "/play");
  };

  return (
    <div className="flex h-full flex-col">
      {/* 顶部状态栏 */}
      <div className="flex items-center justify-between border-b border-line bg-surface px-4 py-2 text-sm">
        <button onClick={exit} className="text-muted underline">
          {t("game.exit")}
        </button>
        <div className="flex gap-4">
          <span>
            {t("game.progress")}: {matchedTermTokens.length}/{session.block_count}
          </span>
          <span>
            {t("game.time")}: {(elapsed / 1000).toFixed(1)}s
          </span>
          <span className="text-danger">
            {t("game.wrong")}: {wrongCount}
          </span>
        </div>
        <button onClick={onHint} className="rounded-full border border-line px-3 py-1 text-xs">
          💡 {t("action.hint")}
        </button>
      </div>
      <p className="px-4 pt-2 text-center text-xs text-muted">{t("game.selectTerm")}</p>
      {error && <p className="px-4 text-center text-xs text-danger">{error}</p>}

      {/* 上屏：词条 */}
      <div className="min-h-0 flex-1 overflow-y-auto p-3">
        <div className="grid grid-cols-2 content-start gap-2 sm:grid-cols-3 lg:grid-cols-4">
          {session.terms.map((tile) => (
            <TileBlock
              key={tile.token}
              text={tile.text}
              state={termState(tile.token)}
              onClick={() => selectTerm(tile.token)}
            />
          ))}
        </div>
      </div>

      {/* 分隔条 */}
      <div className="flex items-center justify-center py-1">
        <span className="h-1 w-16 rounded-full bg-line" aria-hidden />
      </div>

      {/* 下屏：释义 */}
      <div className="min-h-0 flex-1 overflow-y-auto p-3">
        <div className="grid grid-cols-2 content-start gap-2 sm:grid-cols-3 lg:grid-cols-4">
          {session.definitions.map((tile) => (
            <TileBlock
              key={tile.token}
              text={tile.text}
              state={defState(tile.token)}
              onClick={() => onDefinitionClick(tile.token)}
            />
          ))}
        </div>
      </div>

      <span className="hidden" data-testid="correct-count">
        {correctCount}
      </span>
    </div>
  );
}
