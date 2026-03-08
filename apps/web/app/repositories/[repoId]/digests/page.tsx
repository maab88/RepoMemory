"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";

import { DigestDetail } from "@/components/digests/digest-detail";
import { DigestList } from "@/components/digests/digest-list";
import { useGenerateDigest } from "@/lib/hooks/use-generate-digest";
import { useJobStatus } from "@/lib/hooks/use-job-status";
import { useRepositoryDetail } from "@/lib/hooks/use-repository-detail";
import { useRepositoryDigests } from "@/lib/hooks/use-repository-digests";

export default function RepositoryDigestsPage() {
  const params = useParams<{ repoId: string }>();
  const repoId = params.repoId;
  const queryClient = useQueryClient();

  const repositoryQuery = useRepositoryDetail(repoId);
  const digestsQuery = useRepositoryDigests(repoId);
  const generateDigest = useGenerateDigest();
  const [activeDigestID, setActiveDigestID] = useState<string | null>(null);
  const [digestJobID, setDigestJobID] = useState<string | null>(null);
  const [digestError, setDigestError] = useState<string | null>(null);
  const digestJobQuery = useJobStatus(digestJobID);

  useEffect(() => {
    const status = digestJobQuery.data?.job.status;
    if (status === "succeeded" || status === "failed") {
      void queryClient.invalidateQueries({ queryKey: ["repository-digests", repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repository-detail", repoId] });
      void queryClient.invalidateQueries({ queryKey: ["repositories"] });
    }
    if (status === "failed") {
      setDigestError(digestJobQuery.data?.job.lastError ?? "Digest generation failed. Please try again.");
    }
  }, [digestJobQuery.data?.job.lastError, digestJobQuery.data?.job.status, queryClient, repoId]);

  useEffect(() => {
    if (!activeDigestID && (digestsQuery.data?.digests.length ?? 0) > 0) {
      setActiveDigestID(digestsQuery.data?.digests[0]?.id ?? null);
    }
  }, [activeDigestID, digestsQuery.data?.digests]);

  const selectedDigest = useMemo(() => {
    if (!activeDigestID) return null;
    return (digestsQuery.data?.digests ?? []).find((item) => item.id === activeDigestID) ?? null;
  }, [activeDigestID, digestsQuery.data?.digests]);

  const onGenerateDigest = async () => {
    setDigestError(null);
    try {
      const response = await generateDigest.mutateAsync(repoId);
      setDigestJobID(response.jobId);
    } catch {
      setDigestError("Could not queue digest generation.");
    }
  };

  return (
    <section className="space-y-6">
      <header className="rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50 p-6 shadow-sm">
        <p className="text-xs uppercase tracking-[0.18em] text-slate-500">Repository Digests</p>
        <h2 className="mt-2 text-3xl font-semibold tracking-tight text-slate-900">
          {repositoryQuery.data?.repository.fullName ?? "Loading repository..."}
        </h2>
        <p className="mt-2 text-sm text-slate-600">Browse concise weekly digests built from persisted pull requests, issues, and memory entries.</p>
        <div className="mt-4 flex flex-wrap gap-3">
          <button
            type="button"
            onClick={onGenerateDigest}
            disabled={generateDigest.isPending}
            className="inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {generateDigest.isPending ? "Queueing digest job..." : "Generate weekly digest"}
          </button>
          <Link href={`/repositories/${repoId}`} className="rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:border-slate-400">
            Back to repository
          </Link>
        </div>
        {digestJobQuery.data?.job.status ? <p className="mt-3 text-xs text-slate-500">Digest job status: {digestJobQuery.data.job.status}</p> : null}
        {digestError ? <p className="mt-2 text-sm text-rose-700">{digestError}</p> : null}
      </header>

      {digestsQuery.isLoading ? (
        <div className="space-y-4" aria-label="loading digests">
          <div className="h-40 animate-pulse rounded-2xl border border-slate-200 bg-white" />
          <div className="h-40 animate-pulse rounded-2xl border border-slate-200 bg-white" />
        </div>
      ) : null}

      {digestsQuery.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">Failed to load digests.</div>
      ) : null}

      {!digestsQuery.isLoading && !digestsQuery.error && (digestsQuery.data?.digests.length ?? 0) === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
          <h3 className="text-xl font-semibold text-slate-900">No weekly digests yet</h3>
          <p className="mt-2 text-sm text-slate-600">Generate a digest to summarize this week's repository activity.</p>
          <button
            type="button"
            onClick={onGenerateDigest}
            disabled={generateDigest.isPending}
            className="mt-4 inline-flex rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-800 hover:border-slate-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {generateDigest.isPending ? "Queueing digest job..." : "Generate weekly digest"}
          </button>
        </div>
      ) : null}

      {(digestsQuery.data?.digests.length ?? 0) > 0 ? (
        <div className="grid gap-5 lg:grid-cols-2">
          <DigestList digests={digestsQuery.data?.digests ?? []} selectedDigestId={activeDigestID} onSelectDigest={setActiveDigestID} />
          <DigestDetail digest={selectedDigest} />
        </div>
      ) : null}
    </section>
  );
}
