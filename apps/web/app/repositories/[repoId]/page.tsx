"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";

import { JobFailureBanner } from "@/components/jobs/job-failure-banner";
import { RepositoryMemoryActions } from "@/components/repositories/repository-memory-actions";
import { RepositorySyncStatusCard } from "@/components/repositories/repository-sync-status-card";
import { RepositoryStatsCards } from "@/components/repositories/repository-stats-cards";
import { ErrorState } from "@/components/shared/error-state";
import { RetryBanner } from "@/components/shared/retry-banner";
import { mapApiError } from "@/lib/errors/map-api-error";
import { useGenerateDigest } from "@/lib/hooks/use-generate-digest";
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
  const generateDigest = useGenerateDigest();
  const [activeJobID, setActiveJobID] = useState<string | null>(null);
  const [memoryJobID, setMemoryJobID] = useState<string | null>(null);
  const [digestJobID, setDigestJobID] = useState<string | null>(null);
  const [syncError, setSyncError] = useState<string | null>(null);
  const [memoryError, setMemoryError] = useState<string | null>(null);
  const [digestError, setDigestError] = useState<string | null>(null);
  const jobQuery = useJobStatus(activeJobID);
  const memoryJobQuery = useJobStatus(memoryJobID);
  const digestJobQuery = useJobStatus(digestJobID);

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

  useEffect(() => {
    const status = digestJobQuery.data?.job.status;
    if (status === "succeeded" || status === "failed") {
      void queryClient.invalidateQueries({ queryKey: ["repository-detail", params.repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repository-digests", params.repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repositories"] });
    }
    if (status === "failed") {
      setDigestError(digestJobQuery.data?.job.lastError ?? "Digest generation failed. Please try again.");
    }
  }, [digestJobQuery.data?.job.lastError, digestJobQuery.data?.job.status, params.repoId, queryClient]);

  const onTriggerSync = async () => {
    setSyncError(null);
    try {
      const response = await triggerSync.mutateAsync(params.repoId);
      setActiveJobID(response.jobId);
    } catch (error) {
      setSyncError(mapApiError(error).message);
    }
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

  const onGenerateDigest = async () => {
    setDigestError(null);
    try {
      const response = await generateDigest.mutateAsync(params.repoId);
      setDigestJobID(response.jobId);
    } catch {
      setDigestError("Could not queue digest generation.");
    }
  };

  return (
    <section className="space-y-6">
      {repoQuery.isLoading ? <div className="h-40 animate-pulse rounded-xl border border-slate-200 bg-white" /> : null}
      {repoQuery.error ? (
        <ErrorState {...mapApiError(repoQuery.error)} />
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
              <Link href={`/repositories/${params.repoId}/digests`} className="text-sm font-medium text-slate-600 underline decoration-slate-300 hover:text-slate-900">
                Open weekly digests
              </Link>
            </div>
            <div className="mt-4">
              <RepositoryMemoryActions
                onGenerateMemory={onGenerateMemory}
                isGenerating={generateMemory.isPending}
                generationStatus={memoryJobQuery.data?.job.status ?? null}
                generationError={memoryError}
              />
              <div className="mt-4 space-y-2">
                <button
                  type="button"
                  onClick={onGenerateDigest}
                  disabled={generateDigest.isPending}
                  className="inline-flex rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-800 transition hover:border-slate-400 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {generateDigest.isPending ? "Queueing digest job..." : "Generate weekly digest"}
                </button>
                {digestJobQuery.data?.job.status ? <p className="text-xs text-slate-500">Digest job status: {digestJobQuery.data.job.status}</p> : null}
                {digestError ? <p className="text-sm text-rose-700">{digestError}</p> : null}
              </div>
            </div>
            <div className="mt-4 space-y-2">
              {syncError ? <ErrorState title="Sync failed to queue" message={syncError} /> : null}
              {jobQuery.data?.job.status === "failed" ? <JobFailureBanner message={jobQuery.data.job.lastError} /> : null}
              {jobQuery.timedOut ? (
                <RetryBanner
                  message="Sync is still running but status polling timed out. You can continue working and retry status polling."
                  onRetry={() => {
                    setActiveJobID(null);
                    window.setTimeout(() => setActiveJobID(jobQuery.data?.job.id ?? null), 0);
                  }}
                />
              ) : null}
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
