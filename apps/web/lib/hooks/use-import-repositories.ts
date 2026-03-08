import { useMutation } from "@tanstack/react-query";

import { importGitHubRepositories } from "@/lib/github-api";

export function useImportRepositories() {
  return useMutation({
    mutationFn: importGitHubRepositories,
  });
}