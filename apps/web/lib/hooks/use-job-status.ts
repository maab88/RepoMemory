import { useQuery } from "@tanstack/react-query";

import { getJob } from "@/lib/repositories-api";

export function useJobStatus(jobId: string | null) {
  return useQuery({
    queryKey: ["job-status", jobId],
    queryFn: () => getJob(jobId ?? ""),
    enabled: Boolean(jobId),
    refetchInterval: (query) => {
      const status = query.state.data?.job.status;
      if (status === "queued" || status === "running") {
        return 2000;
      }
      return false;
    },
  });
}

