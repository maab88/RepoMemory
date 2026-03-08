"use client";

import { useParams } from "next/navigation";

import { RepositoryEmptyState } from "@/components/repositories/repository-empty-state";
import { RepositoryList } from "@/components/repositories/repository-list";
import { useOrganizationRepositories } from "@/lib/hooks/use-organization-repositories";

export default function OrganizationRepositoriesPage() {
  const params = useParams<{ orgId: string }>();
  return <OrganizationRepositoriesContent orgId={params.orgId} />;
}

function OrganizationRepositoriesContent({ orgId }: { orgId: string }) {
  const query = useOrganizationRepositories(orgId);

  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-3xl font-semibold tracking-tight">Repositories</h2>
        <p className="text-slate-600">Imported repositories persisted for this organization.</p>
      </div>

      {query.isLoading ? (
        <div className="space-y-3" aria-label="loading repositories">
          <div className="h-28 animate-pulse rounded-xl border border-slate-200 bg-white" />
          <div className="h-28 animate-pulse rounded-xl border border-slate-200 bg-white" />
        </div>
      ) : null}

      {query.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          Failed to load repositories.
        </div>
      ) : null}

      {!query.isLoading && !query.error && (query.data?.repositories.length ?? 0) === 0 ? <RepositoryEmptyState /> : null}

      {query.data && query.data.repositories.length > 0 ? <RepositoryList repositories={query.data.repositories} /> : null}
    </section>
  );
}
