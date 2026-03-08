import type { Job } from "@repomemory/contracts";

type RepositorySyncStatusCardProps = {
  status: string;
  lastSyncTime?: string | null;
  job?: Job | null;
};

const statusStyles: Record<string, string> = {
  queued: "border-amber-200 bg-amber-50 text-amber-700",
  running: "border-blue-200 bg-blue-50 text-blue-700",
  succeeded: "border-emerald-200 bg-emerald-50 text-emerald-700",
  failed: "border-rose-200 bg-rose-50 text-rose-700",
};

export function RepositorySyncStatusCard({ status, lastSyncTime, job }: RepositorySyncStatusCardProps) {
  const style = statusStyles[status] ?? "border-slate-200 bg-slate-50 text-slate-700";
  const normalizedStatus = status || "not yet synced";

  return (
    <article className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Sync status</h3>
      <div className={`mt-3 inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-wide ${style}`}>
        {normalizedStatus}
      </div>
      <p className="mt-3 text-sm text-slate-600">Last successful sync: {lastSyncTime ? new Date(lastSyncTime).toLocaleString() : "Not yet synced"}</p>
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
