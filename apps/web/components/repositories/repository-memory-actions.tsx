type RepositoryMemoryActionsProps = {
  onGenerateMemory: () => Promise<void> | void;
  isGenerating: boolean;
  generationStatus: string | null;
  generationError: string | null;
};

export function RepositoryMemoryActions({
  onGenerateMemory,
  isGenerating,
  generationStatus,
  generationError,
}: RepositoryMemoryActionsProps) {
  return (
    <div className="space-y-2">
      <button
        type="button"
        onClick={onGenerateMemory}
        disabled={isGenerating}
        className="inline-flex rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-800 transition hover:border-slate-400 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {isGenerating ? "Queueing memory job..." : "Generate memory"}
      </button>
      {generationStatus ? <p className="text-xs text-slate-500">Memory job status: {generationStatus}</p> : null}
      {generationError ? <p className="text-sm text-rose-700">{generationError}</p> : null}
    </div>
  );
}
