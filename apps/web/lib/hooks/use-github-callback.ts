import { useQuery } from "@tanstack/react-query";

import { completeGitHubCallback } from "@/lib/github-api";

export function useGitHubCallback(code: string | undefined, state: string | undefined) {
  return useQuery({
    queryKey: ["github-callback", code, state],
    queryFn: () => completeGitHubCallback(code ?? "", state ?? ""),
    enabled: Boolean(code && state),
    retry: false,
  });
}