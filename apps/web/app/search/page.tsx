"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { MemorySearchBar } from "@/components/search/memory-search-bar";
import { MemorySearchResults } from "@/components/search/memory-search-results";
import { useMemorySearch } from "@/lib/hooks/use-memory-search";
import { useRepositories } from "@/lib/hooks/use-repositories";
import { listOrganizations } from "@/lib/organizations-api";

const DEFAULT_PAGE_SIZE = 20;

export default function SearchPage() {
  const organizationsQuery = useQuery({
    queryKey: ["organizations"],
    queryFn: () => listOrganizations(),
  });
  const repositoriesQuery = useRepositories();

  const [queryInput, setQueryInput] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [organizationId, setOrganizationID] = useState("");
  const [repositoryId, setRepositoryID] = useState("");
  const [page, setPage] = useState(1);

  useEffect(() => {
    if (!organizationId && organizationsQuery.data && organizationsQuery.data.length > 0) {
      setOrganizationID(organizationsQuery.data[0].id);
    }
  }, [organizationId, organizationsQuery.data]);

  const scopedRepositories = useMemo(
    () => (repositoriesQuery.data?.repositories ?? []).filter((repo) => repo.organizationId === organizationId),
    [organizationId, repositoriesQuery.data?.repositories]
  );

  useEffect(() => {
    if (!repositoryId) return;
    const exists = scopedRepositories.some((repo) => repo.id === repositoryId);
    if (!exists) {
      setRepositoryID("");
    }
  }, [repositoryId, scopedRepositories]);

  const searchQuery = useMemorySearch({
    query: submittedQuery,
    organizationId,
    repositoryId: repositoryId || undefined,
    page,
    pageSize: DEFAULT_PAGE_SIZE,
  });

  return (
    <section className="space-y-6">
      <header className="space-y-2">
        <p className="text-xs uppercase tracking-[0.18em] text-slate-500">Memory Search</p>
        <h2 className="text-3xl font-semibold tracking-tight text-slate-900">Search Team Memory</h2>
        <p className="text-slate-600">Search persisted memory entries across titles and summaries with organization and repository scope.</p>
      </header>

      <div className="grid gap-4 md:grid-cols-2">
        <label className="space-y-2 text-sm font-medium text-slate-700">
          Organization scope
          <select
            value={organizationId}
            onChange={(event) => {
              setOrganizationID(event.target.value);
              setPage(1);
            }}
            className="h-11 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-slate-500 focus:outline-none"
          >
            {(organizationsQuery.data ?? []).map((org) => (
              <option key={org.id} value={org.id}>
                {org.name}
              </option>
            ))}
          </select>
        </label>

        <label className="space-y-2 text-sm font-medium text-slate-700">
          Repository scope (optional)
          <select
            value={repositoryId}
            onChange={(event) => {
              setRepositoryID(event.target.value);
              setPage(1);
            }}
            className="h-11 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-slate-500 focus:outline-none"
          >
            <option value="">All repositories</option>
            {scopedRepositories.map((repo) => (
              <option key={repo.id} value={repo.id}>
                {repo.fullName}
              </option>
            ))}
          </select>
        </label>
      </div>

      <MemorySearchBar
        query={queryInput}
        onQueryChange={setQueryInput}
        onSubmit={() => {
          setSubmittedQuery(queryInput);
          setPage(1);
        }}
        loading={searchQuery.isFetching}
      />

      {submittedQuery.trim().length === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-10 text-center shadow-sm">
          <h3 className="text-xl font-semibold text-slate-900">Start with a search term</h3>
          <p className="mt-2 text-sm text-slate-600">Try terms like retry, sync, queue, auth, or migration to find relevant engineering memory.</p>
        </div>
      ) : null}

      {searchQuery.isLoading ? (
        <div className="space-y-3" aria-label="loading search results">
          <div className="h-40 animate-pulse rounded-2xl border border-slate-200 bg-white" />
          <div className="h-40 animate-pulse rounded-2xl border border-slate-200 bg-white" />
        </div>
      ) : null}

      {searchQuery.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          Could not load memory search results. Please verify organization access and try again.
        </div>
      ) : null}

      {!searchQuery.isLoading &&
      !searchQuery.error &&
      searchQuery.data &&
      searchQuery.data.query !== "" &&
      searchQuery.data.results.length === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-10 text-center shadow-sm">
          <h3 className="text-xl font-semibold text-slate-900">No memory results found</h3>
          <p className="mt-2 text-sm text-slate-600">Try a broader query or remove the repository filter.</p>
        </div>
      ) : null}

      {searchQuery.data && searchQuery.data.results.length > 0 ? (
        <MemorySearchResults data={searchQuery.data} page={page} onPageChange={setPage} />
      ) : null}
    </section>
  );
}
