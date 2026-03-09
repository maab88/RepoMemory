import type { ReactNode } from "react";

type ErrorStateProps = {
  title: string;
  message: string;
  requestId?: string;
  action?: ReactNode;
};

export function ErrorState({ title, message, requestId, action }: ErrorStateProps) {
  return (
    <section className="rounded-lg border border-rose-200 bg-rose-50 p-4 text-rose-800">
      <h3 className="text-base font-semibold">{title}</h3>
      <p className="mt-1 text-sm">{message}</p>
      {requestId ? <p className="mt-2 text-xs text-rose-700/80">Request ID: {requestId}</p> : null}
      {action ? <div className="mt-3">{action}</div> : null}
    </section>
  );
}
