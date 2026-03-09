"use client";

import { FormEvent, useState } from "react";
import { signIn } from "next-auth/react";
import { useSearchParams } from "next/navigation";

export function SignInForm({ enableDevCredentials }: { enableDevCredentials: boolean }) {
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") ?? "/organizations";
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [pendingGitHub, setPendingGitHub] = useState(false);
  const [pendingCredentials, setPendingCredentials] = useState(false);

  const onGitHubSignIn = async () => {
    setPendingGitHub(true);
    setError(null);
    await signIn("github", { callbackUrl });
    setPendingGitHub(false);
  };

  const onDevCredentials = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setPendingCredentials(true);
    setError(null);
    const result = await signIn("dev-credentials", {
      email,
      password,
      callbackUrl,
      redirect: false,
    });
    if (!result || result.error) {
      setError("Sign in failed. Check your credentials and try again.");
      setPendingCredentials(false);
      return;
    }

    window.location.href = result.url ?? callbackUrl;
  };

  return (
    <section className="mx-auto max-w-xl rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
      <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Sign in to RepoMemory</h1>
      <p className="mt-2 text-sm text-slate-600">Use your account to access organizations, repositories, and memory timeline features.</p>

      <button
        type="button"
        onClick={onGitHubSignIn}
        disabled={pendingGitHub}
        className="mt-6 inline-flex w-full items-center justify-center rounded-md bg-slate-900 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {pendingGitHub ? "Redirecting to GitHub..." : "Continue with GitHub"}
      </button>

      {enableDevCredentials ? (
        <form onSubmit={onDevCredentials} className="mt-6 space-y-3 rounded-xl border border-slate-200 bg-slate-50 p-4">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Development sign in</p>
          <input
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="Email"
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="Password"
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
          <button
            type="submit"
            disabled={pendingCredentials}
            className="inline-flex rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-800 hover:border-slate-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {pendingCredentials ? "Signing in..." : "Sign in with dev credentials"}
          </button>
        </form>
      ) : null}

      {error ? <p className="mt-4 text-sm text-rose-700">{error}</p> : null}
    </section>
  );
}
