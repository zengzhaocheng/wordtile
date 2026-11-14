import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deckApi, type DeleteDeckResult } from "../api/endpoints";
import { errText } from "../api/client";
import { useI18n } from "../i18n";
import type { Deck } from "../types";

type DeleteTarget = Pick<Deck, "id" | "name" | "visibility" | "archived">;

interface Props {
  deck: DeleteTarget;
  /** 删除成功后的额外动作（例如详情页返回题库列表） */
  onDeleted?: (res: DeleteDeckResult) => void;
}

/**
 * 题库删除入口。公开题库可勾选「同时从题库市场删除」：
 * 不勾选 → 仅归档，从「我的题库」移除，市场公开版与已 fork 的副本都保留；
 * 勾选   → 彻底删除，已 fork 的副本仍是独立题库，不受影响。
 */
export default function DeleteDeckControl({ deck, onDeleted }: Props) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  // 默认保留市场版本，避免误删已公开的内容。
  const [removeMarket, setRemoveMarket] = useState(false);
  const [error, setError] = useState("");

  const del = useMutation({
    mutationFn: () => deckApi.remove(deck.id, deck.archived ? true : removeMarket),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ["myDecks"] });
      qc.invalidateQueries({ queryKey: ["market"] });
      qc.invalidateQueries({ queryKey: ["deckDetail", deck.id] });
      setOpen(false);
      setError("");
      onDeleted?.(res);
    },
    onError: (e) => setError(`${t("deck.deleteFailed")}: ${errText(e)}`),
  });

  if (!open) {
    return (
      <button
        onClick={() => {
          setError("");
          setOpen(true);
        }}
        className="rounded-lg border border-line px-3 py-1.5 text-xs font-semibold text-danger"
      >
        🗑️ {t("deck.delete")}
      </button>
    );
  }

  return (
    <div className="w-full rounded-xl border border-danger bg-surface p-3 text-sm">
      <p className="font-semibold">{t("deck.deleteConfirm", { name: deck.name })}</p>

      {deck.visibility === "public" && !deck.archived ? (
        <>
          <label className="mt-2 flex cursor-pointer items-start gap-2">
            <input
              type="checkbox"
              checked={removeMarket}
              onChange={(e) => setRemoveMarket(e.target.checked)}
              className="mt-1"
            />
            <span>{t("deck.deleteMarketOption")}</span>
          </label>
          <p className="mt-1 text-xs text-muted">
            {removeMarket ? t("deck.deleteHardHint") : t("deck.deleteKeepMarketHint")}
          </p>
        </>
      ) : (
        <p className="mt-1 text-xs text-muted">
          {deck.archived ? t("deck.deleteArchivedHint") : t("deck.deletePrivateHint")}
        </p>
      )}

      <div className="mt-3 flex flex-wrap gap-2">
        <button
          onClick={() => del.mutate()}
          disabled={del.isPending}
          className="rounded-lg bg-danger px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
        >
          {t("action.confirm")}
        </button>
        <button
          onClick={() => setOpen(false)}
          disabled={del.isPending}
          className="rounded-lg border border-line px-3 py-1.5 text-sm font-semibold"
        >
          {t("action.cancel")}
        </button>
      </div>
      {error && <p className="mt-2 text-xs text-danger">{error}</p>}
    </div>
  );
}
