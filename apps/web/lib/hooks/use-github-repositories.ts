import { useQuery } from "@tanstack/react-query";

import { listGitHubRepositories } from "@/lib/github-api";

export function useGitHubRepositories() {
  return useQuery({
    queryKey: ["github-repositories"],
    queryFn: listGitHubRepositories,
    retry: false,
  });
}