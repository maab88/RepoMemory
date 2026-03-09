"use client";

import { useGitHubCallback } from "@/lib/hooks/use-github-callback";
import { ErrorState } from "@/components/shared/error-state";
import { mapApiError } from "@/lib/errors/map-api-error";

export function GitHubCallbackStatus({ code, state }: { code?: string; state?: string }) {
  const query = useGitHubCallback(code, state);

  if (!code || !state) {
    return (
      <section className="mx-auto max-w-2xl rounded-xl border border-rose-200 bg-rose-50 p-6 text-rose-700">
        <h2 className="text-xl font-semibold">Missing OAuth parameters</h2>
        <p className="mt-2 text-sm">The callback URL is missing required parameters. Start the GitHub connection again.</p>
        <a href="/settings/integrations/github" className="mt-4 inline-block text-sm font-semibold underline">
          Retry GitHub connect
        </a>
      </section>
    );
  }

  if (query.isPending) {
    return (
      <section className="mx-auto max-w-2xl rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="text-xl font-semibold">Finishing GitHub connection...</h2>
        <p className="mt-2 text-sm text-slate-600">Validating OAuth state and securing your account connection.</p>
      </section>
    );
  }

  if (query.error || !query.data) {
    const mapped = mapApiError(query.error);
    return (
      <div className="mx-auto max-w-2xl">
        <ErrorState
          title={mapped.title}
          message={mapped.message}
          requestId={mapped.requestId}
          action={
            <a href="/settings/integrations/github" className="inline-block text-sm font-semibold underline">
              Retry GitHub connect
            </a>
          }
        />
      </div>
    );
  }

  return (
    <section className="mx-auto max-w-2xl rounded-xl border border-emerald-200 bg-emerald-50 p-6 text-emerald-800">
      <h2 className="text-xl font-semibold">GitHub connected</h2>
      <p className="mt-2 text-sm">Your account <span className="font-semibold">{query.data.account.githubLogin}</span> is now connected.</p>
      <a href="/onboarding/repositories" className="mt-4 inline-block text-sm font-semibold underline">
        Continue to repository import
      </a>
    </section>
  );
}
