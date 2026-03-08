type MemoryFiltersProps = {
  selectedType: string;
  availableTypes: string[];
  onTypeChange: (value: string) => void;
};

export function MemoryFilters({ selectedType, availableTypes, onTypeChange }: MemoryFiltersProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">Filter by type</span>
      <button
        type="button"
        onClick={() => onTypeChange("all")}
        className={`rounded-full px-3 py-1 text-xs font-medium ${selectedType === "all" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700 hover:bg-slate-200"}`}
      >
        All
      </button>
      {availableTypes.map((type) => (
        <button
          key={type}
          type="button"
          onClick={() => onTypeChange(type)}
          className={`rounded-full px-3 py-1 text-xs font-medium ${selectedType === type ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700 hover:bg-slate-200"}`}
        >
          {type.replaceAll("_", " ")}
        </button>
      ))}
    </div>
  );
}
