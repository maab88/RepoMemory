"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";

import { RepositoryMemoryActions } from "@/components/repositories/repository-memory-actions";
import { RepositorySyncStatusCard } from "@/components/repositories/repository-sync-status-card";
import { RepositoryStatsCards } from "@/components/repositories/repository-stats-cards";
import { useGenerateMemory } from "@/lib/hooks/use-generate-memory";
import { useJobStatus } from "@/lib/hooks/use-job-status";
import { useRepositoryDetail } from "@/lib/hooks/use-repository-detail";
import { useTriggerSync } from "@/lib/hooks/use-trigger-sync";

export default function RepositoryDetailPage() {
  const params = useParams<{ repoId: string }>();
  const queryClient = useQueryClient();
  const repoQuery = useRepositoryDetail(params.repoId);
  const triggerSync = useTriggerSync();
  const generateMemory = useGenerateMemory();
  const [activeJobID, setActiveJobID] = useState<string | null>(null);
  const [memoryJobID, setMemoryJobID] = useState<string | null>(null);
  const [memoryError, setMemoryError] = useState<string | null>(null);
  const jobQuery = useJobStatus(activeJobID);
  const memoryJobQuery = useJobStatus(memoryJobID);

  useEffect(() => {
    const status = jobQuery.data?.job.status;
    if (status === "succeeded" || status === "failed") {
      void queryClient.invalidateQueries({ queryKey: ["repository-detail", params.repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repositories"] });
      void queryClient.invalidateQueries({ queryKey: ["organization-repositories"] });
    }
  }, [jobQuery.data?.job.status, params.repoId, queryClient]);

  useEffect(() => {
    const status = memoryJobQuery.data?.job.status;
    if (status === "succeeded" || status === "failed") {
      void queryClient.invalidateQueries({ queryKey: ["repository-detail", params.repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repository-memory", params.repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repositories"] });
      void queryClient.invalidateQueries({ queryKey: ["organization-repositories"] });
    }
    if (status === "failed") {
      setMemoryError(memoryJobQuery.data?.job.lastError ?? "Memory generation failed. Please try again.");
    }
  }, [memoryJobQuery.data?.job.lastError, memoryJobQuery.data?.job.status, params.repoId, queryClient]);

  const onTriggerSync = async () => {
    const response = await triggerSync.mutateAsync(params.repoId);
    setActiveJobID(response.jobId);
  };

  const onGenerateMemory = async () => {
    setMemoryError(null);
    try {
      const response = await generateMemory.mutateAsync(params.repoId);
      setMemoryJobID(response.jobId);
    } catch {
      setMemoryError("Could not queue memory generation.");
    }
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
          <article className="rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50 p-6 shadow-sm">
            <p className="text-xs uppercase tracking-[0.18em] text-slate-500">Repository Detail</p>
            <h2 className="mt-1 text-3xl font-semibold tracking-tight text-slate-900">{repoQuery.data.repository.fullName}</h2>
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
              <Link href={`/repositories/${params.repoId}/memory`} className="text-sm font-medium text-slate-600 underline decoration-slate-300 hover:text-slate-900">
                Open memory timeline
              </Link>
            </div>
            <div className="mt-4">
              <RepositoryMemoryActions
                onGenerateMemory={onGenerateMemory}
                isGenerating={generateMemory.isPending}
                generationStatus={memoryJobQuery.data?.job.status ?? null}
                generationError={memoryError}
              />
            </div>
          </article>

          <div className="grid gap-4 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <RepositoryStatsCards
                pullRequestCount={repoQuery.data.repository.pullRequestCount ?? 0}
                issueCount={repoQuery.data.repository.issueCount ?? 0}
                memoryEntryCount={repoQuery.data.repository.memoryEntryCount ?? 0}
              />
            </div>
            <div>
              <RepositorySyncStatusCard
                status={jobQuery.data?.job.status ?? repoQuery.data.repository.lastSyncStatus ?? "not yet synced"}
                lastSyncTime={repoQuery.data.repository.lastSyncTime ?? null}
                job={jobQuery.data?.job ?? null}
              />
            </div>
          </div>
        </>
      ) : null}
    </section>
  );
}
