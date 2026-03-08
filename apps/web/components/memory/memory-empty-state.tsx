type MemoryEmptyStateProps = {
  onGenerateMemory: () => Promise<void> | void;
  isGenerating: boolean;
  generationStatus: string | null;
  generationError: string | null;
};

export function MemoryEmptyState({ onGenerateMemory, isGenerating, generationStatus, generationError }: MemoryEmptyStateProps) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
      <h3 className="text-xl font-semibold text-slate-900">No memory entries yet</h3>
      <p className="mt-2 text-sm text-slate-600">Run repository sync first if needed, then generate memory to populate this timeline.</p>
      <button
        type="button"
        onClick={onGenerateMemory}
        disabled={isGenerating}
        className="mt-4 inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {isGenerating ? "Queueing memory job..." : "Generate memory"}
      </button>
      {generationStatus ? <p className="mt-2 text-xs text-slate-500">Memory job status: {generationStatus}</p> : null}
      {generationError ? <p className="mt-2 text-sm text-rose-700">{generationError}</p> : null}
    </div>
  );
}
