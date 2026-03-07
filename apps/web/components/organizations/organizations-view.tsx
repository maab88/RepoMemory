"use client";

import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";

import { listOrganizations } from "@/lib/organizations-api";
import type { Organization } from "@/lib/types";

type OrganizationsViewProps = {
  organizations: Organization[];
};

export function OrganizationsView({ organizations }: OrganizationsViewProps) {
  if (organizations.length === 0) {
    return (
      <div className="rounded-xl border border-slate-200 bg-white p-8 text-center shadow-sm" data-testid="org-empty-state">
        <h2 className="text-xl font-semibold">No organizations yet</h2>
        <p className="mt-2 text-slate-600">Create your first organization to start syncing repositories.</p>
        <a
          href="/onboarding"
          className="mt-5 inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
        >
          Create organization
        </a>
      </div>
    );
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2">
      {organizations.map((org) => (
        <a
          key={org.id}
          href={`/organizations/${org.id}`}
          className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-slate-300"
        >
          <h3 className="text-lg font-semibold text-slate-900">{org.name}</h3>
          <p className="mt-1 text-sm text-slate-500">/{org.slug}</p>
          <p className="mt-3 text-xs uppercase tracking-wide text-slate-500">Role: {org.role}</p>
        </a>
      ))}
    </div>
  );
}

export function OrganizationsPageContent() {
  const query = useQuery({
    queryKey: ["organizations"],
    queryFn: listOrganizations,
  });

  const title = useMemo(() => {
    if (query.data && query.data.length > 0) {
      return "Your organizations";
    }
    return "Welcome to RepoMemory";
  }, [query.data]);

  return (
    <section className="space-y-6">
      <div className="space-y-1">
        <h2 className="text-2xl font-semibold tracking-tight">{title}</h2>
        <p className="text-slate-600">Manage organization access and onboarding.</p>
      </div>

      {query.isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2" aria-label="loading organizations">
          <div className="h-28 animate-pulse rounded-xl border border-slate-200 bg-white" />
          <div className="h-28 animate-pulse rounded-xl border border-slate-200 bg-white" />
        </div>
      ) : null}

      {query.error ? (
        <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          Failed to load organizations. Please refresh.
        </div>
      ) : null}

      {query.data ? <OrganizationsView organizations={query.data} /> : null}
    </section>
  );
}