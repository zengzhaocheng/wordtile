import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { marketApi } from "../api/endpoints";
import { useI18n } from "../i18n";

type SortKey = "likes" | "newest" | "forks";

export default function MarketplacePage() {
  const { t } = useI18n();
  const [search, setSearch] = useState("");
  const [sort, setSort] = useState<SortKey>("likes");

  const { data: decks, isLoading } = useQuery({
    queryKey: ["market", search, sort],
    queryFn: () => marketApi.list(search, sort),
  });

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <h1 className="text-xl font-bold">🌐 {t("market.title")}</h1>

      <div className="flex gap-2">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t("market.search")}
          className="flex-1 rounded-lg border border-line bg-surface px-3 py-2 text-sm"
        />
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as SortKey)}
          className="rounded-lg border border-line bg-surface px-2 py-2 text-sm"
        >
          <option value="likes">{t("market.sortLikes")}</option>
          <option value="newest">{t("market.sortNewest")}</option>
          <option value="forks">{t("market.sortForks")}</option>
        </select>
      </div>

      {isLoading && <p className="text-sm text-muted">{t("common.loading")}</p>}
      {decks && decks.length === 0 && <p className="text-sm text-muted">{t("market.empty")}</p>}

      <ul className="space-y-3">
        {decks?.map((d) => (
          <li key={d.id}>
            <Link
              to={`/market/${d.id}`}
              className="block rounded-2xl border border-line bg-surface p-4 active:scale-[0.99]"
            >
              <div className="flex items-center justify-between">
                <span className="font-semibold">{d.name}</span>
                <span className="text-xs text-muted">{d.owner_name}</span>
              </div>
              <p className="mt-1 line-clamp-2 text-sm text-muted">{d.description || "—"}</p>
              <p className="mt-1 text-xs text-muted">
                ❤️ {d.likes_count} · 🍴 {d.forks_count}
              </p>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
