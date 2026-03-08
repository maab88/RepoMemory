import { useQuery } from "@tanstack/react-query";

import { getRepositoryDetail } from "@/lib/repositories-api";

export function useRepositoryDetail(repoId: string) {
  return useQuery({
    queryKey: ["repository-detail", repoId],
    queryFn: () => getRepositoryDetail(repoId),
    enabled: Boolean(repoId),
  });
}

