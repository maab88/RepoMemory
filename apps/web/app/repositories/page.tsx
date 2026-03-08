"use client";

import { RepositoryEmptyState } from "@/components/repositories/repository-empty-state";
import { RepositoryGrid } from "@/components/repositories/repository-grid";
import { useRepositories } from "@/lib/hooks/use-repositories";

export default function RepositoriesPage() {
  const query = useRepositories();

  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-3xl font-semibold tracking-tight text-slate-900">Repository Dashboard</h2>
        <p className="text-slate-600">Track imported repositories, sync status, and ingestion counts from persisted data.</p>
      </div>

      {query.isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3" aria-label="loading repositories">
          <div className="h-44 animate-pulse rounded-2xl border border-slate-200 bg-white" />
          <div className="h-44 animate-pulse rounded-2xl border border-slate-200 bg-white" />
          <div className="h-44 animate-pulse rounded-2xl border border-slate-200 bg-white" />
        </div>
      ) : null}

      {query.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          Failed to load repositories.
        </div>
      ) : null}

      {!query.isLoading && !query.error && (query.data?.repositories.length ?? 0) === 0 ? <RepositoryEmptyState /> : null}

      {query.data && query.data.repositories.length > 0 ? <RepositoryGrid repositories={query.data.repositories} /> : null}
    </section>
  );
}
