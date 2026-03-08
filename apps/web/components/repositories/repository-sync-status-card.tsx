import type { Job } from "@repomemory/contracts";

type RepositorySyncStatusCardProps = {
  status: string;
  job?: Job | null;
};

export function RepositorySyncStatusCard({ status, job }: RepositorySyncStatusCardProps) {
  return (
    <article className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Sync status</h3>
      <p className="mt-2 text-lg font-semibold text-slate-900">{status || "not_started"}</p>
      {job ? (
        <div className="mt-3 space-y-1 text-sm text-slate-600">
          <p>Job ID: {job.id}</p>
          <p>Attempts: {job.attempts}</p>
          {job.lastError ? <p className="text-rose-700">Last error: {job.lastError}</p> : null}
        </div>
      ) : null}
    </article>
  );
}

