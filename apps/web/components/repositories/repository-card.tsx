import type { RepositorySummary } from "@repomemory/contracts";

type RepositoryCardProps = {
  repository: RepositorySummary;
};

export function RepositoryCard({ repository }: RepositoryCardProps) {
  return (
    <a
      href={`/repositories/${repository.id}`}
      className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-slate-300"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate text-lg font-semibold text-slate-900">{repository.fullName}</h3>
          <p className="mt-1 truncate text-sm text-slate-600">{repository.description || "No description provided."}</p>
        </div>
        <span className="rounded-full border border-slate-300 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-slate-600">
          {repository.private ? "Private" : "Public"}
        </span>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-3 text-xs text-slate-500">
        <p>Default branch: {repository.defaultBranch}</p>
        <p>Sync: {repository.lastSyncStatus || "not_started"}</p>
        <p>PRs: {repository.pullRequestCount}</p>
        <p>Issues: {repository.issueCount}</p>
      </div>
    </a>
  );
}

