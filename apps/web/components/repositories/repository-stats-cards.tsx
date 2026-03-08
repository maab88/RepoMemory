type RepositoryStatsCardsProps = {
  pullRequestCount: number;
  issueCount: number;
  memoryEntryCount: number;
};

export function RepositoryStatsCards({ pullRequestCount, issueCount, memoryEntryCount }: RepositoryStatsCardsProps) {
  const cards = [
    { label: "Pull requests", value: pullRequestCount ?? 0 },
    { label: "Issues", value: issueCount ?? 0 },
    { label: "Memory entries", value: memoryEntryCount ?? 0 },
  ];

  return (
    <section className="grid gap-3 sm:grid-cols-3">
      {cards.map((card) => (
        <article key={card.label} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <p className="text-xs uppercase tracking-wide text-slate-500">{card.label}</p>
          <p className="mt-2 text-2xl font-semibold text-slate-900">{card.value}</p>
        </article>
      ))}
    </section>
  );
}

