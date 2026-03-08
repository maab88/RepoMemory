type MemorySearchBarProps = {
  query: string;
  onQueryChange: (value: string) => void;
  onSubmit: () => void;
  loading: boolean;
};

export function MemorySearchBar({ query, onQueryChange, onSubmit, loading }: MemorySearchBarProps) {
  return (
    <form
      className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit();
      }}
    >
      <label htmlFor="memory-search-input" className="mb-2 block text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
        Search Engineering Memory
      </label>
      <div className="flex flex-col gap-3 md:flex-row">
        <input
          id="memory-search-input"
          type="search"
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder="Search title or summary (e.g. retry, queue, permissions)"
          className="h-11 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-900 placeholder:text-slate-400 focus:border-slate-500 focus:outline-none"
        />
        <button
          type="submit"
          disabled={loading}
          className="h-11 rounded-lg bg-slate-900 px-5 text-sm font-semibold text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loading ? "Searching..." : "Search"}
        </button>
      </div>
    </form>
  );
}
