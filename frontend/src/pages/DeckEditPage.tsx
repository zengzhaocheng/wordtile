import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { deckApi, versionApi } from "../api/endpoints";
import { errText } from "../api/client";
import { useI18n } from "../i18n";
import type { Entry } from "../types";

export default function DeckEditPage() {
  const { id } = useParams();
  const isNew = !id || id === "new";
  const { t } = useI18n();
  const nav = useNavigate();
  const qc = useQueryClient();
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  // 新建表单状态
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  // 新词条表单
  const [term, setTerm] = useState("");
  const [definition, setDefinition] = useState("");
  const [tags, setTags] = useState("");

  const { data: detail } = useQuery({
    queryKey: ["deckDetail", id],
    queryFn: () => deckApi.detail(id!),
    enabled: !isNew && !!id,
  });
  const { data: commits } = useQuery({
    queryKey: ["commits", id],
    queryFn: () => versionApi.commits(id!),
    enabled: !isNew && !!id,
  });

  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ["deckDetail", id] });
    qc.invalidateQueries({ queryKey: ["commits", id] });
    qc.invalidateQueries({ queryKey: ["myDecks"] });
  };

  const createDeck = useMutation({
    mutationFn: () => deckApi.create({ name, description }),
    onSuccess: (deck) => {
      qc.invalidateQueries({ queryKey: ["myDecks"] });
      nav(`/decks/${deck.id}`);
    },
    onError: (e) => setError(errText(e)),
  });

  const addEntry = useMutation({
    mutationFn: () =>
      deckApi.createEntry(id!, {
        term,
        definition,
        tags: tags.split(/[,，]/).map((s) => s.trim()).filter(Boolean),
      }),
    onSuccess: () => {
      setTerm("");
      setDefinition("");
      setTags("");
      invalidate();
    },
    onError: (e) => setError(errText(e)),
  });

  const deleteEntry = useMutation({
    mutationFn: (entryId: string) => deckApi.deleteEntry(entryId),
    onSuccess: invalidate,
    onError: (e) => setError(errText(e)),
  });

  const commit = useMutation({
    mutationFn: (message: string) => versionApi.commit(id!, message),
    onSuccess: () => {
      setNotice("✅ " + t("action.commit"));
      invalidate();
    },
    onError: (e) => setError(errText(e)),
  });

  const publish = useMutation({
    mutationFn: () => deckApi.update(id!, { visibility: "public" }),
    onSuccess: () => {
      setNotice("✅ " + t("action.publish"));
      invalidate();
    },
    onError: (e) => setError(errText(e)),
  });

  const pull = useMutation({
    mutationFn: () => versionApi.merge(id!, {}),
    onSuccess: (res) => {
      if (res.status === "up_to_date") setNotice("✅ 已是最新");
      else if (res.status === "merged") setNotice(`✅ 已合并（${res.merged_count} 词条）`);
      else setNotice(`⚠️ ${res.conflicts} 条冲突待解决`);
      invalidate();
    },
    onError: (e) => setError(errText(e)),
  });

  if (isNew) {
    return (
      <div className="mx-auto max-w-xl space-y-4 p-4">
        <h1 className="text-xl font-bold">{t("deck.createTitle")}</h1>
        {error && <p className="text-sm text-danger">{error}</p>}
        <label className="block text-sm">
          {t("deck.name")}
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2"
          />
        </label>
        <label className="block text-sm">
          {t("deck.description")}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2"
          />
        </label>
        <button
          onClick={() => createDeck.mutate()}
          disabled={createDeck.isPending || !name}
          className="w-full rounded-xl bg-primary px-4 py-3 font-bold text-primary-ink disabled:opacity-50"
        >
          {t("action.create")}
        </button>
      </div>
    );
  }

  if (!detail) {
    return <p className="p-4 text-sm text-muted">{t("common.loading")}</p>;
  }

  const deck = detail.deck;

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      {/* 头部 */}
      <div className="flex items-start justify-between gap-2">
        <div>
          <h1 className="text-xl font-bold">
            {deck.name}
            {deck.forked_from && <span className="ml-1 text-sm text-muted">(fork)</span>}
          </h1>
          <p className="text-sm text-muted">{deck.description || "—"}</p>
        </div>
        <button onClick={() => nav("/decks")} className="text-sm text-muted underline">
          {t("action.back")}
        </button>
      </div>

      {/* 操作区 */}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => {
            const msg = window.prompt("Commit message:", "更新词条");
            if (msg !== null) commit.mutate(msg);
          }}
          className="rounded-lg bg-primary px-3 py-2 text-sm font-semibold text-primary-ink"
        >
          📦 {t("action.commit")}
        </button>
        {deck.visibility !== "public" && (
          <button
            onClick={() => publish.mutate()}
            className="rounded-lg border-2 border-primary px-3 py-2 text-sm font-semibold text-primary"
          >
            🌐 {t("action.publish")}
          </button>
        )}
        {deck.forked_from && (
          <button
            onClick={() => pull.mutate()}
            className="rounded-lg border-2 border-line px-3 py-2 text-sm font-semibold"
          >
            ⬇️ {t("action.pull")}
          </button>
        )}
        {detail.open_conflicts > 0 && (
          <button
            onClick={() => nav(`/decks/${id}/merge`)}
            className="rounded-lg bg-danger px-3 py-2 text-sm font-semibold text-white"
          >
            ⚠️ {t("deck.conflictBadge", { n: detail.open_conflicts })}
          </button>
        )}
      </div>
      {notice && <p className="text-sm text-success">{notice}</p>}
      {error && <p className="text-sm text-danger">{error}</p>}
      {deck.head_commit_id && <p className="text-xs text-muted">{t("deck.uncommitted")}</p>}

      {/* 新增词条 */}
      <section className="rounded-2xl border border-line bg-surface p-4">
        <h2 className="mb-3 font-semibold">{t("deck.addEntry")}</h2>
        <div className="grid gap-2 sm:grid-cols-3">
          <input
            value={term}
            onChange={(e) => setTerm(e.target.value)}
            placeholder={t("deck.term")}
            className="rounded-lg border border-line bg-bg px-3 py-2 text-sm"
          />
          <input
            value={definition}
            onChange={(e) => setDefinition(e.target.value)}
            placeholder={t("deck.definition")}
            className="rounded-lg border border-line bg-bg px-3 py-2 text-sm"
          />
          <input
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder={t("deck.tags")}
            className="rounded-lg border border-line bg-bg px-3 py-2 text-sm"
          />
        </div>
        <button
          onClick={() => addEntry.mutate()}
          disabled={addEntry.isPending || !term || !definition}
          className="mt-2 rounded-lg bg-primary px-3 py-2 text-sm font-semibold text-primary-ink disabled:opacity-50"
        >
          + {t("action.save")}
        </button>
      </section>

      {/* 词条列表 */}
      <section className="space-y-2">
        <h2 className="font-semibold">
          {t("deck.entries")}（{detail.entries.length}）
        </h2>
        {detail.entries.map((e: Entry) => (
          <EntryRow key={e.id} entry={e} onDelete={() => deleteEntry.mutate(e.id)} />
        ))}
      </section>

      {/* 提交历史 */}
      {commits && commits.length > 0 && (
        <section className="space-y-2">
          <h2 className="font-semibold">{t("deck.history")}</h2>
          <ul className="space-y-1 text-sm">
            {commits
              .slice()
              .reverse()
              .map((c) => (
                <li key={c.id} className="rounded-lg bg-raised px-3 py-2">
                  <span className="font-mono text-xs text-muted">{c.hash.slice(0, 8)}</span>{" "}
                  {c.message}
                  <span className="ml-2 text-xs text-muted">
                    {new Date(c.created_at).toLocaleString()}
                  </span>
                </li>
              ))}
          </ul>
        </section>
      )}
    </div>
  );
}

