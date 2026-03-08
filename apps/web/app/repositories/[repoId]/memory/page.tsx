"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";

import { MemoryDetailDrawer } from "@/components/memory/memory-detail-drawer";
import { MemoryFilters } from "@/components/memory/memory-filters";
import { MemoryTimeline } from "@/components/memory/memory-timeline";
import { useMemoryDetail } from "@/lib/hooks/use-memory-detail";
import { useRepositoryDetail } from "@/lib/hooks/use-repository-detail";
import { useRepositoryMemory } from "@/lib/hooks/use-repository-memory";

export default function RepositoryMemoryPage() {
  const params = useParams<{ repoId: string }>();
  const repoId = params.repoId;
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [activeMemoryID, setActiveMemoryID] = useState<string | null>(null);

  const repositoryQuery = useRepositoryDetail(repoId);
  const memoryQuery = useRepositoryMemory(repoId);
  const memoryDetailQuery = useMemoryDetail(repoId, activeMemoryID);

  const availableTypes = useMemo(() => {
    const set = new Set((memoryQuery.data?.memoryEntries ?? []).map((entry) => entry.type));
    return Array.from(set.values()).sort();
  }, [memoryQuery.data?.memoryEntries]);

  const filteredEntries = useMemo(() => {
    const entries = memoryQuery.data?.memoryEntries ?? [];
    if (typeFilter === "all") return entries;
    return entries.filter((entry) => entry.type === typeFilter);
  }, [memoryQuery.data?.memoryEntries, typeFilter]);

  return (
    <section className="space-y-6">
      <header className="rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50 p-6 shadow-sm">
        <p className="text-xs uppercase tracking-[0.18em] text-slate-500">Repository Memory Timeline</p>
        <h2 className="mt-2 text-3xl font-semibold tracking-tight text-slate-900">
          {repositoryQuery.data?.repository.fullName ?? "Loading repository..."}
        </h2>
        <p className="mt-2 text-sm text-slate-600">Browse persisted memory entries generated from synced pull requests and issues.</p>
        <div className="mt-4 flex flex-wrap gap-3">
          <Link href={`/repositories/${repoId}`} className="rounded-md border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:border-slate-400">
            Back to repository
          </Link>
          {repositoryQuery.data?.repository.htmlUrl ? (
            <a href={repositoryQuery.data.repository.htmlUrl} target="_blank" rel="noreferrer" className="rounded-md border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:border-slate-400">
              Open on GitHub
            </a>
          ) : null}
        </div>
      </header>

      <MemoryFilters selectedType={typeFilter} availableTypes={availableTypes} onTypeChange={setTypeFilter} />

      {memoryQuery.isLoading ? (
        <div className="space-y-4" aria-label="loading memory entries">
          <div className="h-36 animate-pulse rounded-2xl border border-slate-200 bg-white" />
          <div className="h-36 animate-pulse rounded-2xl border border-slate-200 bg-white" />
        </div>
      ) : null}

      {memoryQuery.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          Failed to load repository memory timeline.
        </div>
      ) : null}

      {!memoryQuery.isLoading && !memoryQuery.error && (memoryQuery.data?.memoryEntries.length ?? 0) === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
          <h3 className="text-xl font-semibold text-slate-900">No memory entries yet</h3>
          <p className="mt-2 text-sm text-slate-600">Trigger repository sync and memory generation to populate this timeline.</p>
        </div>
      ) : null}

      {!memoryQuery.isLoading && !memoryQuery.error && (memoryQuery.data?.memoryEntries.length ?? 0) > 0 && filteredEntries.length === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
          <h3 className="text-lg font-semibold text-slate-900">No entries match this filter</h3>
          <p className="mt-2 text-sm text-slate-600">Try a different type filter to broaden the timeline.</p>
        </div>
      ) : null}

      {filteredEntries.length > 0 ? <MemoryTimeline entries={filteredEntries} onOpenDetail={setActiveMemoryID} /> : null}

      <MemoryDetailDrawer
        open={Boolean(activeMemoryID)}
        onClose={() => setActiveMemoryID(null)}
        entry={memoryDetailQuery.data?.memoryEntry ?? null}
        loading={memoryDetailQuery.isLoading}
        error={Boolean(memoryDetailQuery.error)}
      />
    </section>
  );
}
