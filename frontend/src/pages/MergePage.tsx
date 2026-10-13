import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { versionApi } from "../api/endpoints";
import { errText } from "../api/client";
import { useI18n } from "../i18n";
import type { Conflict, EntrySnapshot } from "../types";

export default function MergePage() {
  const { id } = useParams();
  const { t } = useI18n();
  const nav = useNavigate();
  const qc = useQueryClient();
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const { data: conflicts, isLoading } = useQuery({
    queryKey: ["conflicts", id],
    queryFn: () => versionApi.conflicts(id!, "open"),
    enabled: !!id,
  });

  const pull = useMutation({
    mutationFn: () => versionApi.merge(id!, {}),
    onSuccess: (res) => {
      if (res.status === "up_to_date") setNotice("✅ 已是最新");
      else if (res.status === "merged") setNotice(`✅ 已自动合并（${res.merged_count} 词条）`);
      else setNotice(`⚠️ ${res.conflicts} 条冲突待解决`);
      qc.invalidateQueries({ queryKey: ["conflicts", id] });
      qc.invalidateQueries({ queryKey: ["deckDetail", id] });
    },
    onError: (e) => setError(errText(e)),
  });

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">🔀 {t("merge.title")}</h1>
        <button onClick={() => nav(`/decks/${id}`)} className="text-sm text-muted underline">
          {t("action.back")}
        </button>
      </div>

      <button
        onClick={() => pull.mutate()}
        disabled={pull.isPending}
        className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-ink disabled:opacity-50"
      >
        ⬇️ {t("action.pull")}
      </button>

      {notice && <p className="text-sm text-success">{notice}</p>}
      {error && <p className="text-sm text-danger">{error}</p>}

      {isLoading && <p className="text-sm text-muted">{t("common.loading")}</p>}
      {conflicts && conflicts.length === 0 && <p className="text-sm text-muted">{t("merge.noConflicts")}</p>}

      <div className="space-y-4">
        {conflicts?.map((cf) => (
          <ConflictCard
            key={cf.id}
            deckId={id!}
            conflict={cf}
            onResolved={(remaining, autoCommitted) => {
              if (remaining === 0 && autoCommitted) {
                setNotice(`✅ ${t("merge.allResolved")}`);
              }
              qc.invalidateQueries({ queryKey: ["conflicts", id] });
              qc.invalidateQueries({ queryKey: ["deckDetail", id] });
              qc.invalidateQueries({ queryKey: ["commits", id] });
            }}
          />
        ))}
      </div>
    </div>
  );
}

function ConflictCard({
  deckId,
  conflict,
  onResolved,
}: {
  deckId: string;
  conflict: Conflict;
  onResolved: (remaining: number, autoCommitted: boolean) => void;
}) {
  const { t } = useI18n();
  const [customOpen, setCustomOpen] = useState(false);
  const [term, setTerm] = useState(conflict.ours.term || conflict.theirs.term);
  const [definition, setDefinition] = useState(
    conflict.ours.definition || conflict.theirs.definition,
  );
  const [tags, setTags] = useState((conflict.ours.tags ?? conflict.theirs.tags ?? []).join(", "));
  const [error, setError] = useState("");

  const resolve = useMutation({
    mutationFn: (body: Parameters<typeof versionApi.resolve>[1]) => versionApi.resolve(deckId, body),
    onSuccess: (res) => onResolved(res.remaining, res.auto_committed),
    onError: (e) => setError(errText(e)),
  });

  const snapView = (label: string, snap: EntrySnapshot, tone: string) => (
    <div className={`rounded-xl border p-3 text-sm ${tone}`}>
      <div className="mb-1 text-xs font-semibold text-muted">{label}</div>
      {snap.exists ? (
        <>
          <div className="font-medium">{snap.term}</div>
          <div>{snap.definition}</div>
          {snap.tags?.length > 0 && <div className="text-xs text-muted">#{snap.tags.join(" #")}</div>}
        </>
      ) : (
        <div className="text-muted">∅（已删除）</div>
      )}
    </div>
  );

  return (
    <div className="rounded-2xl border border-line bg-surface p-4">
      <div className="grid gap-2 sm:grid-cols-3">
        {snapView(t("merge.base"), conflict.base, "border-line bg-bg")}
        {snapView(t("merge.ours"), conflict.ours, "border-primary/50 bg-bg")}
        {snapView(t("merge.theirs"), conflict.theirs, "border-success/50 bg-bg")}
      </div>

      <div className="mt-3 flex flex-wrap gap-2">
        <button
          onClick={() =>
            resolve.mutate({ conflict_id: conflict.id, resolution: "ours" })
          }
          disabled={resolve.isPending}
          className="rounded-lg border-2 border-primary px-3 py-1.5 text-sm font-semibold text-primary disabled:opacity-50"
        >
          {t("merge.useOurs")}
        </button>
        <button
          onClick={() =>
            resolve.mutate({ conflict_id: conflict.id, resolution: "theirs" })
          }
          disabled={resolve.isPending}
          className="rounded-lg border-2 border-success px-3 py-1.5 text-sm font-semibold text-success disabled:opacity-50"
        >
          {t("merge.useTheirs")}
        </button>
        <button
          onClick={() => setCustomOpen(!customOpen)}
          className="rounded-lg border-2 border-line px-3 py-1.5 text-sm font-semibold"
        >
          {t("merge.custom")}
        </button>
      </div>

      {customOpen && (
        <div className="mt-3 space-y-2 rounded-xl bg-raised p-3">
          <input
            value={term}
            onChange={(e) => setTerm(e.target.value)}
            placeholder={t("deck.term")}
            className="w-full rounded-lg border border-line bg-bg px-3 py-2 text-sm"
          />
          <input
            value={definition}
            onChange={(e) => setDefinition(e.target.value)}
            placeholder={t("deck.definition")}
            className="w-full rounded-lg border border-line bg-bg px-3 py-2 text-sm"
          />
          <input
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder={t("deck.tags")}
            className="w-full rounded-lg border border-line bg-bg px-3 py-2 text-sm"
          />
          <button
            onClick={() =>
              resolve.mutate({
                conflict_id: conflict.id,
                resolution: "custom",
                custom: {
                  exists: true,
                  term,
                  definition,
                  tags: tags.split(/[,，]/).map((s) => s.trim()).filter(Boolean),
                },
              })
            }
            disabled={resolve.isPending || !term || !definition}
            className="rounded-lg bg-primary px-3 py-1.5 text-sm font-semibold text-primary-ink disabled:opacity-50"
          >
            {t("action.submit")}
          </button>
        </div>
      )}
      {error && <p className="mt-2 text-sm text-danger">{error}</p>}
    </div>
  );
}