function EntryRow({ entry, onDelete }: { entry: Entry; onDelete: () => void }) {
  const { t } = useI18n();
  const [editing, setEditing] = useState(false);
  const [term, setTerm] = useState(entry.term);
  const [definition, setDefinition] = useState(entry.definition);
  const [tags, setTags] = useState(entry.tags.join(", "));
  const qc = useQueryClient();

  const save = useMutation({
    mutationFn: () =>
      deckApi.updateEntry(entry.id, {
        term,
        definition,
        tags: tags.split(/[,，]/).map((s) => s.trim()).filter(Boolean),
      }),
    onSuccess: () => {
      setEditing(false);
      qc.invalidateQueries({ queryKey: ["deckDetail", entry.deck_id] });
    },
  });

  return (
    <div className="rounded-xl border border-line bg-surface p-3 text-sm">
      {editing ? (
        <div className="space-y-2">
          <input
            value={term}
            onChange={(e) => setTerm(e.target.value)}
            className="w-full rounded-lg border border-line bg-bg px-2 py-1.5"
          />
          <input
            value={definition}
            onChange={(e) => setDefinition(e.target.value)}
            className="w-full rounded-lg border border-line bg-bg px-2 py-1.5"
          />
          <input
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            className="w-full rounded-lg border border-line bg-bg px-2 py-1.5"
          />
          <div className="flex gap-2">
            <button
              onClick={() => save.mutate()}
              className="rounded-lg bg-primary px-3 py-1.5 text-xs font-semibold text-primary-ink"
            >
              {t("action.save")}
            </button>
            <button
              onClick={() => setEditing(false)}
              className="rounded-lg border border-line px-3 py-1.5 text-xs"
            >
              {t("action.cancel")}
            </button>
          </div>
        </div>
      ) : (
        <div className="flex items-center justify-between gap-2">
          <div className="min-w-0">
            <span className="font-medium">{entry.term}</span>
            <span className="mx-1 text-muted">→</span>
            <span>{entry.definition}</span>
            {entry.tags.length > 0 && (
              <span className="ml-2 text-xs text-muted">#{entry.tags.join(" #")}</span>
            )}
          </div>
          <div className="flex shrink-0 gap-2 text-xs">
            <button onClick={() => setEditing(true)} className="text-primary underline">
              {t("action.edit")}
            </button>
            <button onClick={onDelete} className="text-danger underline">
              {t("action.delete")}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
