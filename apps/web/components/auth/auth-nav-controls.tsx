"use client";

import { signOut, useSession } from "next-auth/react";

export function AuthNavControls() {
  const { data: session } = useSession();
  if (!session?.user?.id) {
    return (
      <a href="/sign-in" className="rounded-md border border-slate-300 px-2.5 py-1 text-xs font-medium text-slate-700 hover:border-slate-400">
        Sign in
      </a>
    );
  }

  const name = session.user.name ?? session.user.email ?? "Signed in";

  return (
    <div className="flex items-center gap-3">
      <span className="text-xs text-slate-500">{name}</span>
      <button
        type="button"
        onClick={() => signOut({ callbackUrl: "/sign-in" })}
        className="rounded-md border border-slate-300 px-2.5 py-1 text-xs font-medium text-slate-700 hover:border-slate-400"
      >
        Sign out
      </button>
    </div>
  );
}
