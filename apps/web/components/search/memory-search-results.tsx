import type { MemorySearchData } from "@repomemory/contracts";

import { MemorySearchResultCard } from "@/components/search/memory-search-result-card";

type MemorySearchResultsProps = {
  data: MemorySearchData;
  page: number;
  onPageChange: (page: number) => void;
};

export function MemorySearchResults({ data, page, onPageChange }: MemorySearchResultsProps) {
  const totalPages = Math.max(1, Math.ceil(data.total / data.pageSize));

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between text-sm text-slate-600">
        <p>
          {data.total} result{data.total === 1 ? "" : "s"} for "{data.query}"
        </p>
        <p>
          Page {data.page} of {totalPages}
        </p>
      </div>

      <div className="space-y-3">
        {data.results.map((result) => (
          <MemorySearchResultCard key={result.id} result={result} />
        ))}
      </div>

      <div className="flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
          className="rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:border-slate-400 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Previous
        </button>
        <button
          type="button"
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages}
          className="rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:border-slate-400 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Next
        </button>
      </div>
    </div>
  );
}
