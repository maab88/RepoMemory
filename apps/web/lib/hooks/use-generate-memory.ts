import { useMutation } from "@tanstack/react-query";

import { generateRepositoryMemory } from "@/lib/repositories-api";

export function useGenerateMemory() {
  return useMutation({
    mutationFn: (repoId: string) => generateRepositoryMemory(repoId),
  });
}
