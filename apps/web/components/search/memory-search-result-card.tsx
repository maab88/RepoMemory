import Link from "next/link";
import type { MemorySearchResult } from "@repomemory/contracts";

type MemorySearchResultCardProps = {
  result: MemorySearchResult;
};

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.valueOf())) return "Unknown date";
  return date.toLocaleString();
}

function labelForType(type: string): string {
  return type.replaceAll("_", " ");
}

export function MemorySearchResultCard({ result }: MemorySearchResultCardProps) {
  return (
    <Link
      href={`/repositories/${result.repositoryId}/memory?memoryId=${result.id}`}
      className="group block rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:border-slate-300 hover:shadow"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">{result.repositoryName}</p>
        <span className="rounded-full border border-slate-300 px-2 py-1 text-xs font-semibold uppercase text-slate-700">
          {labelForType(result.type)}
        </span>
      </div>
      <h3 className="mt-2 text-lg font-semibold text-slate-900 group-hover:text-slate-700">{result.title}</h3>
      <p className="mt-2 text-sm leading-6 text-slate-600">{result.summarySnippet}</p>
      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500">
        <span>{formatDate(result.createdAt)}</span>
        {result.sourceUrl ? <span className="truncate">Source: {result.sourceUrl}</span> : null}
      </div>
    </Link>
  );
}
