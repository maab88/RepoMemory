import { useQuery } from "@tanstack/react-query";

import { listRepositoryMemory } from "@/lib/repositories-api";

export function useRepositoryMemory(repoId: string) {
  return useQuery({
    queryKey: ["repository-memory", repoId],
    queryFn: () => listRepositoryMemory(repoId),
    enabled: Boolean(repoId),
  });
}
