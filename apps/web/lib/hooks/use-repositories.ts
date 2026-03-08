import { useQuery } from "@tanstack/react-query";

import { listRepositories } from "@/lib/repositories-api";

export function useRepositories() {
  return useQuery({
    queryKey: ["repositories"],
    queryFn: listRepositories,
  });
}

