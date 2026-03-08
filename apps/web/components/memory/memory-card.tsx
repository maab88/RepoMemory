import type { MemoryEntry } from "@repomemory/contracts";

type MemoryCardProps = {
  entry: MemoryEntry;
  onOpen: (memoryId: string) => void;
};

function typeLabel(type: string): string {
  if (type === "pr_summary") return "PR Summary";
  if (type === "issue_summary") return "Issue Summary";
  return type.replaceAll("_", " ");
}

export function MemoryCard({ entry, onOpen }: MemoryCardProps) {
  return (
    <article className="relative rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:border-slate-300 hover:shadow-md">
      <div className="flex items-start justify-between gap-3">
        <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-slate-700">
          {typeLabel(entry.type)}
        </span>
        <p className="text-xs text-slate-500">{new Date(entry.createdAt).toLocaleString()}</p>
      </div>

      <h3 className="mt-3 text-lg font-semibold text-slate-900">{entry.title}</h3>
      <p className="mt-2 text-sm leading-relaxed text-slate-700">{entry.summary}</p>
      {entry.whyItMatters ? (
        <p className="mt-3 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
          <span className="font-medium text-slate-900">Why it matters:</span> {entry.whyItMatters}
        </p>
      ) : null}

      <div className="mt-4 flex flex-wrap gap-2">
        {entry.impactedAreas.slice(0, 3).map((area) => (
          <span key={area} className="rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700">
            {area}
          </span>
        ))}
      </div>

      <div className="mt-4 flex items-center justify-between">
        <div className="text-xs text-slate-500">
          Source:{" "}
          {entry.sourceUrl ? (
            <a className="font-medium text-slate-700 underline decoration-slate-300 hover:text-slate-900" href={entry.sourceUrl} target="_blank" rel="noreferrer">
              Open source link
            </a>
          ) : (
            "No direct source URL"
          )}
        </div>
        <button
          type="button"
          onClick={() => onOpen(entry.id)}
          className="rounded-md border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:border-slate-400 hover:text-slate-900"
        >
          Open detail
        </button>
      </div>
    </article>
  );
}
