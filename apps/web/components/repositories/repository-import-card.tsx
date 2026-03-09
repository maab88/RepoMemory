"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { ImportedRepository } from "@repomemory/contracts";

import { RepositoryImportTable } from "@/components/repositories/repository-import-table";
import { RepositorySelector } from "@/components/repositories/repository-selector";
import { ErrorState } from "@/components/shared/error-state";
import { mapApiError } from "@/lib/errors/map-api-error";
import { useGitHubRepositories } from "@/lib/hooks/use-github-repositories";
import { useImportRepositories } from "@/lib/hooks/use-import-repositories";
import { listOrganizations } from "@/lib/organizations-api";

export function RepositoryImportCard() {
  const organizationsQuery = useQuery({ queryKey: ["organizations"], queryFn: listOrganizations });
  const reposQuery = useGitHubRepositories();
  const importMutation = useImportRepositories();

  const [organizationId, setOrganizationId] = useState("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [imported, setImported] = useState<ImportedRepository[]>([]);
  const [importFailed, setImportFailed] = useState(false);

  const repositories = reposQuery.data?.repositories ?? [];

  const selectedRepositories = useMemo(() => {
    const selected = new Set(selectedIds);
    return repositories.filter((repo) => selected.has(repo.githubRepoId));
  }, [repositories, selectedIds]);

  const mappedRepositoriesError = reposQuery.error ? mapApiError(reposQuery.error) : null;
  const githubReconnectRequired = mappedRepositoriesError?.code === "GITHUB_RECONNECT_REQUIRED";

  const importDisabled = importMutation.isPending || selectedRepositories.length === 0 || !organizationId;

  const onImport = async () => {
    const payload = {
      organizationId,
      repositories: selectedRepositories,
    };

    try {
      const data = await importMutation.mutateAsync(payload);
      setImported(data.importedRepositories);
      setImportFailed(false);
    } catch {
      setImported([]);
      setImportFailed(true);
    }
  };

  return (
    <section className="space-y-6 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="space-y-1">
        <h2 className="text-2xl font-semibold tracking-tight text-slate-900">Import repositories</h2>
        <p className="text-sm text-slate-600">Select repositories from GitHub and import them into an organization.</p>
      </div>

      <div className="space-y-2">
        <label htmlFor="org-select" className="block text-sm font-medium text-slate-700">
          Target organization
        </label>
        <select
          id="org-select"
          value={organizationId}
          onChange={(event) => setOrganizationId(event.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm outline-none ring-slate-900/10 focus:border-slate-900 focus:ring"
        >
          <option value="">Select organization</option>
          {(organizationsQuery.data ?? []).map((organization) => (
            <option key={organization.id} value={organization.id}>
              {organization.name}
            </option>
          ))}
        </select>
      </div>

      {reposQuery.isPending ? (
        <div className="space-y-2" aria-label="loading github repositories">
          <div className="h-10 animate-pulse rounded-lg bg-slate-100" />
          <div className="h-40 animate-pulse rounded-lg bg-slate-100" />
        </div>
      ) : null}

      {githubReconnectRequired ? (
        <ErrorState
          title={mappedRepositoriesError?.title ?? "Reconnect GitHub required"}
          message={mappedRepositoriesError?.message ?? "Reconnect GitHub to continue importing repositories."}
          requestId={mappedRepositoriesError?.requestId}
          action={
            <a className="font-semibold underline" href="/settings/integrations/github">
              Open GitHub integrations
            </a>
          }
        />
      ) : null}

      {!reposQuery.isPending && !githubReconnectRequired && repositories.length === 0 ? (
        <div className="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">No repositories found for this GitHub account.</div>
      ) : null}

      {!reposQuery.isPending && !githubReconnectRequired && repositories.length > 0 ? (
        <RepositorySelector repositories={repositories} selectedIds={selectedIds} onSelectionChange={setSelectedIds} />
      ) : null}

      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          disabled={importDisabled}
          onClick={onImport}
          className="inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {importMutation.isPending ? "Importing..." : `Import selected (${selectedRepositories.length})`}
        </button>
        <a href="/organizations" className="text-sm font-medium text-slate-600 hover:text-slate-900">
          Back to organizations
        </a>
      </div>

      {importMutation.error || importFailed ? (
        <p className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">Import failed. Please review selection and try again.</p>
      ) : null}

      <RepositoryImportTable repositories={imported} />
      {imported.length > 0 && organizationId ? (
        <a
          href={`/organizations/${organizationId}/repositories`}
          className="inline-flex text-sm font-medium text-slate-700 underline hover:text-slate-900"
        >
          Open persisted organization repository list
        </a>
      ) : null}
    </section>
  );
}
