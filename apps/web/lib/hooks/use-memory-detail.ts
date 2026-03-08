import { useQuery } from "@tanstack/react-query";

import { getRepositoryMemoryDetail } from "@/lib/repositories-api";

export function useMemoryDetail(repoId: string, memoryId: string | null) {
  return useQuery({
    queryKey: ["repository-memory-detail", repoId, memoryId],
    queryFn: () => getRepositoryMemoryDetail(repoId, memoryId ?? ""),
    enabled: Boolean(repoId && memoryId),
  });
}
