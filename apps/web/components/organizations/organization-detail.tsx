"use client";

import { useQuery } from "@tanstack/react-query";

import { getOrganization } from "@/lib/organizations-api";

export function OrganizationDetailClient({ orgId }: { orgId: string }) {
  const query = useQuery({
    queryKey: ["organization", orgId],
    queryFn: () => getOrganization(orgId),
  });

  return (
    <section className="mx-auto max-w-3xl space-y-6">
      {query.isLoading ? <div className="h-32 animate-pulse rounded-xl border border-slate-200 bg-white" /> : null}
      {query.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">Unable to load organization.</div>
      ) : null}
      {query.data ? (
        <article className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-2xl font-semibold tracking-tight">{query.data.name}</h2>
          <p className="mt-1 text-sm text-slate-500">/{query.data.slug}</p>
          <div className="mt-5 grid gap-3 sm:grid-cols-2">
            <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
              <p className="text-xs uppercase tracking-wide text-slate-500">Role</p>
              <p className="mt-1 text-sm font-medium text-slate-900">{query.data.role}</p>
            </div>
            <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
              <p className="text-xs uppercase tracking-wide text-slate-500">Status</p>
              <p className="mt-1 text-sm font-medium text-slate-900">Ready for repository import</p>
            </div>
          </div>
        </article>
      ) : null}
    </section>
  );
}