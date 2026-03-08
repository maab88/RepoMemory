import type { RepositorySummary } from "@repomemory/contracts";

type RepositoryCardProps = {
  repository: RepositorySummary;
};

export function RepositoryCard({ repository }: RepositoryCardProps) {
  const lastSyncLabel = repository.lastSyncTime
    ? new Date(repository.lastSyncTime).toLocaleString()
    : "Not yet synced";

  return (
    <a
      href={`/repositories/${repository.id}`}
      className="group rounded-2xl border border-slate-200 bg-gradient-to-b from-white to-slate-50 p-5 shadow-sm transition hover:-translate-y-0.5 hover:border-slate-300 hover:shadow-md"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate text-lg font-semibold text-slate-900 group-hover:text-slate-950">{repository.fullName}</h3>
          <p className="mt-1 truncate text-sm text-slate-600">{repository.description || "No description provided."}</p>
        </div>
        <span className="rounded-full border border-slate-300 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-slate-600">
          {repository.private ? "Private" : "Public"}
        </span>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-3 text-xs text-slate-600">
        <p>Default branch: {repository.defaultBranch}</p>
        <p>Sync: {repository.lastSyncStatus || "not yet synced"}</p>
        <p>PRs: {repository.pullRequestCount ?? 0}</p>
        <p>Issues: {repository.issueCount ?? 0}</p>
        <p className="col-span-2">Last sync: {lastSyncLabel}</p>
      </div>
    </a>
  );
}
