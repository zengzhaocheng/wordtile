import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { deckApi, marketApi } from "../api/endpoints";
import { errText } from "../api/client";
import { useI18n } from "../i18n";
import { useAuthStore } from "../stores/auth";
import DeleteDeckControl from "../components/DeleteDeckControl";

export default function MarketDeckPage() {
  const { id } = useParams();
  const { t } = useI18n();
  const nav = useNavigate();
  const qc = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const [error, setError] = useState("");
  const [comment, setComment] = useState("");
  const [notice, setNotice] = useState("");
  const [existingForkId, setExistingForkId] = useState("");

  const { data: detail } = useQuery({
    queryKey: ["deckDetail", id],
    queryFn: () => deckApi.detail(id!),
    enabled: !!id,
  });
  const { data: comments } = useQuery({
    queryKey: ["comments", id],
    queryFn: () => marketApi.comments(id!),
    enabled: !!id,
  });

  const like = useMutation({
    mutationFn: () => marketApi.like(id!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["deckDetail", id] });
    },
    onError: (e) => setError(errText(e)),
  });

  const fork = useMutation({
    mutationFn: () => marketApi.fork(id!),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ["myDecks"] });
      if (res.existing) {
        // 幂等：同一用户对同一来源只保留一个副本，不重复创建。
        setExistingForkId(res.deck.id);
        setNotice("ℹ️ " + t("deck.forkExisting"));
        return;
      }
      nav(`/decks/${res.deck.id}`);
    },
    onError: (e) => setError(errText(e)),
  });

  const addComment = useMutation({
    mutationFn: () => marketApi.addComment(id!, comment),
    onSuccess: () => {
      setComment("");
      qc.invalidateQueries({ queryKey: ["comments", id] });
    },
    onError: (e) => setError(errText(e)),
  });

  if (!detail) return <p className="p-4 text-sm text-muted">{t("common.loading")}</p>;
  const deck = detail.deck;

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-bold">
            {deck.name}
            {deck.archived && (
              <span className="ml-1 text-xs font-normal text-muted">
                · {t("deck.archivedBadge")}
              </span>
            )}
          </h1>
          <p className="text-sm text-muted">{deck.description || "—"}</p>
          <p className="mt-1 text-xs text-muted">
            ❤️ {deck.likes_count} · 🍴 {deck.forks_count} · 📚 {detail.entries.length} 词条
          </p>
        </div>
        <button onClick={() => nav("/market")} className="text-sm text-muted underline">
          {t("action.back")}
        </button>
      </div>

      <div className="flex gap-2">
        <button
          onClick={() => like.mutate()}
          disabled={like.isPending}
          className="rounded-lg border-2 border-danger px-4 py-2 text-sm font-semibold text-danger disabled:opacity-50"
        >
          ❤️ {t("action.like")}
        </button>
        <button
          onClick={() => {
            setNotice("");
            setExistingForkId("");
            fork.mutate();
          }}
          disabled={fork.isPending}
          className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-ink disabled:opacity-50"
        >
          🍴 {t("action.fork")}
        </button>
      </div>
      {notice && (
        <div className="rounded-xl border border-primary bg-surface p-3 text-sm">
          <p>{notice}</p>
          {existingForkId && (
            <button
              onClick={() => nav(`/decks/${existingForkId}`)}
              className="mt-2 rounded-lg bg-primary px-3 py-1.5 text-sm font-semibold text-primary-ink"
            >
              {t("deck.forkOpenExisting")}
            </button>
          )}
        </div>
      )}
      {error && <p className="text-sm text-danger">{error}</p>}

      {/* 词条预览 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-2 font-semibold">{t("deck.entries")}</h2>
        <ul className="grid gap-1 text-sm sm:grid-cols-2">
          {detail.entries.slice(0, 20).map((e) => (
            <li key={e.id} className="truncate">
              {e.term} <span className="text-muted">→</span> {e.definition}
            </li>
          ))}
          {detail.entries.length > 20 && (
            <li className="text-muted">… {detail.entries.length - 20}+</li>
          )}
        </ul>
      </section>

      {/* 评论 */}
      <section className="space-y-2">
        <h2 className="font-semibold">{t("market.comments")}</h2>
        <div className="flex gap-2">
          <input
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder={t("market.commentPlaceholder")}
            className="flex-1 rounded-lg border border-line bg-surface px-3 py-2 text-sm"
          />
          <button
            onClick={() => addComment.mutate()}
            disabled={addComment.isPending || !comment}
            className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-ink disabled:opacity-50"
          >
            {t("action.submit")}
          </button>
        </div>
        {comments && comments.length === 0 && (
          <p className="text-sm text-muted">{t("market.noComments")}</p>
        )}
        <ul className="space-y-2">
          {comments?.map((c) => (
            <li key={c.id} className="rounded-xl bg-raised px-3 py-2 text-sm">
              <span className="font-semibold">{c.author_name}</span>{" "}
              <span className="text-xs text-muted">
                {new Date(c.created_at).toLocaleDateString()}
              </span>
              <p>{c.content}</p>
            </li>
          ))}
        </ul>
      </section>

      {/* 市场保留版的原作者仍可彻底删除 */}
      {deck.archived && user?.id === deck.owner_id && (
        <section className="rounded-2xl border border-danger bg-surface p-4">
          <h2 className="mb-2 text-sm font-semibold text-danger">🗑️ {t("deck.delete")}</h2>
          <DeleteDeckControl deck={deck} onDeleted={() => nav("/market")} />
        </section>
      )}
    </div>
  );
}
