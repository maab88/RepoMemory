export function RepositoryEmptyState() {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-8 text-center shadow-sm">
      <h3 className="text-lg font-semibold text-slate-900">No repositories imported yet</h3>
      <p className="mt-2 text-sm text-slate-600">
        Import repositories from GitHub to populate this organization workspace.
      </p>
      <a
        href="/onboarding/repositories"
        className="mt-5 inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
      >
        Import repositories
      </a>
    </div>
  );
}

