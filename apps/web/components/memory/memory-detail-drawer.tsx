import type { MemoryEntryDetail } from "@repomemory/contracts";

type MemoryDetailDrawerProps = {
  open: boolean;
  onClose: () => void;
  entry: MemoryEntryDetail | null;
  loading: boolean;
  error: boolean;
};

export function MemoryDetailDrawer({ open, onClose, entry, loading, error }: MemoryDetailDrawerProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-40 flex justify-end bg-slate-900/40 p-0" role="dialog" aria-modal="true" aria-label="Memory detail">
      <div className="h-full w-full max-w-xl overflow-y-auto bg-white p-6 shadow-2xl">
        <div className="mb-5 flex items-center justify-between">
          <h3 className="text-xl font-semibold text-slate-900">Memory detail</h3>
          <button type="button" onClick={onClose} className="rounded-md border border-slate-300 px-3 py-1 text-sm text-slate-700 hover:border-slate-400">
            Close
          </button>
        </div>

        {loading ? <div className="h-40 animate-pulse rounded-xl border border-slate-200 bg-slate-50" /> : null}
        {error ? <p className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">Failed to load memory detail.</p> : null}

        {entry ? (
          <div className="space-y-4">
            <p className="text-xs uppercase tracking-[0.16em] text-slate-500">{entry.type.replaceAll("_", " ")}</p>
            <h4 className="text-2xl font-semibold text-slate-900">{entry.title}</h4>
            <p className="text-sm leading-relaxed text-slate-700">{entry.summary}</p>
            {entry.whyItMatters ? (
              <div className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700">
                <span className="font-medium text-slate-900">Why it matters:</span> {entry.whyItMatters}
              </div>
            ) : null}

            <Section title="Impacted Areas" items={entry.impactedAreas} />
            <Section title="Risks" items={entry.risks} />
            <Section title="Follow-ups" items={entry.followUps} />

            <section>
              <h5 className="text-sm font-semibold uppercase tracking-wide text-slate-600">Sources</h5>
              <div className="mt-2 space-y-2">
                {entry.sources.length === 0 ? (
                  <p className="text-sm text-slate-500">No linked sources.</p>
                ) : (
                  entry.sources.map((source) => (
                    <div key={`${source.sourceType}-${source.displayLabel}`} className="rounded-md border border-slate-200 p-3">
                      <p className="text-sm font-medium text-slate-900">{source.displayLabel}</p>
                      <p className="text-xs uppercase tracking-wide text-slate-500">{source.sourceType.replaceAll("_", " ")}</p>
                      {source.sourceUrl ? (
                        <a href={source.sourceUrl} target="_blank" rel="noreferrer" className="mt-1 inline-block text-sm text-slate-700 underline decoration-slate-300">
                          Open source
                        </a>
                      ) : null}
                    </div>
                  ))
                )}
              </div>
            </section>
          </div>
        ) : null}
      </div>
    </div>
  );
}

function Section({ title, items }: { title: string; items: string[] }) {
  return (
    <section>
      <h5 className="text-sm font-semibold uppercase tracking-wide text-slate-600">{title}</h5>
      <div className="mt-2 flex flex-wrap gap-2">
        {items.length === 0 ? <p className="text-sm text-slate-500">None noted.</p> : null}
        {items.map((item) => (
          <span key={`${title}-${item}`} className="rounded-md bg-slate-100 px-2 py-1 text-xs font-medium text-slate-700">
            {item}
          </span>
        ))}
      </div>
    </section>
  );
}
