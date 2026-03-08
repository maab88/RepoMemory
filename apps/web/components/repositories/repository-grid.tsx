import type { RepositorySummary } from "@repomemory/contracts";

import { RepositoryCard } from "@/components/repositories/repository-card";

type RepositoryGridProps = {
  repositories: RepositorySummary[];
};

export function RepositoryGrid({ repositories }: RepositoryGridProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {repositories.map((repository) => (
        <RepositoryCard key={repository.id} repository={repository} />
      ))}
    </div>
  );
}

