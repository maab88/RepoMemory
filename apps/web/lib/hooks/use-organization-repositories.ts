import { useQuery } from "@tanstack/react-query";

import { listOrganizationRepositories } from "@/lib/repositories-api";

export function useOrganizationRepositories(orgId: string) {
  return useQuery({
    queryKey: ["organization-repositories", orgId],
    queryFn: () => listOrganizationRepositories(orgId),
    enabled: Boolean(orgId),
  });
}

