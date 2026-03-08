import { useQuery } from "@tanstack/react-query";

import { listRepositoryDigests } from "@/lib/repositories-api";

export function useRepositoryDigests(repoId: string) {
  return useQuery({
    queryKey: ["repository-digests", repoId],
    queryFn: () => listRepositoryDigests(repoId),
    enabled: Boolean(repoId),
  });
}
