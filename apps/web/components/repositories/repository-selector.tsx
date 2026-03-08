"use client";

import { useMemo, useState } from "react";

export type SelectableGitHubRepository = {
  githubRepoId: string;
  ownerLogin: string;
  name: string;
  fullName: string;
  private: boolean;
  defaultBranch: string;
  htmlUrl: string;
  description?: string;
};

type RepositorySelectorProps = {
  repositories: SelectableGitHubRepository[];
  selectedIds: string[];
  onSelectionChange: (selectedIds: string[]) => void;
};

export function RepositorySelector({ repositories, selectedIds, onSelectionChange }: RepositorySelectorProps) {
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) {
      return repositories;
    }
    return repositories.filter((repo) => {
      return (
        repo.name.toLowerCase().includes(needle) ||
        repo.fullName.toLowerCase().includes(needle) ||
        (repo.description ?? "").toLowerCase().includes(needle)
      );
    });
  }, [query, repositories]);

  const toggle = (id: string) => {
    if (selectedIds.includes(id)) {
      onSelectionChange(selectedIds.filter((value) => value !== id));
      return;
    }
    onSelectionChange([...selectedIds, id]);
  };

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search repositories"
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm outline-none ring-slate-900/10 focus:border-slate-900 focus:ring"
          aria-label="Search repositories"
        />
        <p className="text-xs font-medium uppercase tracking-wide text-slate-500">Selected: {selectedIds.length}</p>
      </div>

      <div className="max-h-[380px] overflow-auto rounded-xl border border-slate-200 bg-white">
        {filtered.length === 0 ? (
          <p className="px-4 py-6 text-sm text-slate-600">No repositories match your search.</p>
        ) : (
          <ul className="divide-y divide-slate-200">
            {filtered.map((repo) => {
              const checked = selectedIds.includes(repo.githubRepoId);
              return (
                <li key={repo.githubRepoId} className="flex items-start gap-3 px-4 py-3">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => toggle(repo.githubRepoId)}
                    aria-label={`Select ${repo.fullName}`}
                    className="mt-1 h-4 w-4 rounded border-slate-300 text-slate-900"
                  />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-semibold text-slate-900">{repo.fullName}</p>
                    <p className="truncate text-xs text-slate-500">Default branch: {repo.defaultBranch}</p>
                    {repo.description ? <p className="mt-1 truncate text-xs text-slate-600">{repo.description}</p> : null}
                  </div>
                  <span className="rounded-full border border-slate-300 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-slate-600">
                    {repo.private ? "Private" : "Public"}
                  </span>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </section>
  );
}