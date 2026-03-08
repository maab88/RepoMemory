import { useMutation } from "@tanstack/react-query";

import { generateRepositoryDigest } from "@/lib/repositories-api";

export function useGenerateDigest() {
  return useMutation({
    mutationFn: (repoId: string) => generateRepositoryDigest(repoId),
  });
}
