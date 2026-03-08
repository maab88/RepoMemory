"use client";

import { useState } from "react";
import { useParams } from "next/navigation";

import { RepositorySyncStatusCard } from "@/components/repositories/repository-sync-status-card";
import { useJobStatus } from "@/lib/hooks/use-job-status";
import { useRepositoryDetail } from "@/lib/hooks/use-repository-detail";
import { useTriggerSync } from "@/lib/hooks/use-trigger-sync";

export default function RepositoryDetailPage() {
  const params = useParams<{ repoId: string }>();
  const repoQuery = useRepositoryDetail(params.repoId);
  const triggerSync = useTriggerSync();
  const [activeJobID, setActiveJobID] = useState<string | null>(null);
  const jobQuery = useJobStatus(activeJobID);

  const onTriggerSync = async () => {
    const response = await triggerSync.mutateAsync(params.repoId);
    setActiveJobID(response.jobId);
  };

  return (
    <section className="space-y-6">
      {repoQuery.isLoading ? <div className="h-40 animate-pulse rounded-xl border border-slate-200 bg-white" /> : null}
      {repoQuery.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          Failed to load repository.
        </div>
      ) : null}

      {repoQuery.data ? (
        <>
          <article className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-2xl font-semibold tracking-tight text-slate-900">{repoQuery.data.repository.fullName}</h2>
            <p className="mt-2 text-sm text-slate-600">{repoQuery.data.repository.description || "No description provided."}</p>
            <div className="mt-4 flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={onTriggerSync}
                disabled={triggerSync.isPending}
                className="inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {triggerSync.isPending ? "Queueing..." : "Trigger initial sync"}
              </button>
              <a href={repoQuery.data.repository.htmlUrl} target="_blank" rel="noreferrer" className="text-sm font-medium text-slate-600 hover:text-slate-900">
                View on GitHub
              </a>
            </div>
          </article>

          <RepositorySyncStatusCard status={jobQuery.data?.job.status ?? repoQuery.data.repository.lastSyncStatus ?? "not_started"} job={jobQuery.data?.job ?? null} />
        </>
      ) : null}
    </section>
  );
}
