import { useMutation } from "@tanstack/react-query";

import { startGitHubConnect } from "@/lib/github-api";

export function useGitHubConnect() {
  return useMutation({
    mutationFn: startGitHubConnect,
  });
}