import { useQuery } from "@tanstack/react-query";

import { searchMemory } from "@/lib/repositories-api";

type UseMemorySearchInput = {
  query: string;
  organizationId: string;
  repositoryId?: string;
  page: number;
  pageSize: number;
};

export function useMemorySearch(input: UseMemorySearchInput) {
  const trimmed = input.query.trim();
  return useQuery({
    queryKey: ["memory-search", input.organizationId, input.repositoryId ?? "all", trimmed, input.page, input.pageSize],
    queryFn: () =>
      searchMemory({
        q: trimmed,
        organizationId: input.organizationId,
        repositoryId: input.repositoryId,
        page: input.page,
        pageSize: input.pageSize,
      }),
    enabled: Boolean(input.organizationId) && trimmed.length > 0,
    placeholderData: (previousData) => previousData,
  });
}
