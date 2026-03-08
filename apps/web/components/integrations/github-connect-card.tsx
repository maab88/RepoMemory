"use client";

import { useGitHubConnect } from "@/lib/hooks/use-github-connect";

export function GitHubConnectCard() {
  const connectMutation = useGitHubConnect();

  const onConnect = async () => {
    try {
      const data = await connectMutation.mutateAsync(undefined);
      window.location.assign(data.redirectUrl);
    } catch {
      // Error UI is rendered below.
    }
  };

  return (
    <article className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between gap-4">
        <div className="space-y-2">
          <p className="inline-flex rounded-full border border-slate-300 px-3 py-1 text-xs font-medium uppercase tracking-wide text-slate-700">
            GitHub OAuth
          </p>
          <h2 className="text-2xl font-semibold tracking-tight text-slate-900">Connect your GitHub account</h2>
          <p className="max-w-2xl text-sm text-slate-600">
            Connect once to enable repository discovery and future sync flows. Access tokens stay on the API server and are never exposed in the browser.
          </p>
        </div>
        <div className="hidden rounded-xl bg-slate-900 px-4 py-2 text-xs font-medium uppercase tracking-wider text-white sm:block">v1</div>
      </div>

      <div className="mt-6 flex items-center gap-3">
        <button
          type="button"
          onClick={onConnect}
          disabled={connectMutation.isPending}
          className="inline-flex items-center rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {connectMutation.isPending ? "Preparing GitHub redirect..." : "Connect GitHub"}
        </button>
        <a href="/organizations" className="text-sm font-medium text-slate-600 hover:text-slate-900">
          Back to organizations
        </a>
      </div>

      {connectMutation.error ? (
        <p className="mt-4 rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700" role="alert">
          Could not start GitHub OAuth. Check API GitHub env vars and try again.
        </p>
      ) : null}
    </article>
  );
}
