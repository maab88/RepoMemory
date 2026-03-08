import type { RepositorySummary } from "@repomemory/contracts";

import { RepositoryCard } from "@/components/repositories/repository-card";

type RepositoryListProps = {
  repositories: RepositorySummary[];
};

export function RepositoryList({ repositories }: RepositoryListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {repositories.map((repository) => (
        <RepositoryCard key={repository.id} repository={repository} />
      ))}
    </div>
  );
}

