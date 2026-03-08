import { useMutation } from "@tanstack/react-query";

import { triggerRepositorySync } from "@/lib/repositories-api";

export function useTriggerSync() {
  return useMutation({
    mutationFn: (repoId: string) => triggerRepositorySync(repoId),
  });
}

